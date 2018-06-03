package core

import (
	"strings"
	"time"

	"github.com/genzai-io/sliced/btrdb"
	"github.com/genzai-io/sliced/proto/store"
)

var (
	TblDatabase = &tblDatabase{
		Table: newTable(
			"db",
			func() Serializable { return &store.Database{} },
			btrdb.NewTableIndex("names", NameExtractor),
		),
	}
)

func init() {
	TblDatabase.Names = TblDatabase.Secondary[0]

	StoreSchema.Add(TblDatabase.Table)
}

type tblDatabase struct {
	*btrdb.Table
	Names *btrdb.TableIndex
}

func (t *tblDatabase) Select(db *btrdb.DB, index *btrdb.TableIndex) (databases []*store.Database, err error) {
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
func (t *tblDatabase) Insert(db *btrdb.DB, name string) (database *store.Database, err error) {
	database = nil

	err = db.Update(func(tx *btrdb.Tx) error {
		// Trim the name
		name = strings.TrimSpace(name)

		// Check if the name is already used
		if btrdb.Contains(tx, TblDatabase.Names.Format(strings.ToLower(name))) {
			return btrdb.ErrDuplicateKey
		}

		// Calculate ID sequence
		id, err := TblDatabase.NextID(tx)
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
		return t.Table.Insert(tx, database)
	})

	return
}

func (s *Store) deleteDatabase() {

}
