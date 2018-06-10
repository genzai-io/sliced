package api

import (
	"github.com/genzai-io/sliced/common/redcon"
	"io"
	"bytes"
	"bufio"
	"fmt"
)

var (
	OK     = Ok{}
	PONG   = Pong{}
	QUEUED = Queued{}
	NIL    = Nil{}
)

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

func (o Ok) IsMatch(command CommandReply) bool {
	if o == command {
		return true
	}
	str, ok := command.(String)
	if ok {
		return str == "OK" || str == "ok"
	}
	return false
}

//
//
//
type Queued struct{}

func (c Queued) Name() string   { return "Queued" }
func (c Queued) Help() string   { return "" }
func (c Queued) IsError() bool  { return false }
func (c Queued) IsWorker() bool { return true }

func (e Queued) Marshal(b []byte) []byte {
	return redcon.AppendQueued(b)
}

func (e Queued) Parse(args [][]byte) Command {
	return e
}

func (e Queued) MarshalReply(b []byte) []byte {
	return redcon.AppendQueued(b)
}

func (e Queued) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (e Queued) Handle(ctx *Context) CommandReply {
	return e
}

//
//
//
type Pong struct{}

func (p Pong) Name() string   { return "PONG" }
func (p Pong) Help() string   { return "" }
func (p Pong) IsError() bool  { return false }
func (p Pong) IsWorker() bool { return true }

func (p Pong) Marshal(b []byte) []byte {
	return redcon.AppendString(b, p.Name())
}

func (p Pong) Parse(args [][]byte) Command {
	return p
}

func (p Pong) MarshalReply(b []byte) []byte {
	return redcon.AppendString(b, p.Name())
}

func (p Pong) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (p Pong) Handle(ctx *Context) CommandReply {
	return p
}

//
//
//
type Err string

func (e Err) Name() string   { return "Err" }
func (e Err) Help() string   { return "" }
func (e Err) IsError() bool  { return true }
func (e Err) IsWorker() bool { return false }

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
type BulkString string

func (s BulkString) Name() string   { return "BulkString" }
func (s BulkString) Help() string   { return "" }
func (s BulkString) IsError() bool  { return false }
func (s BulkString) IsWorker() bool { return false }

func (s BulkString) Marshal(b []byte) []byte {
	return redcon.AppendBulkString(b, string(s))
}

func (s BulkString) Parse(args [][]byte) Command {
	return s
}

func (s BulkString) MarshalReply(b []byte) []byte {
	return redcon.AppendBulkString(b, string(s))
}

func (s BulkString) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (s BulkString) Handle(ctx *Context) CommandReply {
	return s
}

//
//
//
type String string

func (s String) Name() string   { return "String" }
func (s String) Help() string   { return "" }
func (s String) IsError() bool  { return false }
func (s String) IsWorker() bool { return false }

func (s String) Marshal(b []byte) []byte {
	return redcon.AppendString(b, string(s))
}

func (s String) Parse(args [][]byte) Command {
	return s
}

func (s String) MarshalReply(b []byte) []byte {
	return redcon.AppendString(b, string(s))
}

func (s String) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (s String) Handle(ctx *Context) CommandReply {
	return s
}

//
//
//
type Bytes []byte

func (by Bytes) Name() string   { return "Bytes" }
func (by Bytes) Help() string   { return "" }
func (by Bytes) IsError() bool  { return false }
func (by Bytes) IsWorker() bool { return false }

func (by Bytes) Marshal(b []byte) []byte {
	return redcon.AppendBulk(b, []byte(by))
}

func (by Bytes) Parse(args [][]byte) Command {
	return by
}

func (by Bytes) MarshalReply(b []byte) []byte {
	return redcon.AppendBulk(b, []byte(by))
}

func (by Bytes) UnmarshalReply(packet []byte, args [][]byte) error {
	by = args[0]
	return nil
}

func (by Bytes) Handle(ctx *Context) CommandReply {
	return by
}

//
//
//
type Array []CommandReply

func (arr Array) Name() string   { return "Array" }
func (arr Array) Help() string   { return "" }
func (arr Array) IsError() bool  { return false }
func (arr Array) IsWorker() bool { return false }

func (arr Array) Marshal(b []byte) []byte {
	if len(arr) == 0 {
		return redcon.AppendArray(b, 0)
	} else {
		b = redcon.AppendArray(b, len(arr))
		for _, element := range arr {
			b = element.MarshalReply(b)
		}
	}
	return b
}

func (arr Array) Parse(args [][]byte) Command {
	return arr
}

func (arr Array) MarshalReply(b []byte) []byte {
	if len(arr) == 0 {
		return redcon.AppendArray(b, 0)
	} else {
		b = redcon.AppendArray(b, len(arr))
		for _, element := range arr {
			b = element.MarshalReply(b)
		}
	}
	return b
}

func (arr Array) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (arr Array) Handle(ctx *Context) CommandReply {
	return arr
}

//
//
//
type Nil struct{}

func (n Nil) Name() string   { return "Nil" }
func (n Nil) Help() string   { return "" }
func (n Nil) IsError() bool  { return false }
func (n Nil) IsWorker() bool { return true }

func (n Nil) Marshal(b []byte) []byte {
	return redcon.AppendNull(b)
}

