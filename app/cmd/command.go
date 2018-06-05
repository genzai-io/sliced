package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

var (
	ok = redcon.AppendOK(nil)
)

// Aliases
type Context = api.Context
type Command = api.Command
type Reply = api.CommandReply
type Err = api.Err
type String = api.String
type Bytes = api.Bytes

func Error(err error) Err {
	return Err("ERR " + err.Error())
}

var Ok Reply = api.Ok{}

//
//
//
func RAW(b []byte) Command {
	return Bytes(b)
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
