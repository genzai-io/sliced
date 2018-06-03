package core

import (
	"strings"
	"github.com/genzai-io/sliced/btrdb"
	"github.com/genzai-io/sliced/proto/store"
	"runtime"
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/pool"
)

var (
	TblNodes = &tblNodes{
		Table: newTable(
			"n",
			func() Serializable { return &store.Node{} },
		),
	}

	TblNodeGroups = &tblNodeGroups{
		Table: newTable(
			"ng",
			func() Serializable { return &store.NodeGroup{} },
		),
		m: make(map[string]*NodeGroup),
	}
)

func init() {
	StoreSchema.Add(TblNodes.Table)
}

type tblNodes struct {
	*btrdb.Table

	local *Node

	m map[string]*Node
}

type tblNodeGroups struct {
	*btrdb.Table

	m map[string]*NodeGroup
}

type nodeService struct {
}

func (t *tblNodeGroups) init(db *btrdb.DB) error {
	// Populate map
	return db.View(func(tx *btrdb.Tx) error {
		t.Ascend(tx, func(key string, value btrdb.Serializable) bool {
			group, ok := value.(*store.NodeGroup)
			if !ok {
				return true
			}
			//t.groups[group.Name],
			_ = group
			return true
		})
		return nil
	})
}

func (t *tblNodes) Select(db *btrdb.DB, index *btrdb.TableIndex) (databases []*store.Database, err error) {
	databases = make([]*store.Database, 0, 8)

	err = db.View(func(tx *btrdb.Tx) error {
		index.Ascend(tx, func(key string, value btrdb.Serializable) bool {
			database, ok := value.(*store.Database)
			if !ok {
				return true
			}
			databases = append(databases, database)
			return true
		})

		return nil
	})

	return
}

// Safely creates and inserts a new store.Database document
func (t *tblNodes) Insert(db *btrdb.DB, address string) (node *store.Node, err error) {
	node = nil

	err = db.Update(func(tx *btrdb.Tx) error {
		// Trim the name
		address = strings.TrimSpace(address)

		// Create new Database document
		//now := uint64(time.Now().UnixNano())
		node = &store.Node{
			Id:         address,
			Host:       address,
			Version:    "",
			InstanceID: "",
			Region:     "",
			Zone:       "",
			Cores:      0,
			Memory:     0,
			Bootstrap:  false,
			WebHost:    "",
			ApiHost:    "",
			ApiLoops:   0,
			Drives:     nil,
		}

		// Insert into db
		return t.Table.Insert(tx, node)
	})

	return
}

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

type NodeGroup struct {
	model store.NodeGroup
}
