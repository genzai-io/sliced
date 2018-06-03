package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(api.RaftJoinName, &RaftJoin{}) }

type RaftJoin struct {
	ID      api.RaftID
	Address string
	Voter   bool

	raft api.RaftService
}

func (c *RaftJoin) IsChange() bool { return false }
func (c *RaftJoin) IsAsync() bool  { return true }

func (c *RaftJoin) Marshal(buf []byte) []byte {
	if c.ID.Schema < 0 {
		buf = redcon.AppendArray(buf, 3)
		buf = redcon.AppendBulk(buf, api.RaftJoin)
		buf = redcon.AppendBulkString(buf, c.Address)
		if c.Voter {
			buf = redcon.AppendBulkInt(buf, 1)
		} else {
			buf = redcon.AppendBulkInt(buf, 0)
		}
	} else {
		buf = redcon.AppendArray(buf, 5)
		buf = redcon.AppendBulk(buf, api.RaftJoin)
		buf = redcon.AppendBulkInt32(buf, c.ID.Schema)
		buf = redcon.AppendBulkInt32(buf, c.ID.Slice)
		buf = redcon.AppendBulkString(buf, c.Address)
		if c.Voter {
			buf = redcon.AppendBulkInt(buf, 1)
		} else {
			buf = redcon.AppendBulkInt(buf, 0)
		}
	}
	return buf
}

func (c *RaftJoin) Parse(ctx *Context) api.Command {
	cmd := &RaftJoin{}

	switch len(ctx.Args) {
	default:
		ctx.Err("invalid params")
		return cmd

	case 2:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		cmd.Address = string(ctx.Args[1])
		cmd.Voter = true
		return cmd

	case 3:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		cmd.Address = string(ctx.Args[1])
		voter, err := api.ParseBool(ctx.Args[2])
		if err != nil {
			ctx.Err("invalid 'voter' param: " + string(ctx.Args[2]))
			return cmd
		}
		cmd.Voter = voter
		return cmd

	case 4:
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

		// Set address
		cmd.Address = string(ctx.Args[3])
		return cmd
	case 5:
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

		// Set address
		cmd.Address = string(ctx.Args[3])
		voter, err := api.ParseBool(ctx.Args[2])
		if err != nil {
			ctx.Err("invalid 'voter' param: " + string(ctx.Args[2]))
			return cmd
		}
		cmd.Voter = voter
		return cmd
	}
	return cmd
}

func (c *RaftJoin) Handle(ctx *Context) {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	// Do we have a matching Raft service
	if c.raft == nil {
		ctx.Err("not exist")
		return
	}

	if err := c.raft.Join(c.Address, c.Voter); err != nil {
		ctx.Error(err)
		return
	}

	ctx.OK()
}

func (c *RaftJoin) Apply(ctx *Context) {}
