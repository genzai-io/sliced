package api

import (
	"errors"

	"github.com/genzai-io/sliced/common/redcon"
)

//
type Durability int

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

// Represents a series of pipelined Commands in the context
// of a single call for a single connection or for Raft log applying.
type Context struct {
	Conn   Conn     // Connection
	Out    []byte   // Output buffer to write to connection
	Index  int      // Index of command
	Name   string   // Current command name
	RaftID RaftID   // Raft ID
	Slot   uint16   // Slot the command is for
	Key    string   // Extract if it exists
	Packet []byte   // Raw byte slice of current command
	Args   [][]byte // Current command args. First index is the command name.
	Ops    int      // Count of total operations that occurred from processing commands
	ERR    error    // Error for current command

	// Transactional holders
	//Set *table.Table

	raft RaftService

	Changes map[RaftID]*ChangeSet
}

type ChangeSet struct {
	Cmds []Command
	Data []byte
}

func (c *Context) Int(index int, or int) int {
	return or
}

// Append a redis "OK" message
func (c *Context) OK() []byte {
	c.Out = redcon.AppendOK(c.Out)
	return c.Out
}

func (c *Context) Error(err error) []byte {
	if err == nil {
		return c.Out
	}
	c.ERR = err
	c.Out = redcon.AppendError(c.Out, err.Error())
	return c.Out
}

// Append a redis "ERR" message
func (c *Context) Err(msg string) []byte {
	c.ERR = errors.New(msg)
	c.Out = redcon.AppendError(c.Out, msg)
	return c.Out
}

// Append a redis "NULL" message
func (c *Context) AppendNull() []byte {
	c.Out = redcon.AppendNull(c.Out)
	return c.Out
}

// Append a redis "BULK" message
func (c *Context) AppendString(bulk string) []byte {
	c.Out = redcon.AppendBulkString(c.Out, bulk)
	return c.Out
}

// Append a redis "BULK" message
func (c *Context) Append(bulk []byte) []byte {
	c.Out = redcon.AppendBulk(c.Out, bulk)
	return c.Out
}

// Append a redis "INT" message
func (c *Context) AppendInt(val int) []byte {
	c.Out = redcon.AppendInt(c.Out, int64(val))
	return c.Out
}

// Append a redis "INT" message
func (c *Context) AppendInt64(val int64) []byte {
	c.Out = redcon.AppendInt(c.Out, val)
	return c.Out
}

// Append a redis "INT" message
func (c *Context) AppendUint64(val uint64) []byte {
	c.Out = redcon.AppendInt(c.Out, int64(val))
	return c.Out
}

func (c *Context) Commit() {
	if len(c.Changes) > 0 {
		c.Conn.Handler().Commit(c)
	}
}

func (c *Context) HasChanges() bool {
	return len(c.Changes) > 0
}

func (c *Context) AddChange(cmd Command, data []byte) {
	if c.Changes == nil {
		c.Changes = map[RaftID]*ChangeSet{
			c.RaftID: {
				Cmds: []Command{cmd},
				Data: data,
			},
		}
	} else {
		set, ok := c.Changes[c.RaftID]
		if !ok {
			c.Changes[c.RaftID] = &ChangeSet{
				Cmds: []Command{cmd},
				Data: c.Packet,
			}
		} else {
			set.Cmds = append(set.Cmds, cmd)
			set.Data = append(set.Data, c.Packet...)
		}
	}
}
