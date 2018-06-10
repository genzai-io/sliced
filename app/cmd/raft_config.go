package cmd

import (
	"errors"
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
	"github.com/genzai-io/sliced/common/raft"
	"strings"
)

func init() { api.Register(&RaftConfig{}) }

// Reconfigures a Raft Service to be a single node Leader.
type RaftConfig struct {
	ID   api.RaftID
	raft api.RaftService
}

func (c *RaftConfig) Name() string   { return "RAFTCONFIG" }
func (c *RaftConfig) Help() string   { return "" }
func (c *RaftConfig) IsError() bool  { return false }
func (c *RaftConfig) IsWorker() bool { return true }

//
//
//
func (c *RaftConfig) Marshal(b []byte) []byte {
	if c.ID.DatabaseID < 0 {
		b = redcon.AppendArray(b, 1)
		b = redcon.AppendBulkString(b, c.Name())
	} else {
		b = redcon.AppendArray(b, 3)
		b = redcon.AppendBulkString(b, c.Name())
		b = redcon.AppendBulkInt32(b, c.ID.DatabaseID)
		b = redcon.AppendBulkInt32(b, c.ID.SliceID)
	}
	return b
}

//
//
//
func (c *RaftConfig) Parse(args [][]byte) Command {
	cmd := &RaftConfig{}

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

//
//
//
func (c *RaftConfig) Handle(ctx *Context) Reply {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	if c.raft == nil {
		return Err("ERR not exist")
	}

	future, err := c.raft.Configuration()
	if err != nil {
		return Error(err)
	}
	if err = future.Error(); err != nil {
		return Error(err)
	}

	return &RaftConfigReply{
		Index:   future.Index(),
		Servers: future.Configuration().Servers,
	}
}

//
//
//
type RaftConfigReply struct {
	Index   uint64
	Servers []raft.Server
}

//
//
//
func (c *RaftConfigReply) IsError() bool { return false }

//
//
//
func (r *RaftConfigReply) MarshalReply(b []byte) []byte {
	l := len(r.Servers)
	b = redcon.AppendArray(b, l+1)
	b = redcon.AppendInt(b, int64(r.Index))

	if l > 0 {
		for _, server := range r.Servers {
			b = redcon.AppendArray(b, 3)
			b = redcon.AppendBulkString(b, string(server.ID))
			b = redcon.AppendBulkString(b, string(server.Address))
			b = redcon.AppendBulkString(b, server.Suffrage.String())
		}
	}

	return b
}

//
//
//
func (r *RaftConfigReply) UnmarshalReply(packet []byte, args [][]byte) error {
	if len(args) < 2 {
		return errors.New("ERR invalid args")
	}
	var err error
	r.Index, err = strconv.ParseUint(string(args[1]), 10, 64)
	if err != nil {
		return err
	}
	for i := 4; i < len(args); i += 3 {
		var server raft.Server
		server.ID = raft.ServerID(args[i-2])
		server.Address = raft.ServerAddress(args[i-1])

		switch strings.ToLower(string(args[i])) {
		case "voter":
			server.Suffrage = raft.Voter
		case "nonvoter", "non-voter":
			server.Suffrage = raft.Nonvoter
		case "staging":
			server.Suffrage = raft.Staging
		default:
			server.Suffrage = -1
		}
	}
	return nil
}
