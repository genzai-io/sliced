package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&DBAddSlice{}) }

// Demotes a Voting member to a Non-Voting member.
type DBAddSlice struct {
	Name_ string
}

func (c *DBAddSlice) Name() string   { return "+SLICE" }
func (c *DBAddSlice) Help() string   { return "" }
func (c *DBAddSlice) IsError() bool  { return false }
func (c *DBAddSlice) IsWorker() bool { return true }

func (c *DBAddSlice) Marshal(buf []byte) []byte {
	buf = resp.AppendArray(buf, 2)
	buf = resp.AppendBulkString(buf, c.Name())
	buf = resp.AppendBulkString(buf, c.Name_)
	return buf
}

func (c *DBAddSlice) Parse(args [][]byte) Command {
	cmd := &DBAddSlice{}

	switch len(args) {
	default:
		return Err("ERR invalid params")

	case 2:
		// Set schema and slice to -1 indicating we want the global store raft
		cmd.Name_ = string(args[1])
		return cmd
	}
	return cmd
}

func (c *DBAddSlice) Handle(ctx *Context) Reply {
	reply := api.Array([]Reply{
		api.Int(10),
		String("hi"),
	})

	return reply
}
