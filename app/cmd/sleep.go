package cmd

import (
	"strconv"
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func init() { api.Register(&Sleep{}) }

type Sleep struct {
	Millis int64
}

func (c Sleep) Name() string   { return "SLEEP" }
func (c Sleep) Help() string   { return "" }
func (c Sleep) IsError() bool  { return false }
func (c Sleep) IsWorker() bool { return true }

func (c Sleep) Marshal(buf []byte) []byte {
	buf = resp.AppendArray(buf, 2)
	buf = resp.AppendBulkString(buf, c.Name())
	buf = resp.AppendBulkInt64(buf, c.Millis)
	return buf
}

func (c Sleep) Parse(args [][]byte) Command {
	if len(args) > 1 {
		cmd := &Sleep{}
		cmd.Millis, _ = strconv.ParseInt(string(args[1]), 10, 64)
		return cmd
	} else {
		return &Sleep{}
	}
}

func (c Sleep) Handle(ctx *Context) Reply {
	if c.Millis < 0 {
		time.Sleep(time.Millisecond)
	} else if c.Millis == 0 {
		//time.Sleep(time.Second)
		time.Sleep(time.Millisecond * 100)
	} else {
		time.Sleep(time.Millisecond * time.Duration(c.Millis))
	}
	return Ok
}
