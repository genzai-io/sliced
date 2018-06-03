package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/raft"
	"github.com/genzai-io/sliced/common/service"
)

type RaftStatus byte

const (
	RaftStopped RaftStatus = 0
	// First and only node of the cluster that is forced to be Leader.
	// This forces a Quorum for a single node.
	RaftBootstrap RaftStatus = 1
	// Quorum failed and the cluster is down.
	// Either bring another node online or move to Bootstrap mode
	RaftQuorumFailure RaftStatus = 2
	// Quorum failed, however there are new node(s) that are in
	// the staging process when, once completed, will bring it back
	// online.
	RaftQuorumStaging RaftStatus = 3
	// Quorum successfully maintained
	RaftQuorum RaftStatus = 3
)

var (
	ErrNotRunning = errors.New("not running")
)

type RaftService struct {
	service.BaseService

	// Context
	ctx    context.Context
	cancel context.CancelFunc

	muCluster sync.RWMutex

	status     RaftStatus
	wgRoutines sync.WaitGroup

	dir      string
	schemaID int32
	sliceID  int32

	store *LogStore

	// A map of voting members
	voters map[string]*Node
	// Map of non-voting members
	nonvoters map[string]*Node
	// Map of staging members.
	// Staging members are currently in the process
	// of getting caught up with the Leader. Once
	// it's caught up it will promote to either
	// voter or non-voter.
	staging map[string]*Node

	// Observer
	observerCh chan raft.Observation
	observer   *raft.Observer

	// Raft
	fsm          raft.FSM
	fsmSnapshot  raft.FSMSnapshot
	config       *raft.Config
	loggerWriter *raftLoggerWriter
	raft         *raft.Raft
	snapshots    raft.SnapshotStore
	transport    *RaftTransport
}

func newRaftService(dir string, schema int32, slice int32) *RaftService {
	ctx, cancel := context.WithCancel(context.Background())

	c := &RaftService{
		ctx:        ctx,
		cancel:     cancel,
		dir:        dir,
		schemaID:   schema,
		sliceID:    slice,
		status:     RaftStopped,
		voters:     make(map[string]*Node),
		nonvoters:  make(map[string]*Node),
		staging:    make(map[string]*Node),
		observerCh: make(chan raft.Observation),
	}

	c.BaseService = *service.NewBaseService(moved.Logger, "cluster", c)
	return c
}

func (rs *RaftService) getLoggerName() string {
	if rs.schemaID < 0 {
		return "raft-cluster"
	} else {
		return fmt.Sprintf("raft-schema-%d-%d", rs.schemaID, rs.sliceID)
	}
}

func (rs *RaftService) getTransportLoggerName() string {
	if rs.schemaID < 0 {
		return "raft-bus-system"
	} else {
		return fmt.Sprintf("raft-bus-%d-%d", rs.schemaID, rs.sliceID)
	}
}

func (rs *RaftService) OnStart() error {
	rs.loggerWriter = &raftLoggerWriter{
		logger: rs.Logger,
	}

	rs.config = raft.DefaultConfig()
	rs.config.LocalID = moved.ClusterID

	var err error
	var logname string
	if rs.dir == ":memory:" {
		// Create snapshot store
		rs.snapshots = raft.NewInmemSnapshotStore()
		logname = rs.dir
	} else {
		rs.snapshots, err = raft.NewFileSnapshotStore(filepath.Join(rs.dir, "snapshots"), 3, rs.loggerWriter)
		if err != nil {
			rs.Logger.Error().Err(err).Msg("file snapshot store failed")
			return fmt.Errorf("file snapshot store: %s", err)
		}

		logname = filepath.Join(rs.dir, "raft.db")
	}

	// Create the log store and stable store.
	rs.store, err = NewLogStore(
		logname,
		Medium,
		rs.Logger.With().Str("logger", "cluster.store").Logger(),
	)
	if err != nil {
		//s.transport.Stop()
		rs.Logger.Error().Err(err).Msg("log store failed")
		return fmt.Errorf("new log store: %s", err)
	}

	// Start transport
	rs.transport = NewTransport(rs.schemaID, rs.sliceID, rs.getTransportLoggerName())
	err = rs.transport.Start()
	if err != nil {
		rs.Logger.Error().Err(err).Msg("raft transport start failed")
		return err
	}

	// Instantiate the Raft systems.
	rs.raft, err = raft.NewRaft(rs.config, rs.fsm, rs.store, rs.store, rs.snapshots, rs.transport)
	if err != nil {
		rs.transport.Stop()
		rs.store.Close()
		rs.Logger.Error().Err(err).Msg("new raft failed")
		return fmt.Errorf("new raft: %s", err)
	}

	// Start up the Raft observer
	rs.observer = raft.NewObserver(rs.observerCh, false, func(o *raft.Observation) bool {
		return true
	})

	rs.goFunc(rs.observe)
	rs.raft.RegisterObserver(rs.observer)

	return nil
}

