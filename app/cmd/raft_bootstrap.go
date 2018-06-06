package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(&RaftBootstrap{}) }

// Reconfigures a Raft Service to be a single node Leader.
type RaftBootstrap struct {
	ID   api.RaftID
	raft api.RaftService
}

func (c *RaftBootstrap) Name() string   { return "BOOTSTRAP" }
func (c *RaftBootstrap) Help() string   { return "" }
func (c *RaftBootstrap) IsError() bool  { return false }
func (c *RaftBootstrap) IsWorker() bool { return true }

func (c *RaftBootstrap) Marshal(buf []byte) []byte {
	if c.ID.DatabaseID < 0 {
		buf = redcon.AppendArray(buf, 1)
		buf = redcon.AppendBulkString(buf, c.Name())
	} else {
		buf = redcon.AppendArray(buf, 3)
		buf = redcon.AppendBulkString(buf, c.Name())
		buf = redcon.AppendBulkInt32(buf, c.ID.DatabaseID)
		buf = redcon.AppendBulkInt32(buf, c.ID.SliceID)
	}
	return buf
}

func (c *RaftBootstrap) Parse(args [][]byte) Command {
	cmd := &RaftBootstrap{}

	switch len(args) {
	default:
		return Err("ERR expected 0 or 2 params")

	case 1:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.ID = api.GlobalRaftID
		return cmd

	case 3:
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
		return cmd
	}
}

func (c *RaftBootstrap) Handle(ctx *Context) Reply {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	if c.raft == nil {
		return Err("ERR not exist")
	}

	if err := c.raft.Bootstrap(); err != nil {
		return Error(err)
	}

	return Ok
}
