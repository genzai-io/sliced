package core

import (
	"fmt"
	"sync"

	//"github.com/coreos/bbolt"
	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/btrdb"
	"github.com/slice-d/genzai/common/service"
)

// Each slice has it's own independent store
type SliceService struct {
	service.BaseService

	ID   uint16
	Path string

	db     *btrdb.DB

	topics       map[int64]*TopicSlice
	topicsByName map[string]*TopicSlice

	raftLock  sync.Mutex
	raft      *RaftService
	raftStore *LogStore
}

func newSliceService(id api.RaftID, path string) *SliceService {
	s := &SliceService{
		Path:   path,
	}

	s.BaseService = *service.NewBaseService(moved.Logger, fmt.Sprintf("slice.%d", id), s)

	return s
}

func (b *SliceService) OnStart() error {
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

func (b *SliceService) OnStop() {
	if err := b.db.Close(); err != nil {
		b.Logger.Error().AnErr("err", err).Msg("db.Close() error OnStop()")
	}
}

func (b *SliceService) Backup() {

}

func (g *SliceService) onExpired(keys []string) {

}

func (g *SliceService) onExpiredSync(key, value string, tx *btrdb.Tx) error {
	return nil
}

//func (s *SliceService) loadSlices() error {
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
