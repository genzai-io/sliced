package cmd

import (
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(api.RaftLeaderName, &RaftLeader{}) }

type RaftLeader struct {
	ID   api.RaftID
	raft api.RaftService
}

func (c *RaftLeader) IsError() bool  { return false }
func (c *RaftLeader) IsChange() bool { return false }
func (c *RaftLeader) IsWorker() bool  { return true }

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
		cmd.ID.Schema = int32(schemaID)

		// Parse slice
		sliceID, err := strconv.Atoi(string(args[2]))
		if err != nil {
			return Err("ERR invalid slice id: " + string(args[2]))
		}
		cmd.ID.Slice = int32(sliceID)
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
