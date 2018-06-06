package api

import "github.com/genzai-io/sliced/common/redcon"

type CommandReply interface {
	IsError() bool

	//
	MarshalReply(b []byte) []byte

	//
	UnmarshalReply(packet []byte, args [][]byte) error
}

//
//
//
type Ok struct{}

func (c Ok) Name() string   { return "OK" }
func (c Ok) Help() string   { return "" }
func (c Ok) IsError() bool  { return false }
func (c Ok) IsWorker() bool { return true }

func (e Ok) Marshal(b []byte) []byte {
	return redcon.AppendOK(b)
}

func (e Ok) Parse(args [][]byte) Command {
	return e
}

func (e Ok) MarshalReply(b []byte) []byte {
	return redcon.AppendOK(b)
}

func (e Ok) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (e Ok) Handle(ctx *Context) CommandReply {
	return e
}

//
//
//
type Err string

func (c Err) Name() string   { return "Err" }
func (c Err) Help() string   { return "" }
func (c Err) IsError() bool  { return true }
func (c Err) IsWorker() bool { return false }

func (e Err) Error() string {
	return string(e)
}

func (e Err) Marshal(b []byte) []byte {
	return redcon.AppendError(b, string(e))
}

func (e Err) Parse(args [][]byte) Command {
	return e
}

func (e Err) MarshalReply(b []byte) []byte {
	return redcon.AppendError(b, string(e))
}

func (e Err) UnmarshalReply(packet []byte, args [][]byte) error {
	e = Err(string(args[0]))
	return nil
}

func (e Err) Handle(ctx *Context) CommandReply {
	return e
}

//
//
//
type Int int64

func (c Int) Name() string   { return "Int" }
func (c Int) Help() string   { return "" }
func (c Int) IsError() bool  { return false }
func (c Int) IsWorker() bool { return false }

func (c Int) Marshal(b []byte) []byte {
	b = redcon.AppendArray(b, 1)
	return redcon.AppendBulkInt64(b, int64(c))
}

func (c Int) Parse(args [][]byte) Command {
	return c
}

func (c Int) MarshalReply(b []byte) []byte {
	return redcon.AppendInt(b, int64(c))
}

func (c Int) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (c Int) Handle(ctx *Context) CommandReply {
	return c
}

//
//
//
type String string

func (c String) Name() string   { return "String" }
func (c String) Help() string   { return "" }
func (c String) IsError() bool  { return false }
func (c String) IsWorker() bool { return false }

func (c String) Marshal(b []byte) []byte {
	return redcon.AppendBulkString(b, string(c))
}

func (c String) Parse(args [][]byte) Command {
	return c
}

func (c String) MarshalReply(b []byte) []byte {
	return redcon.AppendBulkString(b, string(c))
}

func (c String) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (c String) Handle(ctx *Context) CommandReply {
	return c
}

//
//
//
type Bytes []byte

func (c Bytes) Name() string   { return "Bytes" }
func (c Bytes) Help() string   { return "" }
func (c Bytes) IsError() bool  { return false }
func (c Bytes) IsWorker() bool { return false }

func (c Bytes) Marshal(b []byte) []byte {
	return redcon.AppendBulk(b, []byte(c))
}

func (c Bytes) Parse(args [][]byte) Command {
	return c
}

func (c Bytes) MarshalReply(b []byte) []byte {
	return redcon.AppendBulk(b, []byte(c))
}

func (c Bytes) UnmarshalReply(packet []byte, args [][]byte) error {
	c = args[0]
	return nil
}

func (c Bytes) Handle(ctx *Context) CommandReply {
	return c
}
