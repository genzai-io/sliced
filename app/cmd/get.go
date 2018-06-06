package cmd

import "github.com/genzai-io/sliced/common/redcon"

type Get struct {
	Key string
}

func (c *Get) Name() string { return "GET" }
func (c *Get) Help() string { return "" }
func (c *Get) IsError() bool  { return false }
func (c *Get) IsChange() bool { return false }
func (c *Get) IsWorker() bool { return true }

func (c *Get) Marshal(b []byte) []byte {
	b = redcon.AppendArray(b, 2)
	b = redcon.AppendBulkString(b, "GET")
	b = redcon.AppendBulkString(b, c.Key)
	return b
}

func (c *Get) Parse(args [][]byte) Command {
	cmd := &Get{}
	return cmd
}

func (c *Get) Handle(ctx *Context) Reply {
	return Ok
}
