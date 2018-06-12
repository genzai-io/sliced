package server

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync/atomic"
	"unsafe"

	"strings"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/redcon"
)

var maxCommandBacklog = 10000

func ErrWake(err error) error {
	return fmt.Errorf("wake: %s", err.Error())
}

var ErrBufferFilled = errors.New("buffer filled")
var maxRequestBuffer = 65536

var BeginBufferSize = 64

var emptyBuffer []byte
var clearBuffer = unsafe.Pointer(&emptyBuffer)

var workerBuffer []byte

type connStats struct {
	wakes          uint64
	commands       uint64
	commandsWorker uint64
	workerDur      int64
	ingress        uint64
	egress         uint64
}

var emptyList = &[]cmdGroup{}

const (
	loopOwner   int32 = 0
	workerOwner int32 = 1
)

// Non-Blocking
//
// This type adheres to the RESP protocol where Every command must happen in-order
// and must have a single RESP reply with the exception of MULTI groups which require
// up to 2 replies per command:
// 1. Queued OK
// 2. Reply
//
// Great care goes into the lowest latency responses while guaranteeing there
// no be no blocking when on the event-loop.
//
// Worker (background) commands are supported and will be processed in the background which
// will wake the event-loop up when there is data to write. More commands may
// queue up concurrently while the worker is in progress. However, only 1 worker
// may work at a time and once it buffers data to write then it transfers ownership
// back to the event-loop. A custom Worker pool was created to handle Worker processing.
// A worker is opportunistic and will drain as much of the backlog as possible to
// remove as much CPU cycles as possible from the event-loop.
//
//
// A custom non-blocking circular list data structure is used for the command backlog
// of command groupings. It allows the event-loop to "push" new command groups in
// while a worker "pops" them off concurrently without blocking by making novel use
// of atomics.
//
// Transactions are supported via the MULTI, EXEC and DISCARD commands and follow the
// same behavior to the Redis implementation. All commands between a MULTI and EXEC
// command will happen all at once. If one of those commands is a worker then ALL will
// be processed at once in the background.
//
//
type CmdConn struct {
	api.Context

	mutex       uintptr
	mutexMisses uint64

	Ev   evio.Conn // Connection
	done bool      // flag to signal it's done

	// Buffers
	In  []byte // in/ingress or "read" buffer
	Out []byte // out/egress or "write" buffer

	backlog     []*cmdGroup
	worker      []*cmdGroup
	workerState int32
	next        *cmdGroup

	// For "multi" transactions this is registry of vars of named results.
	// $x = GET key
	// if $x == 0 SET key $x.incr()
	register map[string]api.CommandReply

	onDetached func(rwc io.ReadWriteCloser)
	onData     func(in []byte) (out []byte, action evio.Action)

	stats connStats
}

func NewConn(ev evio.Conn) *CmdConn {
	conn := &CmdConn{
		Ev: ev,
		//Out:     &emptyBuffer,
		//backlog: emptyList,
		//worker:  emptyList,
	}
	return conn
}

// Spin-lock
// Only the properties are synchronized and not the command Handle() itself.
// In addition, the Event Loop is inherently single-threaded so the only
// potential race is from a background Worker happening in parallel with
// an Event Loop call.
func (c *CmdConn) Lock() {
	for !atomic.CompareAndSwapUintptr(&c.mutex, 0, 1) {
		atomic.AddUint64(&c.mutexMisses, 1)
		runtime.Gosched()
	}
}

// Spin-lock TryLock
func (c *CmdConn) TryLock() bool {
	if !atomic.CompareAndSwapUintptr(&c.mutex, 0, 1) {
		atomic.AddUint64(&c.mutexMisses, 1)
		return false
	} else {
		return true
	}
}

// Spin-lock Unlock
func (c *CmdConn) Unlock() {
	atomic.StoreUintptr(&c.mutex, 0)
}

