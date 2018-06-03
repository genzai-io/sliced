package table

import (
	"os"
	"sync"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/table/index/btree"
	"github.com/slice-d/genzai/app/table/index/rtree"
)

// Default number of btree degrees
//const btreeDegrees = 64

// exctx is a simple b-tree context for ordering by expiration.
type exctx struct {
	db *Table
}

var defaultFreeList = new(btree.FreeList)

//
type Table struct {
	commitIndex uint64

	hash   map[Key]*ValueItem
	items  *btree.BTree
	idxs   map[string]*Index
	exps   *btree.BTree
	closed bool
	mu     sync.RWMutex

	// Stats
	itemMemory             uint64
	itemMemoryUncompressed uint64
}

func NewTable() *Table {
	return NewTableWithFreelist(defaultFreeList)
}

func NewTableWithFreelist(list *btree.FreeList) *Table {
	s := &Table{}
	s.items = btree.NewWithFreeList(btreeDegrees, list, nil)
	s.exps = btree.NewWithFreeList(btreeDegrees, list, nil)
	s.idxs = make(map[string]*Index)
	return s
}

func (s *Table) Length() uint64 {
	return uint64(s.items.Len())
}

func (s *Table) insertIndex(idx *Index) {
	s.idxs[idx.name] = idx
}

func (s *Table) removeIndex(idx *Index) {
	// delete from the map.
	// this is all that is needed to delete an idx.
	delete(s.idxs, idx.name)
}

// View executes a function within a managed read-only transaction.
// When a non-nil error is returned from the function that error will be return
// to the caller of View().
//
// Executing a manual commit or rollback from inside the function will result
// in a panic.
func (s *Table) View(fn func() error) error {
	s.mu.RLock()
	err := fn()
	s.mu.RUnlock()
	return err
}

// Update executes a function within a managed read/write transaction.
// The transaction has been committed when no error is returned.
// In the event that an error is returned, the transaction will be rolled back.
// When a non-nil error is returned from the function, the transaction will be
// rolled back and the that error will be return to the caller of Update().
//
// Executing a manual commit or rollback from inside the function will result
// in a panic.
func (s *Table) Update(fn func() error) error {
	s.mu.Lock()
	err := fn()
	s.mu.Unlock()
	return err
}

// get return an value or nil if not found.
func (s *Table) get(key Key) *ValueItem {
	item := s.items.Get(key)
	if item != nil {
		return item.(*ValueItem)
	}
	return nil
}

// DeleteAll deletes all items from the database.
func (s *Table) DeleteAll() error {
	// now reset the live database trees
	s.items = btree.NewWithFreeList(btreeDegrees, defaultFreeList, nil)
	s.exps = btree.NewWithFreeList(btreeDegrees, defaultFreeList, &exctx{s})
	s.idxs = make(map[string]*Index)
	return nil
}

// insert performs inserts an value in to the database and updates
// all indexes. If a previous value with the same key already exists, that value
// will be replaced with the new one, and return the previous value.
func (s *Table) insert(item *ValueItem) *ValueItem {
	var pdbi *ValueItem
	prev := s.items.ReplaceOrInsert(item)
	//_i := -1
	if prev != nil {
		// A previous value was removed from the keys tree. Let's
		// fully delete this value from all indexes.
		pdbi = prev.(*ValueItem)
		item.Indexes = pdbi.Indexes

		if pdbi.Expires > 0 {
			// Remove it from the exipres tree.
			s.exps.Delete(pdbi)
		}

		for _, sec := range item.Indexes {
			if sec != nil {
				sec.index().remove(sec)
				idxItem := sec.index().indexer.Index(sec.index(), item)
				if idxItem != nil {
					sec.index().btr.ReplaceOrInsert(idxItem)
				}
			}
		}
	}
	if item.Expires > 0 {
		// The new value has eviction options. Add it to the
		// expires tree
		s.exps.ReplaceOrInsert(item)
	}
	if prev == nil {
		for _, idx := range s.idxs {
			if !idx.match(item.Key) {
				continue
			}

			sk := idx.indexer.Index(idx, item)
			if sk == nil {
				continue
			}

			item.Indexes = append(item.Indexes, sk)

			if idx.btr != nil {
				if sk != nil {
					idx.btr.ReplaceOrInsert(sk)
				} else {
					// Ignored.
				}
			} else if idx.rtr != nil {
				if sk != nil {
					idx.rtr.Insert(sk)
				} else {
					// Ignored.
				}
			}
		}
	}
	// we must return the previous value to the caller.
	return pdbi
}

