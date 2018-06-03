package core

import (
	"sync"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/ring"
	"github.com/slice-d/genzai/common/service"
	store_pb "github.com/slice-d/genzai/proto/store"
)

// Database's are isolated from each other and have their own Slices and
// Slice Raft groups. Manages partitioning and migration within the
// context of a Database.
type Database struct {
	service.BaseService

	mu    sync.RWMutex
	model store_pb.Database
	store *Store

	id     int32
	ring   *ring.Ring
	slices []*Slice

	rollers map[string]*Roller
	topics  map[string]*Topic
	queues  map[string]*Queue

	topicCounter  uint64
	rollerCounter uint64
	queueCounter  uint64
	tableCounter  uint64
}

func newDatabase(store *Store) *Database {
	d := &Database{
		store:   store,
		rollers: make(map[string]*Roller),
		topics:  make(map[string]*Topic),
		queues:  make(map[string]*Queue),
	}
	d.BaseService = *service.NewBaseService(moved.Logger, "database", d)
	return d
}

func (d *Database) OnStart() error {
	return nil
}

func (d *Database) OnStop() {

}

func (s *Database) partition() {

}

type DatabaseCluster struct {
	ring   *ring.Ring
	slices []*Slice
}
