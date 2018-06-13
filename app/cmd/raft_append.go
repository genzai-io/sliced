package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&RaftAppend{}) }

type RaftAppend struct {
	Payload []byte
}

func (c *RaftAppend) Name() string   { return "A" }
func (c *RaftAppend) Help() string   { return "" }
func (c *RaftAppend) IsError() bool  { return false }
func (c *RaftAppend) IsWorker() bool { return true }

func (c *RaftAppend) Marshal(buf []byte) []byte {
	buf = resp.AppendArray(buf, 2)
	buf = resp.AppendBulkString(buf, c.Name())
	if c.Payload == nil {
		buf = resp.AppendBulk(buf, []byte{})
	} else {
		buf = resp.AppendBulk(buf, c.Payload)
	}

	return buf
}

func (c *RaftAppend) Parse(args [][]byte) Command {
	cmd := &RaftAppend{}

	switch len(args) {
	default:
		return Err("ERR expected 1 param")

	case 2:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.Payload = append(cmd.Payload, args[1]...)
		return cmd
	}
}

func (c *RaftAppend) Handle(ctx *Context) Reply {
	raft := ctx.Raft
	if raft == nil {
		return Err("ERR not raft connection")
	}
	return ctx.Raft.Append(c.Payload)
}
