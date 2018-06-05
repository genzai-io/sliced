package api

import (
	"strings"
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
	IsError() bool

	// Flag to determine whether the command and be handled inline or
	// if it requires a worker.
	IsWorker() bool

	//
	Marshal(b []byte) []byte

	// Parses from a Redcon connection
	Parse(args [][]byte) Command

	// Invoke happens on the EvLoop
	Handle(ctx *Context) CommandReply
}

//
//
//
func RAW(b []byte) Command {
	return Bytes(b)
}
