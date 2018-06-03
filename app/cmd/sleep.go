package cmd

import (
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() {
	api.Register("SLEEP", &Sleep{})
}

type Sleep struct {
	Command
}

func (c *Sleep) IsChange() bool { return false }
func (c *Sleep) IsAsync() bool  { return true }

func (c *Sleep) Marshal(buf []byte) []byte {
	buf = redcon.AppendArray(buf, 1)
	buf = redcon.AppendBulkString(buf, "SLEEP")
	return buf
}

func (c *Sleep) Parse(ctx *Context) api.Command {
	return &Sleep{}
}

func (c *Sleep) Handle(ctx *Context) {
	time.Sleep(time.Second)
	ctx.OK()
}

func (c *Sleep) Apply(ctx *Context) {}
