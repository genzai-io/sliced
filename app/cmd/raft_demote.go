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

func (c *RaftDemote) IsError() bool  { return false }
func (c *RaftDemote) IsChange() bool { return false }
func (c *RaftDemote) IsWorker() bool  { return true }

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

func (c *RaftDemote) Parse(args [][]byte) Command {
	cmd := &RaftDemote{}

	switch len(args) {
	default:
		return Err("ERR invalid params")

	case 2:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		cmd.Address = string(args[1])
		return cmd

	case 4:
		// Parse schema
		schemaID, err := strconv.Atoi(string(args[1]))
		if err != nil {
			return Err("ERR invalid schema id: " + string(args[2]))
		}
		cmd.ID.Schema = int32(schemaID)

		// Parse slice
		sliceID, err := strconv.Atoi(string(args[2]))
		if err != nil {
			return Err("ERR invalid slice id: " + string(args[2]))
		}
		cmd.ID.Slice = int32(sliceID)

		// Set address
		cmd.Address = string(args[3])
	}
	return cmd
}

func (c *RaftDemote) Handle(ctx *Context) Reply {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	// Do we have a matching Raft service
	if c.raft == nil {
		return Err("ERR not exist")
	}

	if err := c.raft.Demote(c.Address); err != nil {
		return Error(err)
	}

	return Ok
}
