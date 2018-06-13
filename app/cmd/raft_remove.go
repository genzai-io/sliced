package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&RaftRemove{}) }

// Demotes a Voting member to a Non-Voting member.
type RaftRemove struct {
	ID      api.RaftID
	Address string

	raft api.RaftService
}

func (c *RaftRemove) Name() string   { return "REMOVE" }
func (c *RaftRemove) Help() string   { return "" }
func (c *RaftRemove) IsError() bool  { return false }
func (c *RaftRemove) IsWorker() bool { return true }

func (c *RaftRemove) Marshal(buf []byte) []byte {
	if c.ID.DatabaseID < 0 {
		buf = resp.AppendArray(buf, 2)
		buf = resp.AppendBulkString(buf, c.Name())
		buf = resp.AppendBulkString(buf, c.Address)
	} else {
		buf = resp.AppendArray(buf, 4)
		buf = resp.AppendBulkString(buf, c.Name())
		buf = resp.AppendBulkInt32(buf, c.ID.DatabaseID)
		buf = resp.AppendBulkInt32(buf, c.ID.SliceID)
		buf = resp.AppendBulkString(buf, c.Address)
	}
	return buf
}

func (c *RaftRemove) Parse(args [][]byte) Command {
	cmd := &RaftRemove{}

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
		cmd.ID.DatabaseID = int32(schemaID)

		// Parse slice
		sliceID, err := strconv.Atoi(string(args[2]))
		if err != nil {
			return Err("ERR invalid slice id: " + string(args[2]))
		}
		cmd.ID.DatabaseID = int32(sliceID)

		// Set address
		cmd.Address = string(args[3])
	}
	return cmd
}

func (c *RaftRemove) Handle(ctx *Context) Reply {
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
