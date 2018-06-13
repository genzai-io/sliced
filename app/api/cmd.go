package api

import (
	"strings"
	"unsafe"
)

// Global registry of commands
var Commands map[string]Command

func init() {
	Commands = make(map[string]Command)
}

type CommandStats struct {
	Name string
}

var GlobalRaftID = RaftID{
	DatabaseID: -1,
	SliceID:    -1,
}

type RaftID struct {
	DatabaseID int32
	SliceID    int32
}

//
//
//
type Command interface {
	Name() string

	Help() string

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
func ParseCommand(packet []byte, args [][]byte) Command {
	if len(args) == 0 {
		return Err("ERR empty request")
	}

	name := *(*string)(unsafe.Pointer(&args[0]))

	// Find command
	command, ok := Commands[name]
	if !ok {
		return nil
	}
	if command != nil {
		command = command.Parse(args)
	}

	return command
}

func Register(cmd Command) {
	name := strings.TrimSpace(cmd.Name())
	lower := strings.ToLower(name)
	upper := strings.ToUpper(name)

	if _, ok := Commands[lower]; ok {
		panic("command name '" + lower + "' already used")
	}
	if _, ok := Commands[upper]; ok {
		panic("command name '" + upper + "' already used")
	}

	Commands[lower] = cmd
	Commands[upper] = cmd
}