func (c *CmdConn) Detach() error {
	c.Lock()
	c.Action = evio.Detach
	c.Unlock()
	c.Ev.Wake()
	return nil
}

func (c *CmdConn) OnDetach(rwc io.ReadWriteCloser) {
	if rwc != nil {
		rwc.Close()
	}
}

func (c *CmdConn) Close() error {
	c.Lock()
	c.Action = evio.Close
	conn := c.Ev
	c.Unlock()

	if conn != nil {
		return conn.Wake()
	}
	return nil
}

func (c *CmdConn) OnClosed() {
	c.Lock()
	c.done = true
	c.Action = evio.Close
	c.Ev = nil
	c.Unlock()
}

func (c *CmdConn) Conn() evio.Conn {
	return c.Ev
}

func (c *CmdConn) Stats() {
	c.Lock()
	c.Unlock()
}

// This is not thread safe
func (c *CmdConn) OnData(in []byte) ([]byte, evio.Action) {
	var out []byte
	var action = c.Action

	// Snapshot current working mode
	owner := atomic.LoadInt32(&c.workerState)

	// Flush write atomically
	//out = c.swapOut(&emptyBuffer)
	//c.Lock()
	//out = c.Out
	//out = nil
	//c.Unlock()

	// Flush worker writes if necessary
	out = c.Out
	c.Out = nil

	if c.next == nil {
		c.next = &cmdGroup{}
	}

	if action == evio.Close {
		return out, action
	}

	// Load the backlog
	//backlog := c.loadBacklog()

	data := in
	if len(in) == 0 {
		// Increment loop wake counter
		atomic.AddUint64(&c.stats.wakes, 1)

		// Is there nothing to parse?
		if len(c.In) == 0 {
			//if owner > 0 {
			//	return out, action
			//}

			// Empty backlog if possible
			goto AfterParse
		} else {
			// Were there any leftovers from the previous event?
			data = c.In
			c.In = nil
		}
	} else {
		// Ingress
		atomic.AddUint64(&c.stats.ingress, uint64(len(in)))

		// Were there any leftovers from the previous event?
		if len(c.In) > 0 {
			data = append(c.In, in...)
			c.In = nil
		}
	}

	// Is there any data to parse?
	if len(data) > 0 {
		var
		(
			packet   []byte
			complete bool
			args     [][]byte
			err      error
			command  api.Command
		)

	Parse:
	// Let's parse the commands
		for {
			// Read next command.
			packet, complete, args, _, data, err = redcon.ParseNextCommand(data, args[:0])

			if err != nil {
				c.Lock()
				c.Action = evio.Close
				c.Reason = err
				out = redcon.AppendError(out, err.Error())
				c.Unlock()
				return out, evio.Close
			}

			// Do we need more data?
			if !complete {
				// Exit loop.
				goto AfterParse
			}

			switch len(args) {
			case 0:
				goto AfterParse

			case 1:
				name := *(*string)(unsafe.Pointer(&args[0]))

				switch strings.ToLower(name) {
				case "multi":
					if c.next.isMulti {
						c.next.list = append(c.next.list, api.Err("ERR multi cannot nest"))
						goto Parse
					} else {
						if c.next.size() > 0 {
							c.backlog = append(c.backlog, c.next)
							c.next = &cmdGroup{}
						}

						c.next.isMulti = true
						c.next.qidx = -1
						goto Parse
					}

				case "exec":
					if c.next.isMulti {
						c.backlog = append(c.backlog, c.next)
						c.next = &cmdGroup{}
						goto Parse
					} else {
						c.next.list = append(c.next.list, api.Err("ERR exec not expected"))
						goto Parse
					}

				case "discard":
					if c.next.isMulti {
						c.next = &cmdGroup{}
						c.next.list = append(c.next.list, api.Ok{})
						goto Parse
					} else {
						c.next.list = append(c.next.list, api.Err("ERR discard not expected"))
						goto Parse
					}
				}

			default:
				// Do we have an expression?
				if len(args[1]) > 0 && args[1][0] == '=' {

				}
			}

			if command == nil {
				if c.Parse == nil {
					command = api.ParseCommand(packet, args)
				} else {
					command = c.Parse(packet, args)
				}
			}
			if command == nil {
				command = api.Err(fmt.Sprintf("ERR command '%s' not found", args[0]))
			}

			c.next.isWorker = command.IsWorker()
			c.next.list = append(c.next.list, command)
		}
	}

AfterParse:

// Should push next?
	if c.next.size() > 0 {
		if !c.next.isMulti {
			// Optimize for common scenarios.
			// Let's try to save a slice append.
			// Benchmarking revealed around 8-10% throughput increase under heavy load,
			// so that's pretty nifty.
			if !c.next.isWorker && len(c.backlog) == 0 {
				out = c.execute(out, c.next)
				c.next.clear()
			} else {
				c.backlog = append(c.backlog, c.next)
				c.next = &cmdGroup{}
			}
		}
	}

	if owner == loopOwner {
		if len(c.backlog) > 0 {
			var (
				group *cmdGroup
				index int
				ok    bool
			)

		loop:
			for index, group = range c.backlog {
				if group.isWorker {
					if group.isMulti {
						out, ok = c.sendQueued(out, group)
						if !ok {
							goto loop
						}
					} else {
					bl:
					// Process until the first worker command is foun.
					// This optimizes are time with the event loop by processing
					// as many commands as possible before depending on the Worker.
					// We will then have a write to flush which cuts the latency
					// down significantly.
						for index, command := range group.list {
							if command.IsWorker() {
								if index > 0 {
									// slice it down
									group.list = group.list[index:]
								}
								break bl
							} else {
								out = c.AppendCommand(out, command)
							}
						}
					}

					owner = workerOwner
					if index > 0 {
						c.backlog = c.backlog[index:]
					}
					break loop
				} else {
					if group.isMulti {
						out, ok = c.sendQueued(out, group)
						if !ok {
							goto loop
						}
					}

					// Run all the commands
					out = c.execute(out, group)
				}
			}

			if owner == workerOwner {
				// transfer backlog to worker
				c.worker = append(c.worker, c.backlog...)
				// Clear the backlog
				c.backlog = c.backlog[:0]
				// Move to dispatched owner
				atomic.StoreInt32(&c.workerState, workerOwner)
				Workers.Dispatch(c)
			} else {
				// Clear the backlog
				c.backlog = c.backlog[:0]

				if c.next.isMulti {
					out, _ = c.sendQueued(out, c.next)
				}
			}
		} else {
			if c.next.isMulti {
				out, _ = c.sendQueued(out, c.next)
			}
		}
	}

	// Are there any leftovers (partial commands)?
	// This method has exclusive access to the "In" buffer
	// so no need to do this within the mutex.
	// If the backlog is filled then we will defer command parsing until a later time.
	if len(data) > 0 {
		c.In = append(c.In, data...)
	}

	// Egress stats
	atomic.AddUint64(&c.stats.egress, uint64(len(out)))

	// Return
	return out, action
}