func (rs *RaftService) OnStop() {
	rs.muCluster.Lock()
	defer rs.muCluster.Unlock()

	if !rs.IsRunning() {
		return
	}

	future := rs.raft.Shutdown();
	if err := future.Error(); err != nil {
		rs.Logger.Error().AnErr("err", err).Msg("raft.Shutdown() error")
	}
	if err := rs.transport.Stop(); err != nil {
		rs.Logger.Error().AnErr("err", err).Msg("transport.Stop() error")
	}
	if err := rs.store.Close(); err != nil {
		rs.Logger.Error().AnErr("err", err).Msg("store.Stop() error")
	}
}

func (rs *RaftService) bootstrap() error {
	rs.muCluster.Lock()
	defer rs.muCluster.Unlock()

	if !rs.IsRunning() {
		return ErrNotRunning
	}

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      rs.config.LocalID,
				Address: moved.ClusterAddress,
			},
		},
	}
	future := rs.raft.BootstrapCluster(configuration)
	if err := future.Error(); err != nil {
		return err
	}

	rs.status = RaftBootstrap

	return nil
}

func (rs *RaftService) recoverCluster() error {
	rs.muCluster.Lock()
	defer rs.muCluster.Unlock()

	if !rs.IsRunning() {
		return ErrNotRunning
	}

	return nil
}

// Start a goroutine and properly handle the race between a routine
// starting and incrementing, and exiting and decrementing.
func (rs *RaftService) goFunc(f func()) {
	rs.wgRoutines.Add(1)
	go func() {
		defer rs.wgRoutines.Done()
		f()
	}()
}

func (rs *RaftService) observe() {
	for {
		select {
		case <-rs.ctx.Done():
			return

		case <-time.After(time.Second * 5):

		case observation, ok := <-rs.observerCh:
			if !ok {
				return
			}
			str := ""
			data, err := json.Marshal(observation.Data)
			if err != nil {
				str = string(data)
			} else {
				str = fmt.Sprintf("%s", data)
			}
			rs.Logger.Debug().Msgf("raft observation: %s", str)
		}
	}
}

func (rs *RaftService) nodeRemoved(node *Node) {

}

func (rs *RaftService) nodeLost(node *Node) {

}

func (rs *RaftService) nodeStaging(node *Node) {

}

func (rs *RaftService) nodeOnline(node *Node) {

}


// Only for RaftTransport RAFTAPPEND
func (rs *RaftService) Append(o []byte, args [][]byte) ([]byte, error) {
	return rs.transport.handleAppendEntries(o, args)
}

// Only for RaftTransport RAFTVOTE
func (rs *RaftService) Vote(o []byte, args [][]byte) ([]byte, error) {
	return rs.transport.handleRequestVote(o, args)
}

// Only for RaftTransport RAFTINSTALL
func (rs *RaftService) Install(conn *api.Context, arg []byte) api.Command {
	return rs.transport.HandleInstallSnapshot(conn, arg)
}

