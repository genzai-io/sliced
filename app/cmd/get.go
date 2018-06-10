package cmd

import (
	"github.com/genzai-io/sliced/common/redcon"
	"github.com/genzai-io/sliced/app/api"
)

func init() { api.Register(&Get{}) }

type Get struct {
	Key string
}

func (c Get) Name() string   { return "GET" }
func (c Get) Help() string   { return "" }
func (c Get) IsError() bool  { return false }
func (c Get) IsWorker() bool { return false }

func (c Get) Marshal(b []byte) []byte {
	b = redcon.AppendArray(b, 2)
	b = redcon.AppendBulkString(b, "GET")
	b = redcon.AppendBulkString(b, c.Key)
	return b
}

func (c Get) Parse(args [][]byte) Command {
	if len(args) < 2 {
		return Err("ERR invalid args")
	}
	cmd := Get{
		Key: string(args[1]),
	}
	return cmd
}

func (c Get) Handle(ctx *Context) Reply {
	return api.BulkString("key: " + c.Key)
}