func (c *CmdConn) sendQueued(out []byte, group *cmdGroup) ([]byte, bool) {
	// Send +OK for the "multi" command
	if group.qidx == -1 {
		out = redcon.AppendOK(out)
		group.qidx = 0
	}

	if group.size() == 0 {
		return out, true
	}

	// Followed by +QUEUED for all the other commands in the group
	for i := group.qidx; i < group.size(); i++ {
		command := group.list[i]
		// Errors will cancel the whole group
		if command.IsError() {
			// Append the error
			out = c.AppendCommand(out, command)

			// Reset the group
			group.clear()

			// Exit as error
			return out, false
		}
		out = redcon.AppendQueued(out)
	}
	group.qidx = group.size()

	return out, true
}

func (c *CmdConn) execute(out []byte, group *cmdGroup) ([]byte) {
	if group.isMulti {
		var ok bool
		out, ok = c.sendQueued(out, group)
		if !ok {
			return out
		}

		// let's out as a single Array
		out = redcon.AppendArray(out, int(group.size()))

		// Run all the commands
		for _, command := range group.list {
			out = c.AppendCommand(out, command)
		}
	} else {
		// Run all the commands
		for _, command := range group.list {
			out = c.AppendCommand(out, command)
		}
	}

	return out
}

func (c *CmdConn) wake() {
	if err := c.Ev.Wake(); err != nil {
		c.Reason = err
		c.Action = evio.Close
	}
}

