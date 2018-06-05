package api

import "github.com/genzai-io/sliced/common/raft"

var ApplyCommands map[string]ApplyCommand

// Transaction is a single log entry in the Raft log.
type Transaction struct {
	Requests []ApplyCommand

	log *raft.Log
}


type ApplyCommand interface {
	Marshal(b []byte) []byte

	Unmarshal(b []byte) error

	Handle() CommandReply
}
