package cmd

import (
	"github.com/genzai-io/sliced/common/resp"
	"github.com/genzai-io/sliced/app/api"
)

func init() { api.Register(&Set{}) }

type Set struct {
	Key   string
	Value string
}

func (c *Set) Name() string   { return "SET" }
func (c *Set) Help() string   { return "" }
func (c *Set) IsError() bool  { return false }
func (c *Set) IsWorker() bool { return false }

func (c *Set) Marshal(b []byte) []byte {
	b = resp.AppendArray(b, 2)
	b = resp.AppendBulkString(b, c.Name())
	b = resp.AppendBulkString(b, c.Key)
	return b
}

func (c *Set) Parse(args [][]byte) Command {
	if len(args) > 2 {
		return &Set{
			Key:   string(args[1]),
			Value: string(args[2]),
		}
	} else {
		return Err("ERR invalid params")
	}
}

func (c *Set) Handle(ctx *Context) Reply {
	return Ok
}
