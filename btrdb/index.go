package btrdb

import (
	"strings"

)

type Projector func(val interface{}) interface{}

type TableIndex struct {
	Name  string
	Error error
	table *Table

	projector Projector
	less      func(a, b string) bool
	pk        bool
	pattern   string
	prefix    string
}

func newPKIndex(name string, projector Projector) *TableIndex {
	name = strings.ToLower(strings.TrimSpace(name))
	return &TableIndex{
		Name:      name,
		projector: projector,
		pk:        true,

		prefix:  name + ":pk:",
		pattern: name + ":pk:*",
	}
}

func NewTableIndex(name string, projector Projector) *TableIndex {
	return &TableIndex{
		Name:      strings.ToLower(strings.TrimSpace(name)),
		projector: projector,
	}
}

// Iterating in primary key ascending order
func (ti *TableIndex) Ascend(tx *Tx, iterator func(key string, value Serializable) bool) error {
	if ti.pk {
		return tx.Ascend(ti.Name, func(key, value string) bool {
			document := ti.table.Factory()
			document.Unmarshal([]byte(value))
			return iterator(key, document)
		})
	} else {
		var badKeys []string = nil
		return tx.Ascend(ti.Name, func(key, value string) bool {
			document, err := ti.table.GetByKey(tx, value)
			if err != nil {
				badKeys = append(badKeys, key)
				return true
			}
			return iterator(key, document)
		})
	}
}

// Iterating in primary key descending order
func (ti *TableIndex) Descend(tx *Tx, iterator func(key string, value Serializable) bool) error {
	if ti.pk {
		return tx.Descend(ti.Name, func(key, value string) bool {
			document := ti.table.Factory()
			document.Unmarshal([]byte(value))
			return iterator(key, document)
		})
	} else {
		var badKeys []string = nil
		return tx.Descend(ti.Name, func(key, value string) bool {
			document, err := ti.table.GetByKey(tx, value)
			if err != nil {
				badKeys = append(badKeys, key)
				return true
			}
			return iterator(key, document)
		})
	}
}

//
func (ti *TableIndex) Build(tx *Tx) error {
	ti.Error = ti.Create(tx)
	return ti.Error
}

// Returns true if the index exists. Otherwise, false.
func (ti *TableIndex) Exists(tx *Tx) bool {
	return ti.Create(tx) == ErrIndexExists
}

//
func (ti *TableIndex) Create(tx *Tx) error {
	return tx.CreateIndex(ti.Name, ti.pattern, IndexBinary)
}

func (ti *TableIndex) Format(val interface{}) string {
	f := Format(val)
	if len(f) == 0 {
		return ""
	}
	return ti.prefix + f
}

//
//func (s *TableIndex) Format(key string) string {
//	return fmt.Sprintf(s.Fmt, key)
//}

//
//
//
func (ti *TableIndex) Extract(document Serializable) string {
	value := ti.projector(document)
	formatted := Format(value)
	if len(formatted) == 0 {
		return ""
	}
	return ti.prefix + formatted
}

//
//
//
func (ti *TableIndex) First(tx *Tx) (exists bool, key string, value string, err error) {
	exists = false
	key = ""
	value = ""
	err = tx.Ascend(ti.Name, func(k, v string) bool {
		exists = true
		key = k
		value = v
		return false
	})
	return
}

//
func (ti *TableIndex) Last(tx *Tx) (exists bool, key string, value string, err error) {
	exists = false
	key = ""
	value = ""
	err = tx.Descend(ti.Name, func(k, v string) bool {
		exists = true
		key = k
		value = v
		return false
	})
	return
}
