package core

import (
	"sync"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/ring"
	"github.com/genzai-io/sliced/common/service"
	store_pb "github.com/genzai-io/sliced/proto/store"
	"github.com/genzai-io/sliced/btrdb"
)

// Database's are isolated from each other and have their own Slices and
// Slice Raft groups. Manages partitioning and migration within the
// context of a Database.
type Database struct {
	service.BaseService

	mu    sync.RWMutex
	model store_pb.Database
	db *btrdb.DB

	id     int32
	ring   *ring.Ring
	slices []*Slice

	tblTopics    *btrdb.Table
	topics map[int64]*Topic
	queues map[int64]*Queue

	topicCounter  uint64
	rollerCounter uint64
	queueCounter  uint64
	tableCounter  uint64
}

func newDatabase(db *btrdb.DB, model *store_pb.Database) *Database {
	d := &Database{
		db: db,
		model:  *model,
		topics: make(map[int64]*Topic),
		queues: make(map[int64]*Queue),
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