// Worker run
func (c *CmdConn) Run() {
	// Atomically get and clear the work list
	//groups := c.swapWorker(_EmptyCommands)
	//if groups == nil || len(groups) == 0 {
	//	c.wake()
	//	return
	//}

	// atomic writes
	//out := c.swapOut(&emptyBuffer)

	out := make([]byte, 0, BeginBufferSize)
	groups := c.worker
	c.worker = c.worker[:0]

loop:
// Since concurrent writes may happen we will cap the number of "pops" to
// the snapshot above. Only 1 goroutine "pops" at a time and only the event-loop "pushes".
// Which means the event-loop is in charge of parsing new groups and adding them to the backlog.
// The worker merely processes what it can and atomically flushes "write" buffer for use
// after the event-loop wakes this descriptor up.
	for _, group := range groups {
		if group.size() == 0 {
			continue loop
		}

		out = c.execute(out, group)
		group.clear()
	}

	if out == nil {
		//out = emptyBuffer
	}

	// Atomically set write buffer
	//c.swapOut(&out)

	c.Out = out

	// Flip into non working mode
	atomic.StoreInt32(&c.workerState, loopOwner)

	// Wake up the loop
	c.wake()
}

var _EmptyCommands = &[]cmdGroup{}
var _EmptyCommandsPtr = unsafe.Pointer(&_EmptyCommands)

func (c *CmdConn) swapWorker(to *[]cmdGroup) []cmdGroup {
	ptr := atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&c.worker)),
		unsafe.Pointer(&*to),
	)

	if ptr == nil || ptr == _EmptyCommandsPtr {
		return nil
	} else {
		g := (*[]cmdGroup)(ptr)
		return *g
	}
}

func (c *CmdConn) loadBacklog() []cmdGroup {
	ptr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.backlog)))
	if ptr == nil || ptr == _EmptyCommandsPtr {
		return nil
	} else {
		c := (*[]cmdGroup)(ptr)
		return *c
	}
}

func (c *CmdConn) storeBacklog(to *[]cmdGroup) {
	atomic.StorePointer(
		(*unsafe.Pointer)(unsafe.Pointer(&c.backlog)),
		unsafe.Pointer(&*to),
	)
}

func (c *CmdConn) swapBacklog(to *[]cmdGroup) []cmdGroup {
	ptr := atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&c.backlog)),
		unsafe.Pointer(&*to),
	)

	if ptr == nil || ptr == _EmptyCommandsPtr {
		return nil
	} else {
		g := (*[]cmdGroup)(ptr)
		return *g
	}
}

func (c *CmdConn) swapOut(to *[]byte) []byte {
	ptr := atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&c.Out)),
		unsafe.Pointer(&*to),
	)

	if ptr == nil || ptr == clearBuffer {
		return nil
	} else {
		// De-reference
		g := (*[]byte)(ptr)
		return *g
	}
}

type cmdGroup struct {
	isMulti  bool
	isWorker bool
	qidx     int32
	list     []api.Command

	//left *cmdGroup
}

func (c *cmdGroup) clear() {
	c.isMulti = false
	c.isWorker = false
	c.qidx = 0
	c.list = c.list[:0]
}

func (c *cmdGroup) size() int32 { return int32(len(c.list)) }