func (n Nil) Parse(args [][]byte) Command {
	return n
}

func (n Nil) MarshalReply(b []byte) []byte {
	return redcon.AppendNull(b)
}

func (n Nil) UnmarshalReply(packet []byte, args [][]byte) error {
	return nil
}

func (n Nil) Handle(ctx *Context) CommandReply {
	return n
}

func ReplyType(reply CommandReply) string {
	switch reply.(type) {
	case String:
		return "String"

	case BulkString:
		return "BulkString"

	case Bytes:
		return "Bytes"

	case Int:
		return "Int"

	case Ok:
		return "Ok"

	case Nil:
		return "Nil"

	case Queued:
		return "Queued"

	case Pong:
		return "Pong"

	case Array:
		return "Array"
	}
	return "Unknown"
}

func ReplyEquals(reply CommandReply, reply2 CommandReply) bool {
	switch rt := reply.(type) {
	case String:
		switch r2t := reply2.(type) {
		case Bytes:
			return string(rt) == string(r2t)
		case String:
			return string(rt) == string(r2t)
		case BulkString:
			return string(rt) == string(r2t)
		}
		return false

	case BulkString:
		switch r2t := reply2.(type) {
		case Bytes:
			return string(rt) == string(r2t)
		case String:
			return string(rt) == string(r2t)
		case BulkString:
			return string(rt) == string(r2t)
		}
		return false

	case Bytes:
		switch r2t := reply2.(type) {
		case Bytes:
			return string(rt) == string(r2t)
		case String:
			return string(rt) == string(r2t)
		case BulkString:
			return string(rt) == string(r2t)
		}
		return false

	case Int:
		switch r2t := reply2.(type) {
		case Int:
			return int64(rt) == int64(r2t)
		}
		return false

	case Ok:
		if _, ok := reply2.(Ok); ok {
			return true
		} else {
			return false
		}

	case Nil:
		if _, ok := reply2.(Nil); ok {
			return true
		} else {
			return false
		}

	case Queued:
		if _, ok := reply2.(Queued); ok {
			return true
		} else {
			return false
		}

	case Pong:
		if _, ok := reply2.(Pong); ok {
			return true
		} else {
			return false
		}

	case Array:
		if av, ok := reply2.(Array); ok {
			if len(rt) != len(av) {
				return false
			}
			for index, value := range rt {
				return ReplyEquals(value, av[index])
			}
		} else {
			return false
		}
	}
	return false
}

//
//
//
type ProtocolError string

func (pe ProtocolError) Error() string {
	return fmt.Sprintf("redigo: %s (possible server error or unsupported concurrent read by application)", string(pe))
}

func readLine(br *bufio.Reader) ([]byte, error) {
	p, err := br.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		return nil, ProtocolError("long response line")
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, ProtocolError("bad response line terminator")
	}
	return p[:i], nil
}

// parseLen parses bulk string and array lengths.
func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, ProtocolError("malformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, ProtocolError("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}

// parseInt parses an integer reply.
func parseInt(p []byte) (Int, error) {
	if len(p) == 0 {
		return 0, ProtocolError("malformed integer")
	}

	var negate bool
	if p[0] == '-' {
		negate = true
		p = p[1:]
		if len(p) == 0 {
			return 0, ProtocolError("malformed integer")
		}
	}

	var n int64
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return 0, ProtocolError("illegal bytes in length")
		}
		n += int64(b - '0')
	}

	if negate {
		n = -n
	}
	return Int(n), nil
}

type ReplyReader struct {
	reader *bufio.Reader
	r      *bytes.Reader
}

func NewReplyReader(b []byte) *ReplyReader {
	r := bytes.NewReader(b)
	return &ReplyReader{
		reader: bufio.NewReader(r),
	}
}

func (rr *ReplyReader) Reset(b []byte) {
	rr.r.Reset(b)
	rr.reader.Reset(rr.r)
}

func (rr *ReplyReader) Next() (CommandReply, error) {
	c := rr.reader

	line, err := readLine(c)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, ProtocolError("short response line")
	}
	switch line[0] {
	case '+':
		switch {
		case len(line) == 3 && line[1] == 'O' && line[2] == 'K':
			// Avoid allocation for frequent "+OK" response.
			return OK, nil

		case len(line) == 5 && line[1] == 'P' && line[2] == 'O' && line[3] == 'N' && line[4] == 'G':
			// Avoid allocation in PING command benchmarks :)
			return PONG, nil

		case len(line) == 7 && line[1] == 'Q' && line[2] == 'U' && line[3] == 'E' && line[4] == 'U' && line[5] == 'E' && line[6] == 'D':
			// Avoid allocation for frequent "+QUEUED" response.
			return QUEUED, nil

		default:
			return String(line[1:]), nil
		}
	case '-':
		return Err(string(line[1:])), nil
	case ':':
		return parseInt(line[1:])
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(rr.reader, p)
		if err != nil {
			return nil, err
		}
		if line, _, err := c.ReadLine(); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, ProtocolError("bad bulk string format")
		}
		return Bytes(p), nil
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return NIL, err
		}
		r := make([]CommandReply, n)
		for i := range r {
			r[i], err = rr.Next()
			if err != nil {
				return nil, err
			}
		}
		return Array(r), nil
	}
	return nil, ProtocolError("unexpected response line")
}
