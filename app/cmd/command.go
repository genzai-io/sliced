package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

var (
	ok = resp.AppendOK(nil)
)

// Aliases
type Context = api.Context
type Command = api.Command
type Reply = api.CommandReply
type Err = api.Err
type String = api.SimpleString
type Bulk = api.Bulk

var Register = api.Register

func Error(err error) Err {
	return Err("ERR " + err.Error())
}

var Ok Reply = api.Ok{}

//
//
//
func RAW(b []byte) Command {
	return Bulk(b)
}

//
func OK() api.Command {
	return RAW(ok)
}

func Int(value int) api.Command {
	return api.Int(value)
}

func BulkString(str string) api.Command {
	return api.BulkString(str)
}

//
//
//
func ERR(message string) api.Command {
	return RAW(resp.AppendError(nil, message))
}

//
//
//
func ERROR(err error) api.Command {
	return RAW(resp.AppendError(nil, err.Error()))
}
