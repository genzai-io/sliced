package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(&RaftJoin{}) }

type RaftJoin struct {
	ID      api.RaftID
	Address string
	Voter   bool

	raft api.RaftService
}

func (c *RaftJoin) Name() string   { return "JOIN" }
func (c *RaftJoin) Help() string   { return "" }
func (c *RaftJoin) IsError() bool  { return false }
func (c *RaftJoin) IsWorker() bool { return true }

func (c *RaftJoin) Marshal(buf []byte) []byte {
	if c.ID.DatabaseID < 0 {
		buf = redcon.AppendArray(buf, 3)
		buf = redcon.AppendBulkString(buf, c.Name())
		buf = redcon.AppendBulkString(buf, c.Address)
		if c.Voter {
			buf = redcon.AppendBulkInt(buf, 1)
		} else {
			buf = redcon.AppendBulkInt(buf, 0)
		}
	} else {
		buf = redcon.AppendArray(buf, 5)
		buf = redcon.AppendBulkString(buf, c.Name())
		buf = redcon.AppendBulkInt32(buf, c.ID.DatabaseID)
		buf = redcon.AppendBulkInt32(buf, c.ID.SliceID)
		buf = redcon.AppendBulkString(buf, c.Address)
		if c.Voter {
			buf = redcon.AppendBulkInt(buf, 1)
		} else {
			buf = redcon.AppendBulkInt(buf, 0)
		}
	}
	return buf
}

func (c *RaftJoin) Parse(args [][]byte) Command {
	cmd := &RaftJoin{}

	switch len(args) {
	default:
		return Err("ERR invalid params")

	case 2:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		cmd.Address = string(args[1])
		cmd.Voter = true
		return cmd

	case 3:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		cmd.Address = string(args[1])
		voter, err := api.ParseBool(args[2])
		if err != nil {
			return Err("ERR invalid 'voter' param: " + string(args[2]))
		}
		cmd.Voter = voter
		return cmd

	case 4:
		// Parse schema
		schemaID, err := strconv.Atoi(string(args[1]))
		if err != nil {
			return Err("ERR invalid schema id: " + string(args[1]))
		}
		cmd.ID.DatabaseID = int32(schemaID)

		// Parse slice
		sliceID, err := strconv.Atoi(string(args[2]))
		if err != nil {
			return Err("ERR invalid slice id: " + string(args[2]))
		}
		cmd.ID.SliceID = int32(sliceID)

		// Set address
		cmd.Address = string(args[3])
		return cmd
	case 5:
		// Parse schema
		schemaID, err := strconv.Atoi(string(args[1]))
		if err != nil {
			return Err("ERR invalid schema id: " + string(args[1]))
		}
		cmd.ID.DatabaseID = int32(schemaID)

		// Parse slice
		sliceID, err := strconv.Atoi(string(args[2]))
		if err != nil {
			return Err("ERR invalid slice id: " + string(args[2]))
		}
		cmd.ID.SliceID = int32(sliceID)

		// Set address
		cmd.Address = string(args[3])
		voter, err := api.ParseBool(args[2])
		if err != nil {
			return Err("ERR invalid 'voter' param: " + string(args[2]))
		}
		cmd.Voter = voter
		return cmd
	}
	return cmd
}

func (c *RaftJoin) Handle(ctx *Context) Reply {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	// Do we have a matching Raft service
	if c.raft == nil {
		return Err("ERR not exist")
	}

	if err := c.raft.Join(c.Address, c.Voter); err != nil {
		return Error(err)
	}

	return Ok
}
