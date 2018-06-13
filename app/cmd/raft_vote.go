package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&RaftVote{}) }

type RaftVote struct {
	Payload []byte
}

func (c *RaftVote) Name() string   { return "VOTE" }
func (c *RaftVote) Help() string   { return "" }
func (c *RaftVote) IsError() bool  { return false }
func (c *RaftVote) IsWorker() bool { return true }

func (c *RaftVote) Marshal(buf []byte) []byte {
	buf = resp.AppendArray(buf, 2)
	buf = resp.AppendBulkString(buf, c.Name())
	if c.Payload == nil {
		buf = resp.AppendBulk(buf, []byte{})
	} else {
		buf = resp.AppendBulk(buf, c.Payload)
	}

	return buf
}

func (c *RaftVote) Parse(args [][]byte) Command {
	cmd := &RaftVote{}

	switch len(args) {
	default:
		return Err("ERR expected 1 param")

	case 2:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.Payload = append(cmd.Payload, args[1]...)
		return cmd
	}
}

func (c *RaftVote) Handle(ctx *Context) Reply {
	raft := ctx.Raft
	if raft == nil {
		return Err("ERR not raft connection")
	}
	return ctx.Raft.Vote(c.Payload)
}
