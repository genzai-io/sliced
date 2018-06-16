package slice

import (
	"fmt"
	"sync"

	//"github.com/coreos/bbolt"
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/raft"
	"github.com/genzai-io/sliced/common/btrdb"
	"github.com/genzai-io/sliced/common/service"
)

// Each slice has it's own independent store
type Service struct {
	service.BaseService

	ID   uint16
	Path string

	db *btrdb.DB

	topics       map[int64]*TopicSlice
	topicsByName map[string]*TopicSlice

	raftLock  sync.Mutex
	raft      *raft_service.Service
	raftStore *raft_service.LogStore
}

func newService(id api.RaftID, path string) *Service {
	s := &Service{
		Path: path,
	}

	s.BaseService = *service.NewBaseService(moved.Logger, fmt.Sprintf("slice.%d", id), s)

	return s
}

func (b *Service) OnStart() error {
	var err error
	b.db, err = btrdb.OpenWithConfig(b.Path, btrdb.Config{
		SyncPolicy:           btrdb.EverySecond,
		AutoShrinkPercentage: 100,              // 100%
		AutoShrinkMinSize:    32 * 1024 * 1024, // 32MiB
		OnExpired:            b.onExpired,
		OnExpiredSync:        b.onExpiredSync,
	})
	if err != nil {
		b.Logger.Fatal().Err(err)
		return err
	}

	err = b.db.Update(func(tx *btrdb.Tx) error {

		return nil
	})

	return nil
}

func (b *Service) OnStop() {
	if err := b.db.Close(); err != nil {
		b.Logger.Error().AnErr("err", err).Msg("db.Close() error OnStop()")
	}
}

func (b *Service) Backup() {

}

func (g *Service) onExpired(keys []string) {

}

func (g *Service) onExpiredSync(key, value string, tx *btrdb.Tx) error {
	return nil
}

//func (s *Service) loadSlices() error {
//	tx, err := s.db.Begin(false)
//	if err != nil {
//		return err
//	}
//	defer tx.Rollback()
//
//	bucket := tx.Bucket(bucketSlices)
//	if bucket != nil {
//		return ErrBucketNotFound
//	}
//
//	//s.appenders = make([]*AppenderSlice, 0)
//	bucket.ForEach(func(k, v []byte) error {
//		sliceBucket := bucket.Bucket(k)
//
//		if sliceBucket == nil {
//			return nil
//		}
//
//		sliceBucket.ForEach(func(k, v []byte) error {
//			// Read excerpts
//			return nil
//		})
//
//		//model := &store_pb.AppenderSlice{}
//		//if err := model.Unmarshal(v); err != nil {
//		//	return err
//		//}
//		//
//		//appender := &AppenderSlice{
//		//	slice: s,
//		//	model: model,
//		//	dirty: false,
//		//}
//		//
//		//s.appenders = append(s.appenders, appender)
//
//		return nil
//	})
//
//	return nil
//}
