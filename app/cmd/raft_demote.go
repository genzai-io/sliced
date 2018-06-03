package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(api.RaftDemote, &RaftDemote{}) }

// Demotes a Voting member to a Non-Voting member.
type RaftDemote struct {
	ID      api.RaftID
	Address string

	raft api.RaftService
}

func (c *RaftDemote) IsChange() bool { return false }
func (c *RaftDemote) IsAsync() bool  { return true }

func (c *RaftDemote) Marshal(buf []byte) []byte {
	if c.ID.Schema < 0 {
		buf = redcon.AppendArray(buf, 2)
		buf = redcon.AppendBulkString(buf, api.RaftDemote)
		buf = redcon.AppendBulkString(buf, c.Address)
	} else {
		buf = redcon.AppendArray(buf, 4)
		buf = redcon.AppendBulkString(buf, api.RaftDemote)
		buf = redcon.AppendBulkInt32(buf, c.ID.Schema)
		buf = redcon.AppendBulkInt32(buf, c.ID.Slice)
		buf = redcon.AppendBulkString(buf, c.Address)
	}
	return buf
}

func (c *RaftDemote) Parse(ctx *Context) api.Command {
	cmd := &RaftDemote{}

	switch len(ctx.Args) {
	default:
		ctx.Err("invalid params")
		return cmd

	case 2:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		cmd.Address = string(ctx.Args[1])
		return cmd

	case 4:
		// Parse schema
		schemaID, err := strconv.Atoi(string(ctx.Args[1]))
		if err != nil {
			ctx.Err("invalid schema id: " + string(ctx.Args[2]))
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

		// Set address
		cmd.Address = string(ctx.Args[3])
	}
	return cmd
}

func (c *RaftDemote) Handle(ctx *Context) {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	// Do we have a matching Raft service
	if c.raft == nil {
		ctx.Err("not exist")
		return
	}

	if err := c.raft.Demote(c.Address); err != nil {
		ctx.Error(err)
		return
	}

	ctx.OK()
}

func (c *RaftDemote) Apply(ctx *Context) {}
