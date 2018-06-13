package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&RaftLeader{}) }

type RaftLeader struct {
	ID   api.RaftID
	raft api.RaftService
}

func (c *RaftLeader) Name() string   { return "LEADER" }
func (c *RaftLeader) Help() string   { return "" }
func (c *RaftLeader) IsError() bool  { return false }
func (c *RaftLeader) IsWorker() bool { return false }

func (c *RaftLeader) Marshal(buf []byte) []byte {
	if c.ID.DatabaseID < 0 {
		buf = resp.AppendArray(buf, 1)
		buf = resp.AppendBulkString(buf, c.Name())
	} else {
		buf = resp.AppendArray(buf, 3)
		buf = resp.AppendBulkString(buf, c.Name())
		buf = resp.AppendBulkInt32(buf, c.ID.DatabaseID)
		buf = resp.AppendBulkInt32(buf, c.ID.SliceID)
	}
	return buf
}

func (c *RaftLeader) Parse(args [][]byte) Command {
	cmd := &RaftLeader{}

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

func (c *RaftLeader) Handle(ctx *Context) Reply {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	if c.raft == nil {
		return Err("ERR not exist")
	}

	return String(c.raft.Leader())
}
