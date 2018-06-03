package core

import (
	"runtime"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/pool"
	"github.com/slice-d/genzai/proto/store"
	"github.com/slice-d/genzai/app/api"
)

// A single instance of the application. A node may be part of any
// number of Database Slices.
type Node struct {
	model store.Node

	groups map[string]*NodeGroup
	boot   bool
	local  bool
	slices []*Slice

	// Transport to use to send RESP requests if not local
	transport *RaftTransport

	queue []api.Command
}

func newNode(model *store.Node, local bool) *Node {
	n := &Node{
		groups: make(map[string]*NodeGroup),
		model:  *model,
		local:  local,
		slices: make([]*Slice, 0),
	}
	return n
}

func newLocalNode() *Node {
	node := &Node{}
	node.syncModel()
	return node
}

func (n *Node) syncModel() {
	n.model = store.Node{
		Bootstrap:  moved.Bootstrap,
		Id:         string(moved.ClusterID),
		Host:       string(moved.ClusterAddress),
		Version:    moved.VersionStr,
		InstanceID: moved.InstanceID,
		Region:     moved.Region,
		Cores:      uint32(runtime.NumCPU()),
		Memory:     pool.MaxMemory,
		WebHost:    moved.WebHost,
		ApiHost:    moved.ApiHost,
		ApiLoops:   uint32(moved.EventLoops),

		Drives: moved.GetDrivesList(),
	}

	// Send update to leader
}