func (rs *RaftService) IsLeader() bool {
	return rs.raft.Leader() == moved.ClusterAddress
}

func (rs *RaftService) Leader() raft.ServerAddress {
	return rs.raft.Leader()
}

func (rs *RaftService) Stats() map[string]string {
	return rs.raft.Stats()
}

func (rs *RaftService) State() raft.RaftState {
	return rs.raft.State()
}

func (rs *RaftService) Configuration() (raft.ConfigurationFuture, error) {
	future := rs.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return future, err
	}
	return future, nil
}

func (rs *RaftService) Bootstrap() error {
	rs.Logger.Info().Msg("bootstrap request")

	configFuture := rs.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		rs.Logger.Error().AnErr("err", err).Msg("failed to get raft configuration")
		return err
	}

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      moved.ClusterID,
				Address: moved.ClusterAddress,
			},
		},
	}

	future := rs.raft.BootstrapCluster(configuration)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}

func (rs *RaftService) Demote(nodeID string) error {
	rs.Logger.Info().Str("node", nodeID).Str("addr", nodeID).Msg("node demote request")

	serverID := raft.ServerID(nodeID)
	address := raft.ServerAddress(nodeID)

	configFuture := rs.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		rs.Logger.Error().AnErr("err", err).Msg("failed to get raft configuration")
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node'rs ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == address {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == address && srv.ID == serverID && srv.Suffrage != raft.Voter {
				rs.Logger.Warn().Msgf("node %rs at %rs is a %rs and not a Voter, ignoring demote request", nodeID, address, srv.Suffrage)
				return nil
			}

			future := rs.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				rs.Logger.Error().AnErr("err", err).Msgf("error removing existing node %rs at %rs: %rs", nodeID, address, err)
				return err
			}
		}
	}

	future := rs.raft.DemoteVoter(serverID, 0, 0)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (rs *RaftService) Join(nodeID string, voter bool) error {
	rs.Logger.Info().Str("node", nodeID).Str("addr", nodeID).Msg("node join request")

	serverID := raft.ServerID(nodeID)
	address := raft.ServerAddress(nodeID)

	configFuture := rs.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		rs.Logger.Error().AnErr("err", err).Msg("failed to get raft configuration")
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node'rs ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == address {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == address && srv.ID == serverID {
				rs.Logger.Warn().Msgf("node %rs at %rs already member of cluster, ignoring join request", nodeID, address)
				return nil
			}

			future := rs.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				rs.Logger.Error().AnErr("err", err).Msgf("error removing existing node %rs at %rs: %rs", nodeID, address, err)
				return err
			}
		}
	}

	if voter {
		f := rs.raft.AddVoter(serverID, address, 0, 0)
		if err := f.Error(); err != nil {
			rs.Logger.Info().Str("node", nodeID).Str("addr", string(address)).Err(f.Error()).Msg("node join failed")
			return f.Error()
		}
	} else {
		f := rs.raft.AddNonvoter(serverID, address, 0, 0)
		if err := f.Error(); err != nil {
			rs.Logger.Info().Str("node", nodeID).Str("addr", string(address)).Err(f.Error()).Msg("node join failed")
			return f.Error()
		}
	}

	rs.Logger.Info().Str("node", nodeID).Str("addr", string(address)).Msg("node joined successfully")
	return nil
}

func (rs *RaftService) Leave(nodeID string) error {
	addr := nodeID
	rs.Logger.Info().Str("node", nodeID).Str("addr", addr).Msg("node leave request")

	f := rs.raft.RemoveServer(raft.ServerID(nodeID), 0, 0)
	if f.Error() != nil {
		rs.Logger.Info().Str("node", nodeID).Str("addr", addr).Err(f.Error()).Msg("node leave failed")
		return f.Error()
	}
	rs.Logger.Info().Str("node", nodeID).Str("addr", addr).Msg("node left successfully")
	return nil
}

