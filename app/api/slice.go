package api

// A slice is an instance of a Database that does not share state with
// any other slice within the schema or otherwise. It can be thought of
// as an independent database. State is consistently maintained
// through a Raft log. (CA of CAP) system.
//
// Object definitions are inherited from it's Database. A Slice represents
// some range(s) of the total slots (16384) of the Database. Each record
// in a slice is assigned a Slot number usually based on the hash of it's key.
// A Slice's log is serialized but it's reads may happen in parallel.
//
// To scale a Database add another Slice and let the Database re-balance.
type Slice interface {
	// Each slice has it's own Raft cluster
	Raft() RaftService

	Schema() Database

	Handle(ctx *Context, command Command)

	TopicAppend()

	QueueAppend()

	Set()

	Get()

	Iterate()

	Incr()

	Decr()
}
