package api

import (
	"strings"

	"github.com/slice-d/genzai/common/redcon"
)

var (
	ok = redcon.AppendOK(nil)
)

var Commands map[string]Command

func init() {
	Commands = make(map[string]Command)
}

func Register(name string, cmd Command) {
	Commands[name] = cmd
	Commands[strings.ToLower(name)] = cmd
	Commands[strings.ToUpper(name)] = cmd
}

type CommandStats struct {
	Name string
}

var GlobalRaftID = RaftID{
	Schema: -1,
	Slice:  -1,
}

type RaftID struct {
	Schema int32
	Slice  int32
}

//
//
//
type Command interface {
	Marshal([]byte) []byte

	// Requires Raft applier
	IsChange() bool

	// Flag to determine whether the command and be handled inline or
	// if it requires a worker.
	IsAsync() bool

	// Parses from a Redcon connection
	Parse(ctx *Context) Command

	// Invoke happens on the EvLoop
	Handle(ctx *Context)

	// Applies the command to the FSM
	Apply(ctx *Context)
}

//
//
//
type Cmd struct {
	Command
}

func (c *Cmd) Name() []byte {
	return nil
}

func (c *Cmd) Marshal(buf []byte) []byte {
	return buf
}

func (c *Cmd) Parse(ctx *Context) Command {
	return ERR("not implemented")
}

func (c *Cmd) IsChange() bool {
	return false
}

func (c *Cmd) IsAsync() bool {
	return false
}

func (c *Cmd) Handle(ctx *Context) {
	ctx.Err("not implemented")
}

func (c *Cmd) Apply(ctx *Context) {

}

//
//
//
func RAW(b []byte) Command {
	return RawCmd(b)
}

type RawCmd []byte

func (c RawCmd) Marshal(buf []byte) []byte {
	buf = append(buf, c...)
	return buf
}

func (c RawCmd) Parse(ctx *Context) Command {
	return ERR("not implemented")
}

func (c RawCmd) IsChange() bool {
	return false
}

func (c RawCmd) IsAsync() bool {
	return false
}

func (c RawCmd) Handle(ctx *Context) {
	if c == nil {
		return
	}
	ctx.Out = append(ctx.Out, c...)
}

func (c RawCmd) Apply(ctx *Context) {}

//
func OK() Command {
	return RAW(ok)
}

func Int(value int) Command {
	return RAW(redcon.AppendInt(nil, int64(value)))
}

//
//
//
func ERR(message string) Command {
	return RAW(redcon.AppendError(nil, message))
}

//
//
//
func ERROR(err error) Command {
	return RAW(redcon.AppendError(nil, err.Error()))
}

// ERR command
type ErrCmd struct {
	Command
	Result []byte
}

func (c *ErrCmd) Invoke(out []byte) []byte {
	return append(out, c.Result...)
}

//
var WRITE = &WriteCmd{}

//
//
//
type WriteCmd struct {
	Command
}

func (c *WriteCmd) Parse(ctx *Context) Command {
	return ERR("not implemented")
}

func (c *WriteCmd) IsChange() bool {
	return true
}

func (c *WriteCmd) IsAsync() bool {
	return false
}

func (c *WriteCmd) Apply(ctx *Context) []byte {
	return nil
}

//
// Command needs to be dispatched and ran on a Worker
//
type BgCmd struct {
	Command
}

func (c *BgCmd) IsAsync() bool {
	return true
}

func (c *BgCmd) Handle(ctx *Context) {
	ctx.Err("not implemented")
}
