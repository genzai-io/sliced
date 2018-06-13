package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
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
	b = resp.AppendArray(b, 2)
	b = resp.AppendBulkString(b, "GET")
	b = resp.AppendBulkString(b, c.Key)
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
	return api.BulkString("" + c.Key)
}
