package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/common/raft"
	"github.com/genzai-io/sliced/common/service"
)

func newCluster(schema *Store) *ClusterService {
	ctx, cancel := context.WithCancel(context.Background())

	c := &ClusterService{
		ctx:        ctx,
		cancel:     cancel,
		schema:     schema,
		observerCh: make(chan raft.Observation, 1),
	}
	c.BaseService = *service.NewBaseService(moved.Logger, "cluster-raft", c)
	return c
}

// Highest level object that represents the Cluster in it's entirety.
// A Raft group is used to tie every node together with the master of
// that group responsible for managing the adding and removal of system
// objects such as Streams, Tables, etc. It also manages migration related
// tasks. Every node has the system dictionary replicated to it which maintains
// a data store of every system object in the system and who owns what.
type ClusterService struct {
	service.BaseService

	Path  string
	Nodes []*Node

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	schema *Store

	// Raft
	raft      *raft.Raft
	snapshots raft.SnapshotStore
	transport *RaftTransport
	//transport  *raft.NetworkTransport
	store      IRaftStore
	observerCh chan raft.Observation
	observer   *raft.Observer
}

func (s *ClusterService) OnStart() error {
	s.Path = moved.ClusterDir

	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = moved.ClusterID
	raftLogger := &raftLoggerWriter{
		logger: s.Logger,
	}
	config.Logger = log.New(raftLogger, "", 0)
	var err error

	//if s.Path == ":memory:" {
	//	addr, t := raft.NewInmemTransport("")
	//	s.trans = t
	//	_ = addr
	//	s.Logger.Info().Msgf("Transport: %s", addr)
	//} else {

	// Setup Raft communication.
	//addr, err := net.ResolveTCPAddr("tcp", moved.RaftHost)
	//if err != nil {
	//	return err
	//}

	//s.transport, err = raft.NewTCPTransport(string(config.LocalID), addr, 3, time.Second*10, raftLogger)
	//raft.NewDiscardSnapshotStore()
	s.transport = newRaftTransport(-1, -1, "cluster-transport")

	//err = s.transport.Start()
	if err != nil {
		s.Logger.Error().Err(err).Msg("raft transport start failed")
		return err
	}
	//}

	// Create Snapshot Store
	if s.Path == ":memory:" {
		s.snapshots = raft.NewInmemSnapshotStore()
	} else {
		s.snapshots = raft.NewInmemSnapshotStore()

		// Create the snapshot store. This allows the Raft to truncate the log.
		//s.snapshots, err = raft.NewFileSnapshotStore(s.Path, retainSnapshotCount, raftLogger)
		//if err != nil {
		//	s.transport.Stop()
		//	s.Logger.Error().Err(err).Msg("file snapshot store failed")
		//	return fmt.Errorf("file snapshot store: %s", err)
		//}
	}

	// Create Log Store
	var logpath string
	if s.Path == ":memory:" {
		logpath = ":memory:"
	} else {
		if err := os.MkdirAll(s.Path, moved.PathMode); err != nil {
			if err != os.ErrExist {
				s.Logger.Error().AnErr("err", err).Msgf("MkdirAll(\"%s\") error", s.Path)
			}
		}
		logpath = filepath.Join(s.Path, "raft.log")
	}

	// Create the log store and stable store.
	s.store, err = NewLogStore(
		logpath,
		Low,
		s.Logger.With().Str("logger", "cluster.store").Logger(),
	)
	if err != nil {
		//s.transport.Stop()
		s.Logger.Error().Err(err).Msg("log store failed")
		return fmt.Errorf("new log store: %s", err)
	}

	//s.set = item.NewSortedSet()
	//bootstrap := s.set.Length() == 0
	//if bootstrap {
	//	//config.StartAsLeader = true
	//}

	// Instantiate the Raft systems.
	s.raft, err = raft.NewRaft(config, (*clusterFSM)(s), s.store, s.store, s.snapshots, s.transport)
	if err != nil {
		//s.transport.Stop()
		s.store.Close()
		s.Logger.Error().Err(err).Msg("new raft failed")
		return fmt.Errorf("new raft: %s", err)
	}

	// Start up the Raft observer
	s.observer = raft.NewObserver(s.observerCh, false, func(o *raft.Observation) bool {
		return true
	})
	go s.runObserver()
	s.raft.RegisterObserver(s.observer)

	// Bootstrap?
	if moved.Bootstrap {
		//config.StartAsLeader = false
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID: config.LocalID,
					//Address: s.transport.LocalAddr(),
					Address: moved.ClusterAddress,
				},
			},
		}
		s.raft.BootstrapCluster(configuration)
	}

	return nil
}

func (s *ClusterService) runObserver() {
	for {
		select {
		case <-s.ctx.Done():
			return

		case observation, ok := <-s.observerCh:
			if !ok {
				return
			}

			switch val := observation.Data.(type) {
			case raft.RaftState:
				s.Logger.Debug().Msgf("state changed to %s", val)

			case *raft.RequestVoteRequest:
				s.Logger.Debug().Msgf("vote request %s", val)
			}
		}
	}
}

func (c *ClusterService) OnStop() {
	c.cancel()

	if err := c.raft.Shutdown(); err.Error() != nil {
		c.Logger.Error().AnErr("err", err.Error()).Msg("raft.Shutdown() error")
	}

	if err := c.transport.Close(); err != nil {
		c.Logger.Error().AnErr("err", err).Msg("transport.Close() error")
	}

	//if err := c.transport.Stop(); err != nil {
	//	c.Logger.Error().Err(err)
	//}
}

func (c *ClusterService) GetMaster() *Node {
	return nil
}