// delete removes and value from the database and indexes. The input
// value must only have the key field specified thus "&dbItem{key: key}" is all
// that is needed to fully remove the value with the matching key. If an value
// with the matching key was found in the database, it will be removed and
// returned to the caller. A nil return value means that the value was not
// found in the database
func (s *Table) delete(item *ValueItem) *ValueItem {
	var pdbi *ValueItem
	prev := s.items.Delete(item)
	if prev != nil {
		pdbi = prev.(*ValueItem)
		if pdbi.Expires > 0 {
			// Remove it from the exipres tree.
			s.exps.Delete(pdbi)
		}
		for _, sec := range pdbi.Indexes {
			if sec != nil {
				sec.index().remove(sec)
			}
		}
	}
	return pdbi
}

//
//
//
func (s *Table) Set(key Key, value string, expires int64) (previousValue string,
	replaced bool, err error) {
	var prev *ValueItem
	//prev = s.get(key)
	//if prev != nil {
	//	return prev.Value, true, nil
	//}

	item := &ValueItem{Key: key, Value: value}
	if expires > 0 {
		// The caller is requesting that this value expires. Convert the
		// TTL to an absolute time and bind it to the value.
		item.Expires = expires
	}
	// Insert the value into the keys tree.
	prev = s.insert(item)

	if prev == nil {
		return "", false, nil
	} else {
		return prev.Value, true, nil
	}
}

// SliceForKey returns a value for a key. If the value does not exist or if the value
// has expired then ErrNotFound is returned.
func (s *Table) Get(key Key) (val string, err error) {
	if s == nil {
		return "", os.ErrNotExist
	}
	item := s.get(key)
	if item == nil || item.expired() {
		// The value does not exists or has expired. Let's assume that
		// the caller is only interested in items that have not expired.
		return "", moved.ErrNotFound
	}
	return item.Value, nil
}

// Delete removes an value from the database based on the value's key. If the value
// does not exist or if the value has expired then ErrNotFound is returned.
//
// Only a writable transaction can be used for this operation.
// This operation is not allowed during iterations such as Ascend* & Descend*.
func (s *Table) Delete(key Key) (val string, err error) {

	item := s.delete(&ValueItem{Key: key})
	if item == nil {
		return "", moved.ErrNotFound
	}

	// Even though the value has been deleted, we still want to check
	// if it has expired. An expired value should not be returned.
	if item.expired() {
		// The value exists in the tree, but has expired. Let's assume that
		// the caller is only interested in items that have not expired.
		return "", moved.ErrNotFound
	}
	return item.Value, nil
}

//
//
//
func (s *Table) scanPrimary(desc, gt, lt bool, start, stop Key,
	iterator func(value *ValueItem) bool) error {
	// wrap a btree specific iterator around the user-defined iterator.
	iter := func(item btree.Item) bool {
		switch dbi := item.(type) {
		case *ValueItem:
			return iterator(dbi)
		}
		return false
	}

	// create some limit items
	var itemA, itemB Key
	if gt || lt {
		itemA = start
		itemB = stop
	}

	if desc {
		if gt {
			if lt {
				s.items.DescendRange(itemA, itemB, iter)
			} else {
				s.items.DescendGreaterThan(itemA, iter)
			}
		} else if lt {
			s.items.DescendLessOrEqual(itemA, iter)
		} else {
			s.items.Descend(iter)
		}
	} else {
		if gt {
			if lt {
				s.items.AscendRange(itemA, itemB, iter)
			} else {
				s.items.AscendGreaterOrEqual(itemA, iter)
			}
		} else if lt {
			s.items.AscendLessThan(itemA, iter)
		} else {
			s.items.Ascend(iter)
		}
	}

	return nil
}

