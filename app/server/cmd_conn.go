package server

import (
	"fmt"
	"io"
	"errors"
	"runtime"
	"sync/atomic"
	"unsafe"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/redcon"
	"strings"
)

var maxCommandBacklog = 10000

func ErrWake(err error) error {
	return fmt.Errorf("wake: %s", err.Error())
}

var ErrBufferFilled = errors.New("buffer filled")
var maxRequestBuffer = 65536

var emptyBuffer []byte
var clearBuffer = unsafe.Pointer(&emptyBuffer)

type connStats struct {
	wakes          uint64
	commands       uint64
	commandsWorker uint64
	workerDur      int64
	ingress        uint64
	egress         uint64
}

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

	backlogMode int32
	backlog     backlog
	next        cmdGroup

	onDetached func(rwc io.ReadWriteCloser)
	onData     func(in []byte) (out []byte, action evio.Action)

	stats connStats
}

func NewConn(ev evio.Conn) *CmdConn {
	conn := &CmdConn{
		Ev:  ev,
		Out: &emptyBuffer,
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
	backlogMode := atomic.LoadInt32(&c.backlogMode)

	// Flush write atomically
	out = *(*[]byte)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.Out)), clearBuffer))
	if len(out) == 0 {
		out = nil
	}

	if action == evio.Close {
		return out, action
	}

	data := in
	if len(in) == 0 {
		// Wake checkpoint
		//c.wakeCheckpoint = c.wakeRequest

		// Increment loop wake counter
		atomic.AddUint64(&c.stats.wakes, 1)

		// Is there nothing to parse?
		if len(c.In) == 0 {
			if backlogMode > 0 {
				return out, action
			}

			// Empty backlog if possible
			out, backlogMode = c.emptyFromLoop(out)
			return out, action
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

	// Is there nothing in the request buffer?
	if len(data) == 0 {
		if backlogMode > 0 {
			return out, action
		}

		// Empty backlog if possible
		out, backlogMode = c.emptyFromLoop(out)
		return out, action
	}

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
			switch strings.ToLower(string(args[0])) {
			case "multi":
				if c.next.isMulti {
					c.next.list = append(c.next.list, api.Err("ERR multi cannot nest"))
					goto Parse
				} else {
					if backlogMode == 0 {
						out, backlogMode = c.emptyFromLoop(out)
					} else {
						c.backlog.push(c.next)
					}

					c.next = cmdGroup{}
					c.next.isMulti = true
					c.next.list = append(c.next.list, api.Ok{})
					goto Parse
				}

			case "exec":
				if c.next.isMulti {
					if backlogMode == 0 {
						out, backlogMode = c.emptyFromLoop(out)
					} else {
						// Did we run out of space in the backlog?
						c.backlog.push(c.next)
						c.next = cmdGroup{}
						goto Parse
					}
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

AfterParse:

// Try to process any commands if possible and/or dispatch a worker.
	if backlogMode == 0 {
		out, backlogMode = c.emptyFromLoop(out)
	} else {
		if c.next.size() > 0 && !c.next.isMulti {
			c.backlog.push(c.next)
			c.next = cmdGroup{}
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

func (c *CmdConn) getBacklogMode() int32 {
	return atomic.LoadInt32(&c.backlogMode)
}

func (c *CmdConn) dispatch() {
	if atomic.CompareAndSwapInt32(&c.backlogMode, 0, 1) {
		Workers.Dispatch(c)
	}
}

func (c *CmdConn) Run() {
	c.emptyFromWorker()
}

func (c *CmdConn) pushNext(out []byte) ([]byte, int32) {
	// Let's determine what to do with "next" group.
	item := c.next

	if item.size() == 0 {
		return out, 0
	}

	if item.isMulti {
		// Continue queuing but do not push next onto the backlog until
		// we receive an EXEC or DISCARD command
		//var ok bool
		out, _ = c.ensureQueued(out, &item)
		return out, 0
	} else {
		if item.isWorker {
			// Process until the first worker command is foun.
			// This optimizes are time with the event loop by processing
			// as many commands as possible before depending on the Worker.
			// We will then have a write to flush which cuts the latency
			// down significantly.
			var (
				index   int
				command api.Command
			)
		loop:
			for index, command = range item.list {
				if command.IsWorker() {
					if index > 0 {
						// slice it down
						item.list = item.list[index:]
					}
					break loop
				} else {
					out = c.AppendCommand(out, command)
				}
			}
			if index > 0 {
				item.list = item.list[index:]
			}
			if item.size() > 0 {
				c.backlog.push(item)
			}
			// Clear next
			c.next = cmdGroup{}

			// moving into worker mode
			c.dispatch()

			return out, 1
		} else {
			out = c.execute(out, &item)
			c.next = cmdGroup{}
		}
	}
	return out, 0
}

// This must ONLY be called when no worker is currently in-progress.
func (c *CmdConn) emptyFromLoop(out []byte) ([]byte, int32) {
	var (
		item *cmdGroup
		ok   = false
	)

Next:
	item, ok = c.backlog.peek()
	if !ok {
		// Let's determine what to do with "next" group.
		return c.pushNext(out)
	} else {
		if !item.isWorker {
			// pop it since we can execute immediately
			c.backlog.pop()

			if item.isMulti {
				out, ok = c.ensureQueued(out, item)
				if !ok {
					goto Next
				}
			}

			// Run all the commands
			out = c.execute(out, item)

			goto Next
		} else {
			if item.isMulti {
				out, ok = c.ensureQueued(out, item)
				if !ok {
					goto Next
				}
			} else {
				// Process until the first worker command is foun.
				// This optimizes are time with the event loop by processing
				// as many commands as possible before depending on the Worker.
				// We will then have a write to flush which cuts the latency
				// down significantly.
				for index, command := range item.list {
					if command.IsWorker() {
						if index > 0 {
							// slice it down
							item.list = item.list[index:]
						}
						return out, 0
					} else {
						out = c.AppendCommand(out, command)
					}
				}
			}

			// moving into worker mode
			c.dispatch()

			// Exit
			return out, 1
		}
	}

	return out, 0
}

func (c *CmdConn) ensureQueued(out []byte, item *cmdGroup) ([]byte, bool) {
	// Send Queued responses
	if item.qidx < item.size() {
		if item.qidx == 0 {
			out = c.AppendCommand(out, item.list[0])
			item.qidx = 1
		}
		for i := item.qidx; i < item.size(); i++ {
			command := item.list[i]
			if command.IsError() {
				out = c.AppendCommand(out, command)
				item.isMulti = false
				item.isWorker = false
				item.qidx = 0
				item.list = item.list[:]
				return out, false
			}
			out = redcon.AppendQueued(out)
		}
		item.qidx = item.size()
	}
	return out, true
}

func (c *CmdConn) execute(out []byte, item *cmdGroup) ([]byte) {
	// Run all the commands
	for _, command := range item.list {
		out = c.AppendCommand(out, command)
	}
	//*item = cmdGroup{}
	//item.list = item.list[:0]
	//item.list = nil
	//item.isWorker = false
	//item.isMulti = false
	//item.qidx = 0
	return out
}

func (c *CmdConn) emptyFromWorker() {
	// Snapshot current backlog size
	size := atomic.LoadInt32(&c.backlog.size)
	if size == 0 {
		return
	}

	//
	count := uint32(size)
	head := atomic.LoadUint32(&c.backlog.head)

	// atomic writes
	out := *(*[]byte)(atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&c.Out)),
		clearBuffer),
	)
	if len(out) == 0 {
		out = nil
	}

loop:
// Since concurrent writes may happen we will cap the number of "pops" to
// the snapshot above. Only 1 goroutine "pops" at a time and only the event-loop "pushes".
// Which means the event-loop is in charge of parsing new commands and adding them to the backlog.
// The worker merely processes what it can and atomically flushes "write" buffer for use
// after the event-loop wakes this descriptor up.
	for index := head; index < head+count; index++ {
		item, ok := c.backlog.pop()
		if !ok {
			break loop
		}

		if item.size() == 0 {
			continue loop
		}

		if item.isMulti {
			// Send Queued responses first
			if item.qidx < int32(item.size()) {
				for i := item.qidx; i < item.size(); i++ {
					command := item.list[i]
					if command.IsError() {
						out = c.AppendCommand(out, command)
						continue loop
					}
				}
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
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.Out)), unsafe.Pointer(&out))

	// Flip into non working mode
	atomic.StoreInt32(&c.backlogMode, 0)

	// Wake up the loop
	c.Ev.Wake()
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
	c.list = nil
}

func (c *cmdGroup) size() int32 { return int32(len(c.list)) }

//func (c cmdGroup) add(command api.Command) {
//	c.list = append(c.list, command)
//}

var emptyCmdGroup = cmdGroup{}

const maxBacklog = 3

//const maxBacklogIdx = 1

var errFilled = errors.New("backlog filled")

// Lock free circular list
// It supports 1 concurrent reader and 1 concurrent writer.
type backlog struct {
	head uint32
	tail uint32
	size int32
	list []cmdGroup


}

func (b *backlog) isFilled() bool {
	return atomic.LoadInt32(&b.size) == maxBacklog
}

func (b *backlog) push(group cmdGroup) error {
	//if atomic.LoadInt32(&b.size) == maxBacklog {
	//	return errFilled
	//}

	//b.list[b.tail%maxBacklog] = group
	b.list = append(b.list, group)
	atomic.AddUint32(&b.tail, 1)
	// Increase size last
	atomic.AddInt32(&b.size, 1)
	return nil
}

func (b *backlog) pop() (*cmdGroup, bool) {
	if atomic.LoadInt32(&b.size) == 0 {
		return &emptyCmdGroup, false
	}

	//val := &b.list[b.head%maxBacklog]
	val := b.list[0]
	b.list = b.list[1:]
	// Decrement size first
	atomic.AddInt32(&b.size, -1)
	// Increment head
	//atomic.AddUint32(&b.head, 1)

	return &val, true
}

func (b *backlog) peek() (*cmdGroup, bool) {
	if atomic.LoadInt32(&b.size) == 0 {
		return &emptyCmdGroup, false
	}
	return &b.list[0], true
	//return &b.list[b.head%maxBacklog], false
}

//func (c *Conn) push(group cmdGroup) error {
//	if atomic.LoadInt32(&c.backlog.size) == maxBacklog {
//		return errFilled
//	}
//
//	c.backlog.list[c.backlog.tail%maxBacklog] = group
//	atomic.AddUint32(&c.backlog.tail, 1)
//	// Increase size last
//	atomic.AddInt32(&c.backlog.size, 1)
//	return nil
//}
//
//func (c *Conn) pop() (*cmdGroup, bool) {
//	if atomic.LoadInt32(&c.backlog.size) == 0 {
//		return &emptyCmdGroup, false
//	}
//
//	val := &c.backlog.list[c.backlog.head%maxBacklog]
//	// Decrement size first
//	atomic.AddInt32(&c.backlog.size, -1)
//	// Increment head
//	atomic.AddUint32(&c.backlog.head, 1)
//
//	return val, true
//}
//
//func (c *Conn) peek() (*cmdGroup, bool) {
//	size := atomic.LoadInt32(&c.backlog.size)
//	if size == 0 {
//		return nil, false
//	}
//	return &c.backlog.list[c.backlog.head%maxBacklog], true
//}
