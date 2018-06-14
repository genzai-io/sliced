package slice

import (
	"github.com/genzai-io/sliced/common/btrdb"
)

var StoreSchema = &btrdb.Schema{}

func newTable(name string, factory btrdb.Factory, secondary ...*btrdb.TableIndex) *btrdb.Table {
	return btrdb.NewTable(name, factory, idProjector, secondary...)
}

type Serializable = btrdb.Serializable

type HasName interface {
	GetName() string
}

type HasStringId interface {
	GetId() string
}

type HasInt32Id interface {
	GetId() int32
}

type HasUint32Id interface {
	GetId() int32
}

type HasUint64Id interface {
	GetId() uint64
}

type HasInt64Id interface {
	GetId() int64
}

var idProjector btrdb.Projector = func(val interface{}) interface{} {
	switch v := val.(type) {
	case HasStringId:
		return v.GetId()

	case HasInt32Id:
		return v.GetId()

	case HasUint32Id:
		return v.GetId()

	case HasInt64Id:
		return v.GetId()

	case HasUint64Id:
		return v.GetId()
	}
	return ""
}

var nameProjector btrdb.Projector = func(val interface{}) interface{} {
	switch v := val.(type) {
	case HasName:
		return v.GetName()
	}
	return ""
}

func dbBuildStoreSchema(db *btrdb.DB) error {
	return db.Update(func(tx *btrdb.Tx) error {
		return StoreSchema.Create(tx)
	})
}
