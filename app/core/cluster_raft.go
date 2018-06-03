package core

import (
	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/common/raft"

	cmd "github.com/slice-d/genzai/app/command"
)

// IRaftStore represents a raft store that conforms to
// raft.PeerStore, raft.LogStore, and raft.StableStore.
type IRaftStore interface {
	Close() error
	FirstIndex() (uint64, error)
	LastIndex() (uint64, error)
	GetLog(idx uint64, log *raft.Log) error
	StoreLog(log *raft.Log) error
	StoreLogs(logs []*raft.Log) error
	DeleteRange(min, max uint64) error
	Set(k, v []byte) error
	Get(k []byte) ([]byte, error)
	SetUint64(key []byte, val uint64) error
	GetUint64(key []byte) (uint64, error)
	Peers() ([]string, error)
	SetPeers(peers []string) error
}

func (c *ClusterService) Raft() api.RaftService {
	return c
}

func (c *ClusterService) HRaft() *raft.Raft {
	return c.raft
}

// "SHRINK" client command.
func (c *ClusterService) Shrink() error {
	return moved.ErrLogNotShrinkable
}

// "RAFTSHRINK" client command.
func (s *ClusterService) ShrinkLog() error {
	//if s, ok := s.store.(shrinkable); ok {
	//	err := s.Shrink()
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}
	return moved.ErrLogNotShrinkable
}

// Only for RaftTransport RAFTAPPEND
func (s *ClusterService) Append(o []byte, args [][]byte) ([]byte, error) {
	return s.transport.handleAppendEntries(o, args)
	//return nil, nil
}

// Only for RaftTransport RAFTVOTE
func (s *ClusterService) Vote(o []byte, args [][]byte) ([]byte, error) {
	return s.transport.handleRequestVote(o, args)
	//return nil, nil
}

// Only for RaftTransport RAFTINSTALL
func (s *ClusterService) Install(conn *cmd.Context, arg []byte) cmd.Command {
	return s.transport.HandleInstallSnapshot(conn, arg)
	//return nil
}

func (s *ClusterService) Snapshot() error {
	f := s.raft.Snapshot()
	if err := f.Error(); err != nil {
		return err
	}
	return nil
}

func (c *ClusterService) IsLeader() bool {
	return moved.ClusterAddress == c.raft.Leader()
}

func (s *ClusterService) Leader() raft.ServerAddress {
	return s.raft.Leader()
}

func (s *ClusterService) Stats() map[string]string {
	return s.raft.Stats()
}

func (s *ClusterService) State() raft.RaftState {
	return s.raft.State()
}

func (s *ClusterService) Configuration() (raft.ConfigurationFuture, error) {
	future := s.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return future, err
	}
	return future, nil
}

func (s *ClusterService) Bootstrap() error {
	s.Logger.Info().Msg("bootstrap request")

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		s.Logger.Error().AnErr("err", err).Msg("failed to get raft configuration")
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

	future := s.raft.BootstrapCluster(configuration)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}

func (s *ClusterService) Demote(nodeID string) error {
	s.Logger.Info().Str("node", nodeID).Str("addr", nodeID).Msg("node demote request")

	serverID := raft.ServerID(nodeID)
	address := raft.ServerAddress(nodeID)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		s.Logger.Error().AnErr("err", err).Msg("failed to get raft configuration")
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == address {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == address && srv.ID == serverID && srv.Suffrage != raft.Voter {
				s.Logger.Warn().Msgf("node %s at %s is a %s and not a Voter, ignoring demote request", nodeID, address, srv.Suffrage)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				s.Logger.Error().AnErr("err", err).Msgf("error removing existing node %s at %s: %s", nodeID, address, err)
				return err
			}
		}
	}

	future := s.raft.DemoteVoter(serverID, 0, 0)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *ClusterService) Join(nodeID string, voter bool) error {
	s.Logger.Info().Str("node", nodeID).Str("addr", nodeID).Msg("node join request")

	serverID := raft.ServerID(nodeID)
	address := raft.ServerAddress(nodeID)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		s.Logger.Error().AnErr("err", err).Msg("failed to get raft configuration")
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == address {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == address && srv.ID == serverID {
				s.Logger.Warn().Msgf("node %s at %s already member of cluster, ignoring join request", nodeID, address)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				s.Logger.Error().AnErr("err", err).Msgf("error removing existing node %s at %s: %s", nodeID, address, err)
				return err
			}
		}
	}

	if voter {
		f := s.raft.AddVoter(serverID, address, 0, 0)
		if err := f.Error(); err != nil {
			s.Logger.Info().Str("node", nodeID).Str("addr", string(address)).Err(f.Error()).Msg("node join failed")
			return f.Error()
		}
	} else {
		f := s.raft.AddNonvoter(serverID, address, 0, 0)
		if err := f.Error(); err != nil {
			s.Logger.Info().Str("node", nodeID).Str("addr", string(address)).Err(f.Error()).Msg("node join failed")
			return f.Error()
		}
	}

	s.Logger.Info().Str("node", nodeID).Str("addr", string(address)).Msg("node joined successfully")
	return nil
}

func (s *ClusterService) Leave(nodeID string) error {
	addr := nodeID
	s.Logger.Info().Str("node", nodeID).Str("addr", addr).Msg("node leave request")

	f := s.raft.RemoveServer(raft.ServerID(nodeID), 0, 0)
	if f.Error() != nil {
		s.Logger.Info().Str("node", nodeID).Str("addr", addr).Err(f.Error()).Msg("node leave failed")
		return f.Error()
	}
	s.Logger.Info().Str("node", nodeID).Str("addr", addr).Msg("node left successfully")
	return nil
}
