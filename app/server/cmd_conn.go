package server

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/cmd"
	"github.com/genzai-io/sliced/common/evio"
	"fmt"
	"github.com/genzai-io/sliced/common/redcon"
	"unsafe"
	"io"
)

var maxCommandBacklog = 10000

var emptyCtx = &api.Context{}

func ErrWake(err error) error {
	return fmt.Errorf("wake: %s", err.Error())
}

type Conn struct {
	api.Context

	done   bool        // flag to signal it's done

	// Durability setting
	durability api.Durability
	// Current raft context
	// This is used for the RaftTransport to support multiple Raft clusters
	// over the same port
	raft api.RaftService

	reason error

	// Buffers
	in  []byte
	out []byte

	// Locks
	loopMutex sync.Mutex

	// Counter to manage race conditions with the event loop and workers
	wakeProcessed uint64
	wakeRequired  uint64

	// Backlog
	multi   bool
	backlog []cmd.Command // commands that must wait until current worker finishes
	worker  cmd.Command

	onDetached func(rwc io.ReadWriteCloser)
	onData     func(in []byte) (out []byte, action evio.Action)

	// Stats
	statsTotalCommands uint64
	statsIngress       uint64
	statsEgress        uint64
	statsWakes         uint64

	workerStart        time.Time
	statsAsyncCommands uint64
	statsWorkers       uint64
	statsWorkerDur     time.Duration
}

func (c *Conn) Detach() error {
	c.Lock()
	c.Action = evio.Detach
	c.Unlock()
	c.Ev.Wake()
	return nil
}

func (c *Conn) OnDetach(rwc io.ReadWriteCloser) {
	if rwc != nil {
		rwc.Close()
	}
}

func (c *Conn) Close() error {
	c.Lock()
	c.Action = evio.Close
	conn := c.Ev
	c.Unlock()

	if conn != nil {
		return conn.Wake()
	}
	return nil
}

func (c *Conn) OnClosed() {
	c.Lock()
	c.done = true
	c.Action = evio.Close
	c.Ev = nil
	c.Unlock()
}

func (c *Conn) Conn() evio.Conn {
	return c.Ev
}

func (c *Conn) Stats() {
	c.Lock()
	c.Unlock()
}

func (c *Conn) GetKind() api.ConnKind {
	c.Lock()
	k := c.Kind
	c.Unlock()
	return k
}

func (c *Conn) SetKind(kind api.ConnKind) {
	c.Lock()
	c.Kind = kind
	c.Unlock()
}

func (c *Conn) Raft() api.RaftService {
	c.Lock()
	r := c.raft
	c.Unlock()
	return r
}

func (c *Conn) SetRaft(raft api.RaftService) {
	c.Lock()
	c.raft = raft
	c.Unlock()
}

func (c *Conn) Durability() api.Durability {
	return c.durability
}

// Invoked from the Worker
// This must be careful to synchronize access to the CommandConn's properties
// since this is called from a different goroutine from the Event Loop.
// However, since there can only be a single "active" worker for a connection
// at a time, there is no need to worry about synchronizing multiple Run() calls.
func (c *Conn) Run() {
	c.Lock()
	worker := c.worker
	c.Unlock()

	// Was the job already canceled?
	if worker == nil {
		return
	}

	// Run job.
	reply := worker.Handle(nil)
	if reply == nil {
		reply = api.Err("ERR not implemented")
	}

	dur := time.Now().Sub(c.workerStart)

	// Use spin-lock
	c.Lock()

	// Add stats
	c.statsWorkerDur += dur

	// Append to write buffer
	before := len(c.out)
	c.out = reply.MarshalReply(c.out)
	if len(c.out) == before {
		c.out = redcon.AppendError(c.out, "ERR not implemented")
	}

	// Increment wake tx
	c.wakeRequired++

	if len(c.backlog) > 0 {
		atomic.AddUint64(&c.statsWakes, 1)

		next := c.backlog[0]
		// Check if the next command is also a Worker
		if next.IsWorker() {
			worker = next
			// Start next worker
			c.worker = next
			// Pop it off the backlog
			c.backlog = c.backlog[1:]
		} else {
			worker = nil
			c.worker = nil
			// Should we process in the worker?
		}
	} else {
		worker = nil
		c.worker = nil
	}
	c.Unlock()

	// Dispatch the next worker if needed.
	// We could maybe use the same worker instance?
	if worker != nil {
		c.dispatch()
	}

	// Notify event loop of our write.
	if err := c.Ev.Wake(); err != nil {
		c.Lock()
		c.reason = ErrWake(err)
		c.Action = evio.Close
		c.Unlock()
	}
}

