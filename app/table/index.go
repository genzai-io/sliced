package table

import (
	"io"

	"github.com/armon/go-radix"
	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/table/index/btree"
	"github.com/slice-d/genzai/app/table/index/rtree"
)

const btreeDegrees = 32

type IndexIterator func(item IndexItem) bool

type IndexType uint8

const (
	BTree   IndexType = 0
	RTree   IndexType = 1
	RadTree IndexType = 2
)

// idx represents a b-tree or r-tree idx and also acts as the
// b-tree/r-tree context for itself.
type Index struct {
	db      *Table // the origin database
	t       IndexType
	btr     *btree.BTree // contains the items
	rtr     *rtree.RTree // contains the items
	radtr   *radix.Tree
	name    string // name of the idx
	pattern string // a required key pattern, more fuzzy
	prefix  string // key prefix, table pattern

	// Stats - index structure may contain additional stats
	length uint64 // Number of entries
	memory uint64 // Estimated size of index in main memory

	// Indexer projects keys from raw values and creates IndexItems
	indexer Indexer
}

func (i *Index) Length() int {
	if i.btr != nil {
		return i.btr.Len()
	} else if i.rtr != nil {
		return i.rtr.Count()
	} else {
		return 0
	}
}

func (i *Index) remove(item IndexItem) IndexItem {
	if i.btr != nil {
		r := i.btr.Delete(item)
		if r != nil {
			return r.(IndexItem)
		} else {
			return nil
		}
	} else if i.rtr != nil {
		i.rtr.Remove(item)
		return nil
	}
	return nil
}

// Snapshot of an idx only includes meta-data to re-create the idx since
// it is built from the items in the tree.
func (i *Index) Snapshot(writer io.Writer) error {
	return nil
}

func (i *Index) Restore(writer io.Writer) error {
	return nil
}

// match matches the pattern to the key
func (idx *Index) match(key Key) bool {
	return key.Match(idx.pattern)
}

// clearCopy creates a copy of the idx, but with an empty dataset.
func (idx *Index) clearCopy() *Index {
	// copy the idx meta information
	nidx := &Index{
		name:    idx.name,
		pattern: idx.pattern,
		db:      idx.db,
		indexer: idx.indexer,
	}
	switch idx.t {
	case BTree:
		idx.btr = btree.New(btreeDegrees, nidx)
	case RTree:
		idx.rtr = rtree.New(nidx)
	}
	return nidx
}

// rebuild rebuilds the idx
// may need to be invoked from a worker if the data set is large
func (idx *Index) rebuild() {
	switch idx.t {
	case BTree:
		idx.btr = btree.New(btreeDegrees, idx)
	case RTree:
		idx.rtr = rtree.New(idx)
	}
	// iterate through all keys and fill the idx
	idx.db.items.Ascend(func(value btree.Item) bool {
		dbi := value.(*ValueItem)
		if !idx.match(dbi.Key) {
			// does not match the pattern, continue
			return true
		}

		var sk IndexItem
		if len(dbi.Indexes) > 0 {
			for _, sec := range dbi.Indexes {
				secIdx := sec.index()
				if secIdx != nil {
					secIdx.remove(sec)
				}
			}
		}

		switch idx.t {
		case BTree:
			if sk == nil {
				sk = idx.indexer.Index(idx, dbi)
				if sk != nil {
					dbi.Indexes = append(dbi.Indexes, sk)
				}
			} else {
				idx.btr.Delete(sk)
			}

			if sk != nil {
				idx.btr.ReplaceOrInsert(sk)
			} else {
				// Ignored.
			}

		case RTree:
			if sk == nil {
				sk = idx.indexer.Index(idx, dbi)
				if sk != nil {
					dbi.Indexes = append(dbi.Indexes, sk)
				}
			} else {
				idx.rtr.Remove(sk)
			}

			if sk != nil {
				idx.rtr.Insert(sk)
			} else {
				// Ignored.
			}
		}
		return true
	})
}

// CreateIndex builds a new idx and populates it with items.
// The items are ordered in an b-tree and can be retrieved using the
// Ascend* and Descend* methods.
// An error will occur if an idx with the same name already exists.
//
// When a pattern is provided, the idx will be populated with
// keys that match the specified pattern. This is a very simple pattern
// match where '*' matches on any number characters and '?' matches on
// any one character.
// The less function compares if string 'a' is less than string 'b'.
// It allows for indexes to create custom ordering. It's possible
// that the strings may be textual or binary. It's up to the provided
// less function to handle the content format and comparison.
// There are some default less function that can be used such as
// IndexString, IndexBinary, etc.
func (s *Table) CreateIndex(name, pattern string, indexer Indexer) error {
	return s.createIndex(BTree, name, pattern, indexer)
}

// CreateSpatialIndex builds a new idx and populates it with items.
// The items are organized in an r-tree and can be retrieved using the
// Intersects method.
// An error will occur if an idx with the same name already exists.
//
// The rect function converts a string to a rectangle. The rectangle is
// represented by two arrays, min and max. Both arrays may have a length
// between 1 and 20, and both arrays must match in length. A length of 1 is a
// one dimensional rectangle, and a length of 4 is a four dimension rectangle.
// There is support for up to 20 dimensions.
// The values of min must be less than the values of max at the same dimension.
// Thus min[0] must be less-than-or-equal-to max[0].
// The IndexRect is a default function that can be used for the rect
// parameter.
func (s *Table) CreateSpatialIndex(name, pattern string, indexer Indexer) error {
	return s.createIndex(RTree, name, pattern, indexer)
}

// createIndex is called by CreateIndex() and CreateSpatialIndex()
func (s *Table) createIndex(
	idxType IndexType,
	name string,
	pattern string,
	indexer Indexer,
) error {
	if name == "" {
		// cannot create an idx without a name.
		// an empty name idx is designated for the main "keys" tree.
		return moved.ErrIndexExists
	}
	// check if an idx with that name already exists.
	if _, ok := s.idxs[name]; ok {
		// idx with name already exists. error.
		return moved.ErrIndexExists
	}

	// intialize new idx
	idx := &Index{
		t:       idxType,
		name:    name,
		pattern: pattern,
		db:      s,
		indexer: indexer,
	}

	// save the idx
	s.insertIndex(idx)

	idx.rebuild()
	return nil
}

// DropIndex removes an idx.
func (s *Table) DropIndex(name string) error {
	if name == "" {
		// cannot drop the default "keys" idx
		return moved.ErrInvalidOperation
	}
	idx, ok := s.idxs[name]
	if !ok {
		return moved.ErrNotFound
	}

	s.removeIndex(idx)

	return nil
}
