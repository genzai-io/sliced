package server

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"strings"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/fastlane"
	"github.com/genzai-io/sliced/common/resp"
)

var (
	ErrBufferFilled = errors.New("buffer filled")
	ErrWake         = func(err error) error {
		return fmt.Errorf("wake error: %s", err.Error())
	}

	maxCommandBacklog = 10000
	emptyBuffer       []byte
	stopMsg           = &cmdGroup{}
)

type connStats struct {
	wakes          uint64
	commands       uint64
	commandsWorker uint64
	workerDur      int64
	ingress        uint64
	egress         uint64
}

const (
	loopOwner   int32 = 0
	workerOwner int32 = 1
)

// Non-Blocking RESP Connection
//
// RESP protocol event-loop connection where every command must happen in-order
// and must have a single RESP reply. There are two exceptions to that rule.
//
// 1. "MULTI" transactions where every command may have a
//			+QUEUED followed by a Reply after "EXEC"
//
// 2. PUB/SUB connections which supports push replies from server to client.
//
// This is optimized for low-latency and tries to flush output as often as possible.
// There are two distict ownership states for the command executor.
//
// 1. Event-Loop
//	  This is ran from the event-loop thread and cannot block under any circumstance.
//	  If a command publishes itself as a "Worker" then ownership must transfer to.
//
// 2. Worker
//    The processing happens on a private goroutine that may execute any kind of command.
//    A special non-blocking channel implementation was used to communicate between
//    the event-loop and the worker goroutine.
//
// Since commands must be executed to its completion before the next command is executed,
// this can result in a backlog of commands if a command requires "Worker" execution.
// Furthermore, all commands in a transaction must queue until they can all be executed
// at once. Again, if any of those commands require "worker" execution then it will backlog
// all commands in that group as well as subsequent commands.
type CmdConn struct {
	api.Context

	ID uint64

	ev   evio.Conn // evio Connection
	done bool      // flag to signal it's done

	// Partial commands get saved here while waiting for more data
	leftovers []byte

	// Determines where processing may take place
	ownership int32
	backlog   []*cmdGroup
	next      *cmdGroup
	worker    cmdConnWorker

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
		ev: ev,
	}
	conn.onData = conn.OnData
	return conn
}

func (c *CmdConn) tick() {
	// Determine if there is a weird state that needs to be fixed
	if c.Reason != nil && !c.done {
		c.wake()
	}
}

func (c *CmdConn) Detach() error {
	c.Action = evio.Detach
	c.ev.Wake()
	return nil
}

func (c *CmdConn) OnDetach(rwc io.ReadWriteCloser) {
	if rwc != nil {
		rwc.Close()
	}
}

func (c *CmdConn) Close() error {
	c.Action = evio.Close
	conn := c.ev

	if conn != nil {
		return conn.Wake()
	}
	return nil
}

func (c *CmdConn) OnClosed() {
	c.done = true
	c.Action = evio.Close
	c.ev = nil
	c.stopWorker()
}

func (c *CmdConn) Conn() evio.Conn {
	return c.ev
}

func (c *CmdConn) Stats() {
}

