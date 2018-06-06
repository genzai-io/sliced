package cmd

import (
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { Register(&Multi{}) }

type Multi struct{}

func (c *Multi) Name() string   { return "MULTI" }
func (c *Multi) Help() string   { return "" }
func (c *Multi) IsError() bool  { return false }
func (c *Multi) IsWorker() bool { return false }

func (c *Multi) Marshal(b []byte) []byte {
	b = redcon.AppendArray(b, 1)
	b = redcon.AppendBulkString(b, c.Name())
	return b
}

func (c *Multi) Parse(args [][]byte) Command {
	cmd := &Multi{}

	switch len(args) {
	default:
		return Err("ERR expected 0 params")

	case 1:
		return cmd
	}
}

func (c *Multi) Handle(ctx *Context) Reply {
	return Ok
}
