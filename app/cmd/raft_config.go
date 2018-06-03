package cmd

import (
	"encoding/json"
	"strconv"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(api.RaftConfigName, &RaftConfig{}) }

// Reconfigures a Raft Service to be a single node Leader.
type RaftConfig struct {
	ID   api.RaftID
	raft api.RaftService
}

func (c *RaftConfig) IsChange() bool { return false }
func (c *RaftConfig) IsAsync() bool  { return true }

func (c *RaftConfig) Marshal(buf []byte) []byte {
	if c.ID.Schema < 0 {
		buf = redcon.AppendArray(buf, 1)
		buf = redcon.AppendBulkString(buf, api.RaftConfigName)
	} else {
		buf = redcon.AppendArray(buf, 3)
		buf = redcon.AppendBulkString(buf, api.RaftConfigName)
		buf = redcon.AppendBulkInt32(buf, c.ID.Schema)
		buf = redcon.AppendBulkInt32(buf, c.ID.Slice)
	}
	return buf
}

func (c *RaftConfig) Parse(ctx *Context) api.Command {
	cmd := &RaftConfig{}

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

func (c *RaftConfig) Handle(ctx *Context) {
	if c.raft == nil {
		// Find Raft
		c.raft = api.GetRaftService(c.ID)
	}

	if c.raft == nil {
		ctx.Err("not exist")
		return
	}

	future, err := c.raft.Configuration()
	if err != nil {
		ctx.Error(err)
		return
	}
	if err = future.Error(); err != nil {
		ctx.Error(err)
		return
	}

	var out []byte
	out = redcon.AppendArray(out, len(future.Configuration().Servers)+1)
	out = redcon.AppendInt(out, int64(future.Index()))
	for _, key := range future.Configuration().Servers {
		j, err := json.Marshal(key)
		if err != nil {
			out = redcon.AppendBulkString(out, ""+err.Error())
		} else {
			out = redcon.AppendBulk(out, j)
		}
	}

	ctx.Out = append(ctx.Out, out...)
}

func (c *RaftConfig) Apply(ctx *Context) {}