//
//
//
func (s *Table) scanSecondary(desc, gt, lt bool, index string, start, stop Key,
	iterator func(key IndexItem) bool) error {
	// wrap a btree specific iterator around the user-defined iterator.
	iter := func(item btree.Item) bool {
		dbi, ok := item.(IndexItem)
		if !ok {
			return false
		} else {
			return iterator(dbi)
		}
	}

	idx := s.idxs[index]
	if idx == nil {
		// idx was not found. return error
		return moved.ErrNotFound
	}
	tr := idx.btr
	if tr == nil {
		return nil
	}

	// create some limit items
	var itemA, itemB Key
	if gt || lt {
		itemA = start
		itemB = stop
	}

	if desc {
		if gt {
			if lt {
				tr.DescendRange(itemA, itemB, iter)
			} else {
				tr.DescendGreaterThan(itemA, iter)
			}
		} else if lt {
			tr.DescendLessOrEqual(itemA, iter)
		} else {
			tr.Descend(iter)
		}
	} else {
		if gt {
			if lt {
				tr.AscendRange(itemA, itemB, iter)
			} else {
				tr.AscendGreaterOrEqual(itemA, iter)
			}
		} else if lt {
			tr.AscendLessThan(itemA, iter)
		} else {
			tr.Ascend(iter)
		}
	}
	return nil
}

// Nearby searches for rectangle items that are nearby a target rect.
// All items belonging to the specified idx will be returned in order of
// nearest to farthest.
// The specified idx must have been created by AddIndex() and the target
// is represented by the rect string. This string will be processed by the
// same bounds function that was passed to the CreateSpatialIndex() function.
// An invalid idx will return an error.
func (s *Table) Nearby(index, bounds string,
	iterator func(key Rect, value *ValueItem, dist float64) bool) error {
	if index == "" {
		// cannot search on keys tree. just return nil.
		return nil
	}
	// // wrap a rtree specific iterator around the user-defined iterator.
	iter := func(item rtree.Item, dist float64) bool {
		dbi, ok := item.(*rectItem)
		if !ok {
			return true
		}
		return iterator(dbi.key, dbi.value, dist)
	}
	idx := s.idxs[index]
	if idx == nil {
		// idx was not found. return error
		return moved.ErrNotFound
	}
	if idx.rtr == nil {
		// not an r-tree idx. just return nil
		return nil
	}
	// execute the nearby search

	// set the center param to false, which uses the box dist calc.
	idx.rtr.KNN(ParseRect(bounds), false, iter)
	return nil
}

// Intersects searches for rectangle items that intersect a target rect.
// The specified idx must have been created by AddIndex() and the target
// is represented by the rect string. This string will be processed by the
// same bounds function that was passed to the CreateSpatialIndex() function.
// An invalid idx will return an error.
func (s *Table) Intersects(index, bounds string,
	iterator func(key *rectItem, value *ValueItem) bool) error {
	if index == "" {
		// cannot search on keys tree. just return nil.
		return nil
	}
	// wrap a rtree specific iterator around the user-defined iterator.
	iter := func(item rtree.Item) bool {
		dbi := item.(*rectItem)
		return iterator(dbi, dbi.value)
	}
	idx := s.idxs[index]
	if idx == nil {
		// idx was not found. return error
		return moved.ErrNotFound
	}
	if idx.rtr == nil {
		// not an r-tree idx. just return nil
		return nil
	}

	idx.rtr.Search(ParseRect(bounds), iter)
	return nil
}

// Ascend calls the iterator for every value in the database within the range
// [first, last], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) AscendPrimary(iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(false, false, false, MinKey, MaxKey, iterator)
}