func (c *Conn) dispatch() {
	atomic.AddUint64(&c.statsWorkers, 1)
	Workers.Dispatch(c)
}

func (c *Conn) process(out []byte, backlog []api.Command) []byte {
	if c.multi {
		return nil
	}

	l := len(backlog)

	var (
		wrkidx  = -1
		i       int
		command api.Command
	)
	if l > 0 {
	LOOP:
		for i, command = range backlog {
			if command.IsWorker() {
				wrkidx = i
				break LOOP
			} else {
				before := len(out)
				// Run job.
				reply := command.Handle(nil)
				if reply != nil {
					out = reply.MarshalReply(out)
				}
				if len(out) == before {
					out = redcon.AppendError(out, "ERR not implemented")
				}
			}
		}
	}

	c.Lock()
	if wrkidx > -1 {
		c.worker = backlog[wrkidx]
		if wrkidx+1 < l {
			c.backlog = backlog[wrkidx+1:]
		} else {
			// Current worker was the last command in the backlog.
			// We can fully clear the backlog.
			c.backlog = nil
		}

		c.dispatch()
	}
	c.Unlock()

	return out
}

func (c *Conn) handle(packet []byte, args [][]byte) api.Command {
	name := *(*string)(unsafe.Pointer(&args[0]))

	// Find command
	command := api.Commands[name]
	if command != nil {
		command = command.Parse(args)
	}
	return command
}

// This is called from the Event Loop goroutine / thread
func (c *Conn) OnData(in []byte) (o []byte, action evio.Action) {
	// Loop lock
	c.loopMutex.Lock()
	defer c.loopMutex.Unlock()

	// Is this a "Wake"?
	if len(in) == 0 {
		c.Lock()
		// Flush write buffer
		o = c.out
		c.out = nil
		backlog := c.backlog
		c.backlog = nil
		c.Unlock()

		o = c.process(o, backlog)

		return
	}

	// Ingress
	atomic.AddUint64(&c.statsIngress, uint64(len(in)))

	action = c.Action
	switch action {
	case evio.Close, evio.Shutdown:
		return
	case evio.Detach:
	}

	// Were there any leftovers from the previous event?
	data := in
	if len(c.in) > 0 {
		data = append(c.in, data...)
		c.in = nil
	}

	var
	(
		packet   []byte
		complete bool
		args     [][]byte
		err      error
		command  api.Command
		commands []api.Command
	)

LOOP:
	for c.Action == evio.None {
		// Read next command.
		packet, complete, args, _, data, err = redcon.ParseNextCommand(data, args[:0])
		_ = packet

		if err != nil {
			action = evio.Close
			o = redcon.AppendError(o, err.Error())
			break LOOP
		}

		// Do we need more data?
		if !complete {
			// Exit loop.
			break LOOP
		}

		if len(args) == 0 {
			break LOOP
		}

		if c.Parse == nil {
			command = c.handle(packet, args)
		} else {
			command = c.Parse(packet, args)
		}

		if command == nil {
			command = cmd.Err(fmt.Sprintf("ERR command '%s' not found", args[0]))
		}

		commands = append(commands, command)
	}

	if len(commands) > 0 {
		// Add stats
		atomic.AddUint64(&c.statsTotalCommands, uint64(len(commands)))

		var backlog []api.Command = nil

		c.Lock()
		o = c.out
		c.out = nil
		action = c.Action
		worker := c.worker

		if len(c.backlog) > 0 {
			backlog = append(c.backlog, commands...)
		} else {
			backlog = commands
		}

		c.backlog = backlog
		c.Unlock()

		if worker == nil && len(backlog) > 0 {
			o = c.process(o, backlog)
		}
	}

	// Egress stats
	atomic.AddUint64(&c.statsEgress, uint64(len(o)))

	if action != evio.None {
		action = c.Action
		return o, action
	}

	// Are there any leftovers (partial commands)?
	if len(data) > 0 {
		if len(data) != len(c.in) {
			c.in = append(c.in[:0], data...)
		}
	} else if len(c.in) > 0 {
		c.in = c.in[:0]
	}

	action = c.Action

	return o, action
}
