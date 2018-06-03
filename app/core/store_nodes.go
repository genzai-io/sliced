package core

import (
	"strings"
	"github.com/slice-d/genzai/btrdb"
	"github.com/slice-d/genzai/proto/store"
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
		groups: make(map[string]*store.NodeGroup),
	}
)

func init() {
	StoreSchema.Add(TblNodes.Table)
}

type tblNodes struct {
	*btrdb.Table

	local *Node
}

type tblNodeGroups struct {
	*btrdb.Table

	groups map[string]*store.NodeGroup
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

