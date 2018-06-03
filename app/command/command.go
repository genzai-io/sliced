package command

import (
	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/common/redcon"
)

var (
	ok = redcon.AppendOK(nil)
)

// Aliases
type Context = api.Context
type Command = api.Command

//
//
//
func RAW(b []byte) api.Command {
	return api.RawCmd(b)
}

//
func OK() api.Command {
	return RAW(ok)
}

func Int(value int) api.Command {
	return RAW(redcon.AppendInt(nil, int64(value)))
}

func Bulk(b []byte) api.Command {
	return RAW(redcon.AppendBulk(nil, b))
}

func BulkString(str string) api.Command {
	return RAW(redcon.AppendBulkString(nil, str))
}

//
//
//
func ERR(message string) api.Command {
	return RAW(redcon.AppendError(nil, message))
}

//
//
//
func ERROR(err error) api.Command {
	return RAW(redcon.AppendError(nil, err.Error()))
}