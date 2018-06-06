package core

import (
	"strings"
	"time"

	"github.com/genzai-io/sliced/btrdb"
	"github.com/genzai-io/sliced/proto/store"
	"sync"
	"github.com/genzai-io/sliced/common/service"
	"github.com/genzai-io/sliced"
)

type databaseStore struct {
	sync.RWMutex
	service.BaseService

	db      *btrdb.DB
	ByNames *btrdb.TableIndex

	byID   map[int32]*Database
	byName map[string]*Database

	tblDatabases *btrdb.Table
}

func newDatabases(db *btrdb.DB) *databaseStore {
	d := &databaseStore{
		db: db,
		tblDatabases: newTable(
			"db",
			func() Serializable { return &store.Database{} },
			btrdb.NewTableIndex("names", nameProjector),
		),
	}

	d.ByNames = d.tblDatabases.Secondary[0]

	d.BaseService = *service.NewBaseService(moved.Logger, "store-databases", d)

	return d
}

func (d *databaseStore) OnStart() error {
	return d.db.Update(func(tx *btrdb.Tx) error {
		if err := d.tblDatabases.Build(tx); err != nil {
			return err
		}

		return d.load(tx)
	})
}

func (d *databaseStore) OnStop() {
	d.Lock()
	defer d.Unlock()

	if !d.IsRunning() {
		return
	}

	for _, v := range d.byID {
		v.Stop()
	}
}

func (t *databaseStore) load(tx *btrdb.Tx) (err error) {
	t.Lock()
	defer t.Unlock()

	if t.byID == nil {
		t.byID = make(map[int32]*Database)
	}

	// Clear
	t.byName = make(map[string]*Database)

	var created []*Database = nil

	// Populate map
	if e := t.tblDatabases.Ascend(tx, func(key string, value btrdb.Serializable) bool {
		model, ok := value.(*store.Database)
		if !ok {
			return true
		}

		database, ok := t.byID[model.Id]
		if !ok {
			database := newDatabase(t.db, model)
			created = append(created, database)
			t.byID[model.Id] = database
		} else {
			database.model = *model
		}

		t.byName[model.Name] = database

		return true
	}); e != nil {
		return e
	}

	if len(created) > 0 {
		for _, database := range created {
			if err = database.Start(); err != nil {
				return err
			}
		}
	}

	return
}

func (t *databaseStore) selectAll(index *btrdb.TableIndex) (databases []*store.Database, err error) {
	databases = make([]*store.Database, 0, 8)

	err = t.db.View(func(tx *btrdb.Tx) error {
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
func (t *databaseStore) Insert(name string) (database *store.Database, err error) {
	database = nil

	err = t.db.Update(func(tx *btrdb.Tx) error {
		// Trim the name
		name = strings.TrimSpace(name)

		// Check if the name is already used
		if btrdb.Contains(tx, t.ByNames.Format(strings.ToLower(name))) {
			return btrdb.ErrDuplicateKey
		}

		// Calculate ID sequence
		id, err := t.tblDatabases.NextID(tx)
		if err != nil {
			return err
		}

		// Create new Database document
		now := uint64(time.Now().UnixNano())
		database = &store.Database{
			Id:      int32(id),
			Name:    name,
			Created: now,
			Changed: now,
			Dropped: 0,
			Removed: 0,
		}

		// Insert into db
		return t.tblDatabases.Insert(tx, database)
	})

	return
}
