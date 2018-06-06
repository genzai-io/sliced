package api

import (
	"sync/atomic"
	"runtime"
	"github.com/genzai-io/sliced/common/evio"
)

//
type Durability int

const (
	Low    Durability = -1
	Medium Durability = 0
	High   Durability = 1
)

type ContextPurpose byte

const (
	Parse ContextPurpose = 0
	Apply ContextPurpose = 1
	Log   ContextPurpose = 2
)

type CmdContext interface {
	Conn() CommandConn

	Lock()

	TryLock()

	Unlock()
}

type MultiContext interface {
	CmdContext

	Commands() []CmdContext
}

// Represents a series of pipelined Commands in the context
// of a single call for a single connection or for Raft log applying.
type Context struct {
	Ev     evio.Conn // Connection
	Action evio.Action
	Parse  func(packet []byte, args [][]byte) Command
	Kind   ConnKind
	Raft   RaftService // Raft service for Raft connections

	lock uintptr
}

// Spin-lock
// Only the properties are synchronized and not the command Handle() itself.
// In addition, the Event Loop is inherently single-threaded so the only
// potential race is from a background Worker happening in parallel with
// an Event Loop call.
func (c *Context) Lock() {
	for !atomic.CompareAndSwapUintptr(&c.lock, 0, 1) {
		runtime.Gosched()
	}
}

// Spin-lock TryLock
func (c *Context) TryLock() bool {
	return atomic.CompareAndSwapUintptr(&c.lock, 0, 1)
}

// Spin-lock Unlock
func (c *Context) Unlock() {
	atomic.StoreUintptr(&c.lock, 0)
}
