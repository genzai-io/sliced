package cmd

import (
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(&Sleep{}) }

type Sleep struct{}

func (c *Sleep) Name() string   { return "SLEEP" }
func (c *Sleep) Help() string   { return "" }
func (c *Sleep) IsError() bool  { return false }
func (c *Sleep) IsWorker() bool { return true }

func (c *Sleep) Marshal(buf []byte) []byte {
	buf = redcon.AppendArray(buf, 1)
	buf = redcon.AppendBulkString(buf, c.Name())
	return buf
}

func (c *Sleep) Parse(args [][]byte) Command {
	return &Sleep{}
}

func (c *Sleep) Handle(ctx *Context) Reply {
	time.Sleep(time.Second)
	return Ok
}
