package api

import "github.com/genzai-io/sliced/common/raft"

const (
	GetName    = "GET"
	SetName    = "SET"
	DelName    = "DEL"
	TxModeName = "TX"

	CreateDatabase = "CREATEDB"
	DeleteDatabase = "DELETEDB"

	RaftInstallSnapshotName = "I"
	RaftAppendName          = "A"
	RaftVoteName            = "V"
	RaftChunkName           = "C"
	RaftDoneName            = "D"
	RaftSnapshotName        = "RSNAPNAME"
	RaftSnapshotsName       = "RSNAPS"
	RaftSlice               = "RSLICE"
	RaftBootstrap           = "BOOTSTRAP"
	RaftJoinName            = "JOIN"
	RaftDemote              = "DEMOTE"
	RaftJoinSlaveName       = "RJOINSLAVE"
	RaftRemoveName          = "REMOVE"
	RaftStatsName           = "RAFTSTATS"
	RaftStateName           = "RAFTSTATE"
	RaftConfigName          = "RCONFIG"
	RaftLeaderName          = "LEADER"
	RaftShrinkName          = "RSHRINK"
)

var (
	RaftInstall = []byte(RaftInstallSnapshotName)
	RaftAppend  = []byte(RaftAppendName)
	RaftVote    = []byte(RaftVoteName)
	RaftChunk   = []byte(RaftChunkName)
	RaftDone    = []byte(RaftDoneName)
	RaftJoin    = []byte(RaftJoinName)
	RaftStats   = []byte(RaftStatsName)
	RaftState   = []byte(RaftStateName)
	RaftLeader  = []byte(RaftLeaderName)
	RaftShrink  = []byte(RaftShrinkName)
)

type RaftService interface {
	IsLeader() bool

	Leader() raft.ServerAddress

	Stats() map[string]string

	State() raft.RaftState

	Append(payload []byte) CommandReply

	Vote(payload []byte) CommandReply

	Install(ctx *Context, arg []byte) Command

	Bootstrap() error

	Join(nodeID string, voter bool) error

	Demote(nodeID string) error

	Leave(nodeID string) error

	Configuration() (raft.ConfigurationFuture, error)
}

type RaftFSM raft.FSM

func GetRaftService(id RaftID) RaftService {
	if id.DatabaseID < 0 {
		return Cluster.Raft()
	} else {
		return nil
	}
}
