package core

import "github.com/genzai-io/sliced/common/raft"

//
type SliceStore struct {
	raft.StableStore
}
