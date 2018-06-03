package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(api.RaftBootstrap, &RaftBootstrap{}) }

// Reconfigures a Raft Service to be a single node Leader.
type RaftBootstrap struct {
	ID   api.RaftID
	raft api.RaftService
}

func (c *RaftBootstrap) IsChange() bool { return false }
func (c *RaftBootstrap) IsAsync() bool  { return true }

func (c *RaftBootstrap) Marshal(buf []byte) []byte {
	if c.ID.Schema < 0 {
		buf = redcon.AppendArray(buf, 1)
		buf = redcon.AppendBulkString(buf, api.RaftBootstrap)
	} else {
		buf = redcon.AppendArray(buf, 3)
		buf = redcon.AppendBulkString(buf, api.RaftBootstrap)
		buf = redcon.AppendBulkInt32(buf, c.ID.Schema)
		buf = redcon.AppendBulkInt32(buf, c.ID.Slice)
	}
	return buf
}

func (c *RaftBootstrap) Parse(ctx *Context) api.Command {
	cmd := &RaftBootstrap{}

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

func (c *RaftBootstrap) Handle(ctx *Context) {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	if c.raft == nil {
		ctx.Err("not exist")
		return
	}

	if err := c.raft.Bootstrap(); err != nil {
		ctx.Error(err)
		return
	}

	ctx.OK()
}

func (c *RaftBootstrap) Apply(ctx *Context) {}
