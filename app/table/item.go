package table

import (
	"errors"
	"time"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/table/index/btree"
)

var (
	ErrNotJSON = errors.New("not json")
)

// Value represents a value in the DB
// It can be set to auto expire
// It can have secondary indexes
type ValueItem struct {
	Key     Key
	Value   string
	Expires int64
	Slot    uint16
	LogID   uint64
	Indexes []IndexItem
}

// BTree key comparison.
func (dbi *ValueItem) Less(than btree.Item, ctx interface{}) bool {
	//return dbi.K < than.(*Value).K
	return dbi.Key.Less(than, ctx)
}

// expired evaluates id the value has expired. This will always return false when
// the value does not have `opts.ex` set to true.
func (dbi *ValueItem) expired() bool {
	return dbi.Expires > 0 && time.Now().Unix() > dbi.Expires
	//return dbi.opts != nil && dbi.opts.ex && time.Now().After(dbi.opts.exat)
}

// expiresAt will return the time when the value will expire. When an value does
// not expire `maxTime` is used.
func (dbi *ValueItem) expiresAt() int64 {
	return dbi.Expires
}

func (dbi *ValueItem) Type() moved.DataType {
	return dbi.Key.Type()
}

func (dbi *ValueItem) Match(pattern string) bool {
	return dbi.Key.Match(pattern)
}

func (dbi *ValueItem) LessThan(key Key) bool {
	return dbi.LessThan(key)
}

func (dbi *ValueItem) LessThanItem(than btree.Item, item *ValueItem) bool {
	return dbi.LessThanItem(than, item)
}