// AscendGreaterOrEqual calls the iterator for every value in the database within
// the range [pivot, last], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) AscendGreaterOrEqualPrimary(pivot Key,
	iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(false, true, false, pivot, MaxKey, iterator)
}

// AscendLessThan calls the iterator for every value in the database within the
// range [first, pivot), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) AscendLessThanPrimary(pivot Key,
	iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(false, false, true, pivot, MaxKey, iterator)
}

// AscendRange calls the iterator for every value in the database within
// the range [greaterOrEqual, lessThan), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) AscendRangePrimary(greaterOrEqual, lessThan Key,
	iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(
		false, true, true, greaterOrEqual, lessThan, iterator)
}

// Descend calls the iterator for every value in the database within the range
// [last, first], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) DescendPrimary(iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(true, false, false, MinKey, MinKey, iterator)
}

// DescendGreaterThan calls the iterator for every value in the database within
// the range [last, pivot), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) DescendGreaterThanPrimary(pivot Key,
	iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(true, true, false, pivot, MinKey, iterator)
}

// DescendLessOrEqual calls the iterator for every value in the database within
// the range [pivot, first], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) DescendLessOrEqualPrimary(pivot Key,
	iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(true, false, true, pivot, MinKey, iterator)
}

// DescendRange calls the iterator for every value in the database within
// the range [lessOrEqual, greaterThan), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) DescendRangePrimary(lessOrEqual, greaterThan Key,
	iterator func(value *ValueItem) bool) error {
	return s.scanPrimary(
		true, true, true, lessOrEqual, greaterThan, iterator,
	)
}

// Ascend calls the iterator for every value in the database within the range
// [first, last], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) Ascend(index string, iterator IndexIterator) error {
	return s.scanSecondary(false, false, false, index, MinKey, MaxKey, iterator)
}

// AscendGreaterOrEqual calls the iterator for every value in the database within
// the range [pivot, last], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) AscendGreaterOrEqual(index string, pivot Key, iterator IndexIterator) error {
	return s.scanSecondary(false, true, false, index, pivot, MinKey, iterator)
}

// AscendLessThan calls the iterator for every value in the database within the
// range [first, pivot), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) AscendLessThan(index string, pivot Key, iterator IndexIterator) error {
	return s.scanSecondary(false, false, true, index, pivot, MinKey, iterator)
}

// AscendRange calls the iterator for every value in the database within
// the range [greaterOrEqual, lessThan), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) AscendRange(index string, greaterOrEqual, lessThan Key, iterator IndexIterator) error {
	return s.scanSecondary(
		false, true, true, index, greaterOrEqual, lessThan, iterator)
}

// Descend calls the iterator for every value in the database within the range
// [last, first], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) Descend(index string, iterator IndexIterator) error {
	return s.scanSecondary(true, false, false, index, MaxKey, MinKey, iterator)
}

// DescendGreaterThan calls the iterator for every value in the database within
// the range [last, pivot), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) DescendGreaterThan(index string, pivot Key, iterator IndexIterator) error {
	return s.scanSecondary(true, true, false, index, pivot, MinKey, iterator)
}

// DescendLessOrEqual calls the iterator for every value in the database within
// the range [pivot, first], until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) DescendLessOrEqual(index string, pivot Key, iterator IndexIterator) error {
	return s.scanSecondary(true, false, true, index, pivot, MinKey, iterator)
}

// DescendRange calls the iterator for every value in the database within
// the range [lessOrEqual, greaterThan), until iterator returns false.
// When an idx is provided, the results will be ordered by the value values
// as specified by the less() function of the defined idx.
// When an idx is not provided, the results will be ordered by the value key.
// An invalid idx will return an error.
func (s *Table) DescendRange(index string, lessOrEqual, greaterThan Key, iterator IndexIterator) error {
	return s.scanSecondary(
		true, true, true, index, lessOrEqual, greaterThan, iterator,
	)
}

//// Point is a helper function that converts a series of float64s
//// to a rectangle for a spatial idx.
//func Point(coords ...float64) string {
//	return Rect(coords, coords)
//}
