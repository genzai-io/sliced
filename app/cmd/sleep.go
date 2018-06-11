package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
	"time"
)

func init() { api.Register(&Sleep{}) }

type Sleep struct{
	Millis int64
}

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
	if len(args) > 1 {

	}

	return &Sleep{}
}

func (c *Sleep) Handle(ctx *Context) Reply {
	if c.Millis < 0 {

	} else if c.Millis == 0 {
		time.Sleep(time.Second)
	} else {

	}
	return Ok
}