// This is not thread safe
func (c *CmdConn) OnData(in []byte) ([]byte, evio.Action) {
	var (
		out    []byte
		input  []byte
		action = c.Action
	)

	// Snapshot current working mode
	ownership := atomic.LoadInt32(&c.ownership)

	if len(in) == 0 {
		// Flush leftovers
		in = c.leftovers
		c.leftovers = nil
	} else {
		// Ingress
		atomic.AddUint64(&c.stats.ingress, uint64(len(in)))

		// Flush leftovers
		if len(c.leftovers) > 0 {
			input = append(c.leftovers, in...)
			c.leftovers = nil
		} else {
			input = in
		}
	}

	wakes := atomic.LoadUint64(&c.worker.wakes)
	if atomic.LoadUint64(&c.worker.wakeSnapshot) < wakes {
		// Increment loop wakes counter
		atomic.AddUint64(&c.stats.wakes, 1)
		// Snapshot wakes state
		atomic.StoreUint64(&c.worker.wakeSnapshot, wakes)

		// Flush any existing writes
		// Snapshot outCount
		if outCount := atomic.LoadInt32(&c.worker.outCount); outCount > 0 {
			for i := int32(0); i < outCount; i++ {
				b := c.worker.outCh.Recv()
				if len(b) > 0 {
					out = append(out, b...)
				}
			}
			// Mark them as processed
			atomic.AddInt32(&c.worker.outCount, -outCount)
		}
	}

	if c.next == nil {
		c.next = &cmdGroup{}
	}

	if action == evio.Close {
		return out, action
	}

	// Is there any input to parse?
	if len(input) > 0 {
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
			packet, complete, args, _, input, err = resp.ParseNextCommand(input, args[:0])

			if err != nil {
				c.Action = evio.Close
				c.Reason = err
				out = resp.AppendError(out, err.Error())
				return out, evio.Close
			}

			// Do we need more input?
			if !complete {
				// Exit loop.
				goto AfterParse
			}

			switch len(args) {
			case 0:
				goto AfterParse

			case 1:
				switch strings.ToLower(*(*string)(unsafe.Pointer(&args[0]))) {
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

	if ownership == loopOwner {
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

					ownership = workerOwner
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

			if ownership == workerOwner {
				// Move to dispatched ownership
				atomic.StoreInt32(&c.ownership, workerOwner)

				// Transfer backlog to worker
				for _, group := range c.backlog {
					c.sendToWorker(group)
				}

				// Clear the backlog
				c.backlog = c.backlog[:0]
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
	if len(input) > 0 {
		c.leftovers = append(c.leftovers, input...)
	}

	// Egress stats
	atomic.AddUint64(&c.stats.egress, uint64(len(out)))

	// Return
	return out, action
}

func (c *CmdConn) sendQueued(out []byte, group *cmdGroup) ([]byte, bool) {
	// Send +OK for the "multi" command
	if group.qidx == -1 {
		out = resp.AppendOK(out)
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
		out = resp.AppendQueued(out)
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
		out = resp.AppendArray(out, int(group.size()))

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

//
func (c *CmdConn) wake() {
	ev := c.ev
	if ev != nil {
		if err := c.ev.Wake(); err != nil {
			// This is a fatal error and this connection must be cleaned up.
			c.Reason = err
			c.Action = evio.Close
		}
	}
}

type cmdGroup struct {
	isMulti  bool
	isWorker bool
	qidx     int32
	list     []api.Command
}

func (c *cmdGroup) clear() {
	c.isMulti = false
	c.isWorker = false
	c.qidx = -1
	// Reset already allocated slice
	//c.list = c.list[:0]
	c.list = nil
}

// Size of the list
func (c *cmdGroup) size() int32 { return int32(len(c.list)) }

//
type cmdConnWorker struct {
	wg      sync.WaitGroup
	mutex   uintptr
	counter int32
	ch      cmdConnWorkerChan

	wakeSnapshot uint64
	wakes        uint64

	open         bool
	waitingSince int64

	outCh    bufferChan
	outCount int32
}

// This will close the background goroutine
func (c *CmdConn) stopWorker() {
	if c.worker.open {
		c.sendToWorker(stopMsg)
		c.worker.open = false
	}
}

// Called when the work queue is finished and ownership
// is transferred back to the event-loop.
// Since the ownership is not the worker anymore, this method
// is not safe to modify the working state.
func (c *CmdConn) workerCaughtUp() {
}

// Must be called from event-loop
func (c *CmdConn) sendToWorker(group *cmdGroup) {
	atomic.AddInt32(&c.worker.counter, 1)

	if !c.worker.open {
		c.worker.open = true
		// Ensure there is only ever a single goroutine running in the background
		c.worker.wg.Wait()
		c.workerLoop()
	}

	c.worker.ch.Send(group)
}

func (c *CmdConn) workerLoop() {
	c.worker.wg.Add(1)
	go func() {
		defer c.worker.wg.Done()
		var msg *cmdGroup
		var count int32

		for {
			// Wait to receive next msg
			c.worker.waitingSince = time.Now().UnixNano()
			msg = c.worker.ch.Recv()
			if msg == stopMsg || msg == nil {
				c.worker.waitingSince = 0
				count = atomic.AddInt32(&c.worker.counter, -1)
				return
			}

			// Process the group
			var b []byte
			b = c.execute(b, msg)
			msg.clear()

			// Decrement count
			count = atomic.AddInt32(&c.worker.counter, -1)

			// Flush write
			atomic.AddInt32(&c.worker.outCount, 1)
			c.worker.outCh.Send(&b)

			// Increment wake count
			wakes := atomic.AddUint64(&c.worker.wakes, 1)

			// Did we catch up?
			if count == 0 {
				// Transfer ownership back to the event-loop
				atomic.StoreInt32(&c.ownership, loopOwner)
				c.workerCaughtUp()
			}

			// Determine if the event-loop is behind.
			// Let's try to save a syscall to "epoll / kqueue / iocp"
			if atomic.LoadUint64(&c.worker.wakeSnapshot) == wakes-1 {
				// Wake the event loop
				c.wake()
			}
		}
	}()
}

// Channel of *cmdGroup
type cmdConnWorkerChan struct {
	base fastlane.ChanPointer
}

func (ch *cmdConnWorkerChan) Send(value *cmdGroup) {
	ch.base.Send(unsafe.Pointer(value))
}

func (ch *cmdConnWorkerChan) Recv() *cmdGroup {
	return (*cmdGroup)(ch.base.Recv())
}

// Channel of []byte
type bufferChan struct {
	base fastlane.ChanPointer
}

func (ch *bufferChan) Send(value *[]byte) {
	// Handle nil
	if value == nil {
		value = &emptyBuffer
	}
	ch.base.Send(unsafe.Pointer(value))
}

func (ch *bufferChan) Recv() []byte {
	// Dereference to []byte
	return *(*[]byte)(ch.base.Recv())
}

