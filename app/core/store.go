package core

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/btrdb"
	"github.com/slice-d/genzai/common/service"
	"github.com/slice-d/genzai/proto/store"
)

// Database of all Database meta-data and schema membership and cluster membership services.
// This contains the state of all clusters.
type Store struct {
	service.BaseService

	Path string
	mu   sync.RWMutex
	db   *btrdb.DB

	bootstrap bool
	// Keep track of all nodes
	node  *Node
	nodes map[string]*Node

	raftLock  sync.Mutex
	raft      *RaftService
	raftStore *LogStore


	databases       map[int32]*Database
	databasesByName map[string]*Database
}

func newSchema() *Store {
	s := &Store{}
	s.BaseService = *service.NewBaseService(moved.Logger, "store", s)
	return s
}

func (s *Store) OnStart() error {
	s.Path = moved.SchemaDir

	var name string
	if s.Path != ":memory:" {
		if err := os.MkdirAll(s.Path, moved.PathMode); err != nil {
			if err != os.ErrExist {
				s.Logger.Error().AnErr("err", err).Msgf("MkdirAll(\"%s\") error", s.Path)
			}
		}

		name = filepath.Join(s.Path, "store.db")
	} else {
		name = s.Path
	}

	var err error
	s.db, err = btrdb.OpenWithConfig(name, btrdb.Config{
		SyncPolicy:           btrdb.EverySecond,
		AutoShrinkPercentage: 100,              // 100%
		AutoShrinkMinSize:    32 * 1024 * 1024, // 32MiB
		OnExpired:            s.onExpired,
		OnExpiredSync:        s.onExpiredSync,
	})
	if err != nil {
		s.Logger.Fatal().Err(err)
		return err
	}

	//err = s.initDB()
	//if err != nil {
	//	return err
	//}

	return nil
}

func (s *Store) OnStop() {
	if err := s.db.Close(); err != nil {
		s.Logger.Error().AnErr("err", err).Msg("db.Close() error")
	}
}

func (s *Store) onExpired(keys []string) {

}

func (s *Store) onExpiredSync(key, value string, tx *btrdb.Tx) error {
	return nil
}

func (s *Store) createLocalNode() *store.Node {
	return &store.Node{
		Id: "",
	}
}

func (s *Store) corruptedNodes(tx *btrdb.Tx) {

}

//func (s *Store) updateLocalNode(tx *btrdb.Tx) error {
//	// Serialize
//	var err error
//	var b []byte
//	b, err = json.Marshal(s.node.model)
//	if err != nil {
//		return err
//	}
//
//	// Set in DB
//	_, _, err = tx.Set(schema_nodes_local_key, string(b), nil)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (s *Store) selectSlices(tx *btrdb.Tx) ([]*store.Slice, error) {
//	return nil, nil
//}
//
//func (s *Store) initDB() error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	// Create indexes
//	if err := s.db.Update(func(tx *btrdb.Tx) error {
//
//		return nil
//	}); err != nil {
//		return err
//	}
//
//	return s.db.Update(func(tx *btrdb.Tx) error {
//		insertLocalNode := func() error {
//			s.node = newLocalNode()
//
//			return s.updateLocalNode(tx)
//		}
//
//		// create node
//		if err := func() error {
//			val, err := tx.Get(schema_nodes_local_key)
//			if err != nil {
//				if err == btrdb.ErrNotFound {
//					s.bootstrap = true
//
//					s.Logger.Warn().Msg("local node not in db")
//					s.Logger.Warn().Msg("adding local node to db...")
//
//					err = insertLocalNode()
//					if err != nil {
//						return err
//					}
//				} else {
//					return err
//				}
//			} else {
//				// Unmarshal value
//				var model = store.Node{}
//				err = json.Unmarshal([]byte(val), &model)
//				if err != nil {
//					s.Logger.Error().
//						Str("val", val).
//						AnErr("err", err).
//						Msg("failed to unmarshal local 'Node' from db")
//					s.Logger.Warn().Msg("creating new local node")
//					return insertLocalNode()
//				} else {
//					// Set local node
//					s.node = newNode(&model, true)
//					s.node.syncModel()
//					s.updateLocalNode(tx)
//					return nil
//				}
//			}
//			return nil
//		}(); err != nil {
//			return err
//		}
//
//		// Load all nodes
//		if err := func() error {
//			if s.nodes == nil {
//				s.nodes = make(map[string]*Node)
//			}
//
//			var err error
//			tx.Ascend(schema_nodes_index_name, func(key, value string) bool {
//				model := &store.Node{}
//				err = model.Unmarshal([]byte(value))
//				if err != nil {
//					return false
//				}
//
//				// Node
//				node := newNode(model, s.node.model.Id == model.Id)
//				s.nodes[model.Id] = node
//
//				return true
//			})
//
//			if err != nil {
//				s.corruptedNodes(tx)
//			}
//
//			return err
//		}(); err != nil {
//			return err
//		}
//
//		// Load all slices
//		if err := func() error {
//			var err error
//			tx.Ascend(schema_nodes_index_name, func(key, value string) bool {
//				model := &store.Slice{}
//				err = model.Unmarshal([]byte(value))
//				if err != nil {
//					return false
//				}
//
//				//&Slice{
//				//	model: model,
//				//	owned: s.node.model.Id == model.Id
//				//}
//				//
//				//// Node
//				//node := newS(model, s.node.model.Id == model.Id)
//				//s.nodes[model.Id] = node
//
//				return true
//			})
//
//			if err != nil {
//				s.corruptedNodes(tx)
//			}
//
//			return err
//		}(); err != nil {
//			return err
//		}
//
//		return nil
//	})
//}
//
//func (s *Store) selectTopics(tx *btrdb.Tx) {
//	tx.Ascend(schema_topics_index_name, func(key, value string) bool {
//		return true
//	})
//}