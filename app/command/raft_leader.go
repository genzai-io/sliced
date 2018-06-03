package command

import (
	"strconv"

	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/common/redcon"
)

func init() { api.Register(api.RaftLeaderName, &RaftLeader{}) }

type RaftLeader struct {
	ID   api.RaftID
	raft api.RaftService
}

func (c *RaftLeader) IsChange() bool { return false }
func (c *RaftLeader) IsAsync() bool  { return true }

func (c *RaftLeader) Marshal(buf []byte) []byte {
	if c.ID.Schema < 0 {
		buf = redcon.AppendArray(buf, 1)
		buf = redcon.AppendBulkString(buf, api.RaftLeaderName)
	} else {
		buf = redcon.AppendArray(buf, 3)
		buf = redcon.AppendBulkString(buf, api.RaftLeaderName)
		buf = redcon.AppendBulkInt32(buf, c.ID.Schema)
		buf = redcon.AppendBulkInt32(buf, c.ID.Slice)
	}
	return buf
}

func (c *RaftLeader) Parse(ctx *Context) api.Command {
	cmd := &RaftLeader{}

	switch len(ctx.Args) {
	default:
		ctx.Err("expected 0 or 2 params")
		return cmd

	case 1:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		return cmd

	case 3:
		// Parse schema
		schemaID, err := strconv.Atoi(string(ctx.Args[1]))
		if err != nil {
			ctx.Err("invalid schema id: " + string(ctx.Args[1]))
			return cmd
		}
		cmd.ID.Schema = int32(schemaID)

		// Parse slice
		sliceID, err := strconv.Atoi(string(ctx.Args[2]))
		if err != nil {
			ctx.Err("invalid slice id: " + string(ctx.Args[2]))
			return cmd
		}
		cmd.ID.Slice = int32(sliceID)
		return cmd
	}
}

func (c *RaftLeader) Handle(ctx *Context) {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	if c.raft == nil {
		ctx.Err("not exist")
		return
	}

	ctx.AppendString(string(c.raft.Leader()))
}

func (c *RaftLeader) Apply(ctx *Context) {}
