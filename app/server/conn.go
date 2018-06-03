package server

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/app/command"
	"github.com/slice-d/genzai/common/evio"
)

var maxCommandBacklog = 10000

type Conn struct {
	evc    evio.Conn
	kind   api.ConnKind
	action evio.Action // event loop status
	done   bool        // flag to signal it's done

	mu   sync.Mutex
	lock uintptr
	in   []byte

	// Backlog
	backlog []command.Command
	active  command.Command

	// Handler to use to process and commit commands
	handler api.IHandler
	// Durability setting
	durability api.Durability

	// Context it's currently associated with
	ctx *api.Context

	// Current raft context
	// This is used for the RaftTransport to support multiple Raft clusters
	// over the same port
	raft api.RaftService

	// Stats
	statsTotalCommands   uint64
	statsTotalUpstream   uint64
	statsTotalDownstream uint64

	statsAsyncCommands uint64
	statsAsyncOps      uint64
	statsAsyncDur      time.Duration
}

// Spinlock Lock
func (l *Conn) Lock() {
	for !atomic.CompareAndSwapUintptr(&l.lock, 0, 1) {
		runtime.Gosched()
	}
}

// Spinlock Unlock
func (l *Conn) Unlock() {
	atomic.StoreUintptr(&l.lock, 0)
}

func (c *Conn) Kind() api.ConnKind {
	c.mu.Lock()
	k := c.kind
	c.mu.Unlock()
	return k
}

func (c *Conn) SetKind(kind api.ConnKind) {
	c.mu.Lock()
	c.kind = kind
	c.mu.Unlock()
}

func (c *Conn) Invoke(ctx *api.Context, cmd api.Command) {
	c.mu.Lock()
	if c.action != evio.None {
		c.mu.Unlock()
		ctx.Err("closed")
		return
	}

	if c.active != nil {
		if len(c.backlog) > maxCommandBacklog {
			c.mu.Unlock()
			ctx.Err("backlog filled")
			return
		}

		c.backlog = append(c.backlog, cmd)
	} else {
		if cmd.IsAsync() {
			atomic.AddUint64(&c.statsAsyncCommands, 1)
			// Set background context
			c.ctx = &api.Context{}
			c.dispatch(cmd)
		} else {
			before := len(ctx.Out)
			cmd.Handle(ctx)

			if len(ctx.Out) == before {
				c.dispatch(cmd)
			}
		}
	}
	c.mu.Unlock()
}

func (c *Conn) Raft() api.RaftService {
	//c.mu.Lock()
	//r := c.raft
	//c.mu.Unlock()
	return c.raft
}

func (c *Conn) SetRaft(raft api.RaftService) {
	c.mu.Lock()
	c.raft = raft
	c.mu.Unlock()
}

func (c *Conn) Close() error {
	c.mu.Lock()
	c.action = evio.Close
	c.mu.Unlock()
	return nil
}

func (c *Conn) Durability() api.Durability {
	return c.durability
}

func (c *Conn) Handler() api.IHandler {
	return c.handler
}

func (c *Conn) SetHandler(handler api.IHandler) api.IHandler {
	prev := c.handler
	c.handler = handler
	return prev
}

func (c *Conn) dispatch(cmd command.Command) {
	c.active = cmd
	Workers.Dispatch(c)
}

func (c *Conn) Run() {
	// Was the job already canceled?
	c.mu.Lock()
	if c.active == nil {
		c.mu.Unlock()
		return
	}

	if c.ctx == nil {
		c.ctx = &api.Context{}
	}

	l := len(c.ctx.Out)
	// Run job.
	c.active.Handle(c.ctx)

	if len(c.ctx.Out) == l {
		c.ctx.Err("not implemented")
	}

	c.active = nil
	c.mu.Unlock()

	// Notify event loop of our write.
	c.wake()
}

// Inform the event loop to close this connection.
func (c *Conn) close() {
	c.action = evio.Close
	c.wake()
}

// Called when the event loop has closed this connection.
func (c *Conn) closed() {
	c.done = true
}

// Asks the event loop to schedule a write event for this connection.
func (c *Conn) wake() {
	c.evc.Wake()
}

// Invoked on the event loop thread.
func (c *Conn) woke() *api.Context {
	// Set output buffer
	c.mu.Lock()
	if c.active != nil {
		c.mu.Unlock()
		return nil
	}

	ctx := c.ctx
	c.ctx = nil
	if ctx == nil {
		ctx = &api.Context{}
	}

	index := -1
LOOP:
	for i, cmd := range c.backlog {
		if cmd.IsAsync() {
			index = i
			break LOOP
		} else {
			c.Invoke(ctx, cmd)
		}
	}

	if index > -1 {
		dispatch := c.backlog[index]
		c.backlog = c.backlog[index+1:]
		c.dispatch(dispatch)
	} else {
		c.backlog = nil
	}

	c.mu.Unlock()
	return ctx
}

// Begin accepts a new packet and returns a working sequence of
// unprocessed bytes.
func (c *Conn) begin(packet []byte) (data []byte) {
	data = packet
	if len(c.in) > 0 {
		c.in = append(c.in, data...)
		data = c.in
	}
	return data
}

// End shift the stream to match the unprocessed data.
func (c *Conn) end(data []byte) {
	if len(data) > 0 {
		if len(data) != len(c.in) {
			c.in = append(c.in[:0], data...)
		}
	} else if len(c.in) > 0 {
		c.in = c.in[:0]
	}

	//if len(c.out) > 0 {
	//	c.outb = c.out
	//	c.out = nil
	//}
}
