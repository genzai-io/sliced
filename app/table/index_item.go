package table

import (
	"github.com/genzai-io/sliced/app/table/index/btree"
	"github.com/genzai-io/sliced/app/table/index/rtree"
)

// IndexItems do not need to be serialized since they
// can be recreated (rebuilt) from the index meta data
//
// Projected keys are converted to the appropriate value type.
// This favors scan speeds. Strings point to a slice of the value
// and are always constant size of a pointer and the string header.
// strings are favored over []byte due to it costing 8 bytes less memory.
// Go structs are 8 byte aligned.
type IndexItem interface {
	btree.Item
	rtree.Item

	// Internal
	index() *Index

	Key() Key
	PK() Key
	Value() *ValueItem
}

//
// Base struct
//
type indexItem struct {
	idx   *Index
	value *ValueItem
}

func (i *indexItem) index() *Index {
	return i.idx
}
func (i *indexItem) Key() Key {
	return Nil
}
func (i *indexItem) PK() Key {
	return i.value.Key
}
func (i *indexItem) Value() *ValueItem {
	return i.value
}

// rtree.Value
func (i *indexItem) Rect(ctx interface{}) (min []float64, max []float64) {
	return nil, nil
}

// btree.Value
func (i *indexItem) Less(than btree.Item, ctx interface{}) bool {
	return false
}

//
//
//
// Nil
//
//
//

type nilItem struct {
	indexItem
}
func (i *nilItem) Key() Key {
	return Nil
}
func (k *nilItem) Less(than btree.Item, ctx interface{}) bool {
	return true
}

//
//
//
// False
//
//
//

type falseItem struct {
	indexItem
}
func (i *falseItem) Key() Key {
	return False
}
func (k *falseItem) Less(than btree.Item, ctx interface{}) bool {
	return False.Less(than, k.value)
}

//
//
//
// True
//
//
//

type trueItem struct {
	indexItem
}
func (i *trueItem) Key() Key {
	return True
}
func (k *trueItem) Less(than btree.Item, ctx interface{}) bool {
	return True.Less(than, k.value)
}

//
//
//
// Rect
//
//
//
// Rect key is for the RTree
type rectItem struct {
	indexItem
	key Rect
}

func (i *rectItem) Key() Key {
	return i.key
}

// rtree.Value
func (r *rectItem) Rect(ctx interface{}) (min []float64, max []float64) {
	return r.key.Min, r.key.Max
}

//
//type anyItem struct {
//	indexItem
//	key Extract
//}
//
//func (k *anyItem) Less(than btree.Item, ctx interface{}) bool {
//	return k.key.LessThanItem(than, k.value)
//}

//
//
//
// String
//
//
//

type stringItem struct {
	indexItem
	key StringKey
}
func (i *stringItem) Key() Key {
	return i.key
}
func (k *stringItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

//
//
//
// String in descending order
//
//
//

type stringDescItem struct {
	indexItem
	key StringDescKey
}
func (i *stringDescItem) Key() Key {
	return i.key
}
func (k *stringDescItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

//
//
//
// String Case Insensitive
//
//
//

type stringCIItem struct {
	indexItem
	key StringCIKey
}
func (i *stringCIItem) Key() Key {
	return i.key
}
func (k *stringCIItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

//
//
//
// String Case Insensitive in descending order
//
//
//

type stringCIDescItem struct {
	indexItem
	key StringCIDescKey
}
func (i *stringCIDescItem) Key() Key {
	return i.key
}
func (k *stringCIDescItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

//
//
//
// Int -> IntKey
//
//
//

type intItem struct {
	indexItem
	key IntKey
}
func (i *intItem) Key() Key {
	return i.key
}
func (k *intItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

type intDescItem struct {
	indexItem
	key IntDescKey
}
func (i *intDescItem) Key() Key {
	return i.key
}
func (k *intDescItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

//
//
//
// Float -> FloatKey
//
//
//

type floatItem struct {
	indexItem
	key FloatKey
}
func (i *floatItem) Key() Key {
	return i.key
}
func (k *floatItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

type floatDescItem struct {
	indexItem
	key FloatDescKey
}
func (i *floatDescItem) Key() Key {
	return i.key
}
func (k *floatDescItem) Less(than btree.Item, ctx interface{}) bool {
	return k.key.LessThanItem(than, k.value)
}

//
//
//
// Composite
//
//
//

type key2Item struct {
	indexItem
	key Key2
}
func (i *key2Item) Key() Key {
	return i.key
}
func (i *key2Item) Less(than btree.Item, ctx interface{}) bool {
	return i.key.LessThanItem(than, i.value)
}

//type composite3Item struct {
//	indexItem
//	K  Key3
//}
//
//func (i *composite3Item) Less(than btree.Item, ctx interface{}) bool {
//	return i.K.LessThanItem(than, i.value)
//}
