package core

import (
	"github.com/genzai-io/sliced/common/btrdb"
	"github.com/genzai-io/sliced/proto/store"
	"runtime"
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/pool"
	"sync"
	"github.com/genzai-io/sliced/common/service"
)

// Global node store
type nodeStore struct {
	service.BaseService

	sync.RWMutex

	db *btrdb.DB

	local *Node

	nodesByID    map[string]*Node
	groupsByID   map[int64]*NodeGroup
	groupsByName map[string]*NodeGroup

	tblNodes      *btrdb.Table
	tblNodeGroups *btrdb.Table
}

func newNodeStore(db *btrdb.DB) *nodeStore {
	n := &nodeStore{
		nodesByID:    make(map[string]*Node),
		groupsByID:   make(map[int64]*NodeGroup),
		groupsByName: make(map[string]*NodeGroup),
		tblNodes: newTable(
			"nodes",
			func() Serializable { return &store.Node{} },
		),
		tblNodeGroups: newTable(
			"groups",
			func() Serializable { return &store.NodeGroup{} },
		),
	}

	n.BaseService = *service.NewBaseService(moved.Logger, "store-nodes", n)

	return n
}

func (n *nodeStore) OnStart() (err error) {
	n.Lock()
	defer n.Unlock()

	n.groupsByName = make(map[string]*NodeGroup)

	var groups []*NodeGroup
	var nodes []*Node

	if err = n.db.Update(func(tx *btrdb.Tx) error {
		if err = n.tblNodeGroups.Ascend(tx, func(key string, value btrdb.Serializable) bool {
			model, ok := value.(*store.NodeGroup)
			if !ok {
				return true
			}

			node, ok := n.groupsByID[model.Id]
			if !ok {
				group := newNodeGroup(model)
				n.groupsByID[model.Id] = group
				groups = append(groups, group)
			} else {
				node.model = *model
			}

			n.groupsByName[model.Name] = node

			return true
		}); err != nil {
			return err
		}

		if err = n.tblNodes.Ascend(tx, func(key string, value btrdb.Serializable) bool {
			model, ok := value.(*store.Node)
			if !ok {
				return true
			}

			node, ok := n.nodesByID[model.Id]
			if !ok {
				local := string(moved.ClusterAddress) == model.Id
				node := newNode(model, local)
				n.nodesByID[model.Id] = node
				nodes = append(nodes, node)

				if local {
					n.local = node
					node.populateModel()
					if _, _, err = n.tblNodes.Update(tx, &node.model); err != nil {
						return false
					}
				}
			} else {
				node.model = *model
			}

			return true
		}); err != nil {
			return err
		}

		// Create and insert local node if it doesn't exist
		if n.local == nil {
			n.local = newNode(&store.Node{}, true)
			if err = n.tblNodes.Insert(tx, &n.local.model); err != nil {
				return err
			}
			n.nodesByID[n.local.model.Id] = n.local
		}

		return nil
	}); err != nil {
		return
	}

	if len(groups) > 0 {
		for _, group := range groups {
			_ = group
		}
	}

	return
}

// A single instance of the application. A node may be part of any
// number of Database Slices.
type Node struct {
	model store.Node

	groups map[string]*NodeGroup
	boot   bool
	local  bool

	// Transport to use to send RESP requests if not local
	transport NodeTransport
}

func newNode(model *store.Node, local bool) *Node {
	n := &Node{
		groups:    make(map[string]*NodeGroup),
		model:     *model,
		local:     local,
		transport: newNodeTransport(model.Id),
	}

	if !local {

	}
	return n
}

func (n *Node) populateModel() {
	if !n.local {
		return
	}

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

type NodeGroup struct {
	model store.NodeGroup
}

func newNodeGroup(model *store.NodeGroup) *NodeGroup {
	n := &NodeGroup{
		model: *model,
	}
	return n
}
