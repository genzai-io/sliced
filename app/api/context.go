package api

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/genzai-io/sliced/common/evio"
	"github.com/genzai-io/sliced/common/resp"
)

func ErrWake(err error) error {
	return fmt.Errorf("wake: %s", err.Error())
}

//
type Durability int32

const (
	Low    Durability = -1
	Medium Durability = 0
	High   Durability = 1
)

type ContextPurpose byte

const (
	Parse ContextPurpose = 0
	Apply ContextPurpose = 1
	Log   ContextPurpose = 2
)

type ContextStats struct {
}

type Multi int

const (
	SINGLE  Multi = iota
	MULTI
	DISCARD
	EXEC
)

type WithConnection interface {
	Conn() evio.Conn

	ConnAction() evio.Action
}

// Represents a series of pipelined Commands in the context
// of a single call for a single connection or for Raft log applying.
type Context struct {
	Reason error

	Action evio.Action
	Kind   ConnKind

	// Durability setting
	Durability Durability

	Multi Multi

	// Assigned raft context
	// This is used for the RaftTransport to support multiple Raft clusters
	// over the same port
	Raft RaftService // Raft service for Raft connections

	Parse func(packet []byte, args [][]byte) Command
}

func (c *Context) GetKind() ConnKind {
	k := c.Kind
	return k
}

func (c *Context) SetKind(kind ConnKind) {
	c.Kind = kind
}

func (c *Context) GetRaft() RaftService {
	r := c.Raft
	return r
}

func (c *Context) SetRaft(raft RaftService) {
	c.Raft = raft
}

func (c *Context) GetDurability() Durability {
	return Durability(atomic.LoadInt32((*int32)(&c.Durability)))
}

func (c *Context) handleReply(reply CommandReply) {
}

func (c *Context) onWake() {

}

func (c *Context) onReply(reply CommandReply) {
	// Black-hole
}

func (c *Context) processMulti() {

}

func (c *Context) AppendCommand(b []byte, command Command) []byte {
	if command == nil {
		command = Err("ERR nil command")
	}
	reply := command.Handle(c)
	if reply == nil {
		reply = Err("ERR nil reply for command '" + command.Name() + "'")
	}

	before := len(b)
	b = reply.MarshalReply(b)
	if len(b) == before {
		b = resp.AppendError(b, "ERR empty reply for command '"+command.Name()+"'")
	}
	return b
}

type ContextReader struct {
}

type ContextInput struct {
	Context

	mu sync.Mutex

	// Buffers
	In  []byte // in/ingress or "read" buffer
	Out []byte // out/egress or "write" buffer

	// Backlog
	Multi     bool
	MultiList []Command
	Backlog   []Command // commands that must wait until current worker finishes

	// Stats
	statsTotalCommands uint64
	statsIngress       uint64
	statsEgress        uint64
}

func (c *ContextInput) InputReply(in []byte) (o []CommandReply) {
	return
}

//func (c *ContextInput) Read()

// Parses a raw stream into a series of commands
func (c *ContextInput) Input(in []byte) (o []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Is this a "Wake"?
	if len(in) == 0 {
		return nil
	}

	// Ingress
	atomic.AddUint64(&c.statsIngress, uint64(len(in)))

	// Were there any leftovers from the previous event?
	data := in
	if len(c.In) > 0 {
		data = append(c.In, data...)
		c.In = nil
	}

	var
	(
		packet   []byte
		complete bool
		args     [][]byte
		err      error
		command  Command
		commands []Command
	)

	inMulti := c.Multi

LOOP:
	for {
		// Read next command.
		packet, complete, args, _, data, err = resp.ParseNextCommand(data, args[:0])
		_ = packet

		if err != nil {
			c.Reason = err
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
			command = ParseCommand(packet, args)
		} else {
			command = c.Parse(packet, args)
		}

		if command == nil {
			command = Err(fmt.Sprintf("ERR command '%s' not found", args[0]))
		}

		if inMulti {
			if !c.Multi {
				// Exec currentMulti
			} else {
				c.MultiList = append(c.MultiList, command)
				o = resp.AppendQueued(o)
			}
		} else {
			reply := command.Handle(&c.Context)
			if reply == nil {
				reply = Err("ERR nil reply for command '" + command.Name() + "'")
			}

			before := len(o)
			o = reply.MarshalReply(o)
			if len(o) == before {
				o = resp.AppendError(o, "ERR empty reply for command '"+command.Name()+"'")
			}
		}

		commands = append(commands, command)
	}

	if len(commands) > 0 {
		// Add stats
		atomic.AddUint64(&c.statsTotalCommands, uint64(len(commands)))

		var backlog []Command = nil

		if len(c.Backlog) > 0 {
			backlog = append(c.Backlog, commands...)
		} else {
			backlog = commands
		}

		for _, command := range backlog {
			reply := command.Handle(&c.Context)
			if reply == nil {
				reply = Err("ERR nil reply for command '" + command.Name() + "'")
			}

			before := len(o)
			o = reply.MarshalReply(o)
			if len(o) == before {
				o = resp.AppendError(o, "ERR empty reply for command '"+command.Name()+"'")
			}
		}
	}

	// Egress stats
	atomic.AddUint64(&c.statsEgress, uint64(len(o)))

	// Are there any leftovers (partial commands)?
	if len(data) > 0 {
		if len(data) != len(c.In) {
			c.In = append(c.In[:0], data...)
		}
	} else if len(c.In) > 0 {
		c.In = c.In[:0]
	}

	return
}
