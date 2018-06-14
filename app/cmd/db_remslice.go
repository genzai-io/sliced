package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&DBRemSlice{}) }

// Demotes a Voting member to a Non-Voting member.
type DBRemSlice struct {
	Name_ string
}

func (c *DBRemSlice) Name() string   { return "-SLICE" }
func (c *DBRemSlice) Help() string   { return "" }
func (c *DBRemSlice) IsError() bool  { return false }
func (c *DBRemSlice) IsWorker() bool { return true }

func (c *DBRemSlice) Marshal(buf []byte) []byte {
	buf = resp.AppendArray(buf, 2)
	buf = resp.AppendBulkString(buf, c.Name())
	buf = resp.AppendBulkString(buf, c.Name_)
	return buf
}

func (c *DBRemSlice) Parse(args [][]byte) Command {
	cmd := &DBRemSlice{}

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

func (c *DBRemSlice) Handle(ctx *Context) Reply {
	reply := api.Array([]Reply{
		api.Int(10),
		String("hi"),
	})

	return reply
}
