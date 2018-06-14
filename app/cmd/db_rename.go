package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&DBRename{}) }

// Demotes a Voting member to a Non-Voting member.
type DBRename struct {
	Name_ string
}

func (c *DBRename) Name() string   { return "*DB" }
func (c *DBRename) Help() string   { return "" }
func (c *DBRename) IsError() bool  { return false }
func (c *DBRename) IsWorker() bool { return true }

func (c *DBRename) Marshal(buf []byte) []byte {
	buf = resp.AppendArray(buf, 2)
	buf = resp.AppendBulkString(buf, c.Name())
	buf = resp.AppendBulkString(buf, c.Name_)
	return buf
}

func (c *DBRename) Parse(args [][]byte) Command {
	cmd := &DBRename{}

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

func (c *DBRename) Handle(ctx *Context) Reply {
	reply := api.Array([]Reply{
		String("old"),
		String(c.Name_),
	})

	return reply
}
