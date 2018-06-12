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
	idleState       int32 = 0
	dispatchedState int32 = 1
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
	In  []byte  // in/ingress or "read" buffer
	Out *[]byte // out/egress or "write" buffer

	backlog     *[]cmdGroup
	worker      *[]cmdGroup
	workerState int32
	next        cmdGroup

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
		Ev:      ev,
		Out:     &emptyBuffer,
		backlog: emptyList,
		worker:  emptyList,
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
	workerState := atomic.LoadInt32(&c.workerState)

	// Flush write atomically
	out = c.swapOut(&emptyBuffer)
	//if len(o) > 0 {
	//	out = append(out, o...)
	//}

	if action == evio.Close {
		return out, action
	}

	// Load the backlog
	backlog := c.loadBacklog()

	data := in
	if len(in) == 0 {
		// Increment loop wake counter
		atomic.AddUint64(&c.stats.wakes, 1)

		// Is there nothing to parse?
		if len(c.In) == 0 {
			//if workerState > 0 {
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
							backlog = append(backlog, c.next)
						}

						c.next = cmdGroup{}
						c.next.isMulti = true
						c.next.qidx = -1
						goto Parse
					}

				case "exec":
					if c.next.isMulti {
						if c.next.size() > 0 {

						}
						backlog = append(backlog, c.next)
						c.next = cmdGroup{}
						goto Parse

						//if workerState == 0 {
						//	out, workerState = c.emptyFromLoop(out)
						//} else {
						//	// Did we run out of space in the backlog?
						//	c.backlog.push(c.next)
						//
						//	goto Parse
						//}
					} else {
						c.next.list = append(c.next.list, api.Err("ERR exec not expected"))
						goto Parse
					}

				case "discard":
					if c.next.isMulti {
						c.next = cmdGroup{}
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
			backlog = append(backlog, c.next)
			c.next.clear()
		}
	}

	if workerState == dispatchedState {
		//fmt.Println(backlog)
	}

	if len(backlog) > 0 {
		if workerState == idleState {
			var (
				group cmdGroup
				index int
				ok    bool
			)

		loop:
			for index, group = range backlog {
				if group.isWorker {
					if group.isMulti {
						out, ok = c.sendQueued(out, &group)
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

					workerState = dispatchedState
					backlog = backlog[index:]
					break loop
				} else {
					if group.isMulti {
						out, ok = c.sendQueued(out, &group)
						if !ok {
							goto loop
						}
					}

					// Run all the commands
					out = c.execute(out, &group)
				}
			}

			if workerState == dispatchedState {
				c.swapWorker(&backlog)
				atomic.StoreInt32(&c.workerState, dispatchedState)
				//c.dispatch()
				Workers.Dispatch(c)
			}
		} else {
			// Append to the backlog
			c.storeBacklog(&backlog)
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

func (c *CmdConn) dispatch() {
	//if atomic.CompareAndSwapInt32(&c.workerState, 0, 1) {
	Workers.Dispatch(c)
	//}
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
	commands := c.swapWorker(_EmptyCommands)
	if commands == nil || len(commands) == 0 {
		c.wake()
		return
	}

	// atomic writes
	out := c.swapOut(&emptyBuffer)

loop:
// Since concurrent writes may happen we will cap the number of "pops" to
// the snapshot above. Only 1 goroutine "pops" at a time and only the event-loop "pushes".
// Which means the event-loop is in charge of parsing new commands and adding them to the backlog.
// The worker merely processes what it can and atomically flushes "write" buffer for use
// after the event-loop wakes this descriptor up.
	for _, item := range commands {
		if item.size() == 0 {
			continue loop
		}

		if item.isMulti {
			// Send Queued responses first
			if item.qidx < int32(item.size()) {
				c.sendQueued(out, &item)
			}
			// Run all the commands
			for _, command := range item.list {
				out = c.AppendCommand(out, command)
			}
		} else {
			// Run all the commands
			for _, command := range item.list {
				out = c.AppendCommand(out, command)
			}
		}

		item.clear()
	}

	if out == nil {
		out = emptyBuffer
	}

	// Atomically set write buffer
	c.swapOut(&out)
	//atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.Out)), unsafe.Pointer(&out))

	// Flip into non working mode
	atomic.StoreInt32(&c.workerState, idleState)

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

//func (c cmdGroup) add(command api.Command) {
//	c.list = append(c.list, command)
//}

//var emptyCmdGroup = cmdGroup{}
//
//const maxBacklog = 3
//
////const maxBacklogIdx = 1
//
//var errFilled = errors.New("backlog filled")
//
//// Lock free circular list
//// It supports 1 concurrent reader and 1 concurrent writer.
//type backlog struct {
//	head uint32
//	tail uint32
//	size int32
//	list []cmdGroup
//}
//
//func (b *backlog) isFilled() bool {
//	return atomic.LoadInt32(&b.size) == maxBacklog
//}
//
//func (b *backlog) push(group cmdGroup) error {
//	//if atomic.LoadInt32(&b.size) == maxBacklog {
//	//	return errFilled
//	//}
//
//	//b.list[b.tail%maxBacklog] = group
//	b.list = append(b.list, group)
//	atomic.AddUint32(&b.tail, 1)
//	// Increase size last
//	atomic.AddInt32(&b.size, 1)
//	return nil
//}
//
//func (b *backlog) pop() (*cmdGroup, bool) {
//	if atomic.LoadInt32(&b.size) == 0 {
//		return &emptyCmdGroup, false
//	}
//
//	//val := &b.list[b.head%maxBacklog]
//	val := b.list[0]
//	b.list = b.list[1:]
//	// Decrement size first
//	atomic.AddInt32(&b.size, -1)
//	// Increment head
//	//atomic.AddUint32(&b.head, 1)
//
//	return &val, true
//}
//
//func (b *backlog) peek() (*cmdGroup, bool) {
//	if atomic.LoadInt32(&b.size) == 0 {
//		return &emptyCmdGroup, false
//	}
//	return &b.list[0], true
//	//return &b.list[b.head%maxBacklog], false
//}
