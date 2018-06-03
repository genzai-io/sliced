package core

import (
	"io"

	"github.com/slice-d/genzai/common/raft"
)

type clusterFSM ClusterService

// Apply applies a Raft log entry to the key-value store.
func (f *clusterFSM) Apply(l *raft.Log) interface{} {
	return nil
}

// Snapshot returns a snapshot of the key-value store.
func (f *clusterFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return &clusterFSMSnapshot{}, nil
}

// Restore stores the key-value store to a previous state.
func (f *clusterFSM) Restore(rc io.ReadCloser) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	//f.db = db

	return nil
}

type clusterFSMSnapshot struct {
	store *Store
}

func (f *clusterFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		err := f.store.db.Save(sink)
		if err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *clusterFSMSnapshot) Release() {}
