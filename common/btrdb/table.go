package btrdb

import "strings"

// Simple table abstraction to help manage multiple index entries that are
// tied to the same JSON document.
type Table struct {
	Factory Factory

	Name string
	Pk   *TableIndex
	// Secondary indexes are Unique
	Secondary []*TableIndex

	built bool
	Error error
}

func NewTable(name string, factory Factory, pkProjector Projector, secondary ...*TableIndex) *Table {
	name = strings.ToLower(strings.TrimSpace(name))

	model := factory()
	pk := newPKIndex(name+":pk", pkProjector)
	pk.less = ChooseLess(pkProjector(model))

	t := &Table{
		Name:      name,
		Factory:   factory,
		Pk:        pk,
		Secondary: secondary,
	}

	t.Pk.table = t

	if len(secondary) > 0 {
		for _, index := range secondary {
			index.table = t
			index.Name = name + ":" + index.Name
			index.prefix = index.Name + ":"
			index.pattern = index.prefix + "*"
			index.less = ChooseLess(index.projector(model))
		}
	}

	return t
}

//
func (t *Table) Build(tx *Tx) error {
	if t.Pk == nil {
		t.Error = ErrPKIndexMissing
		return t.Error
	}

	t.Pk.pk = true
	t.Error = t.Pk.Build(tx)
	if t.Error != nil {
		return t.Error
	}

	if len(t.Secondary) > 0 {
		for _, secondary := range t.Secondary {
			t.Error = secondary.Build(tx)
			if t.Error != nil {
				return t.Error
			}
		}
	}

	t.built = true
	return nil
}

// Iterating in primary key ascending order
func (t *Table) Ascend(tx *Tx, iterator func(key string, value Serializable) bool) error {
	if t.Pk == nil {
		t.Error = ErrPKIndexMissing
		return t.Error
	}
	return t.Pk.Ascend(tx, iterator)
}

// Iterating in primary key descending order
func (t *Table) Descend(tx *Tx, iterator func(key string, value Serializable) bool) error {
	if t.Pk == nil {
		t.Error = ErrPKIndexMissing
		return t.Error
	}
	return t.Pk.Descend(tx, iterator)
}

func (t *Table) Get(tx *Tx, id interface{}) (value Serializable, err error) {
	value = nil
	if t.Pk == nil {
		err = ErrPKIndexMissing
		t.Error = err
		return
	}

	return t.GetByKey(tx, t.Pk.Format(id))
}

func (t *Table) GetByKey(tx *Tx, key string) (value Serializable, err error) {
	value = nil
	if t.Pk == nil {
		err = ErrPKIndexMissing
		t.Error = err
		return
	}

	var data string
	data, err = tx.Get(key)
	if err != nil {
		return
	}
	if len(data) == 0 {
		return
	}

	value = t.Factory()
	err = value.Unmarshal([]byte(data))
	if err != nil {
		err = ErrUnMarshaling(err)
	}
	return
}

// For tables that have an auto-incrementing primary key, this will find
// the "next" Primary Key. If the call succeeds, the id can be set on the
// next JSON document to insert.
func (t *Table) NextID(tx *Tx) (id int64, err error) {
	exists, _, data, e := t.Pk.Last(tx)
	if e != nil {
		return 1, e
	}
	if !exists {
		return 1, nil
	}

	// Unmarshal
	document := t.Factory()
	err = document.Unmarshal([]byte(data))
	if err != nil {
		err = ErrUnMarshaling(err)
		return 0, err
	}

	// Get the field value
	field := t.Pk.projector(document)

	// Cast to int64
	id, isNumber := CastInt64(field)
	// Ensure it succeeded
	if !isNumber {
		return 0, ErrPKNotNumber
	}

	return id + 1, nil
}

//
//
//
func (t *Table) Insert(tx *Tx, document Serializable) (error) {
	if t.Pk == nil {
		t.Error = ErrPKIndexMissing
		return t.Error
	}

	key := t.Pk.Extract(document)
	if len(key) == 0 {
		return ErrMissingKey
	}

	existing, err := tx.Get(key)
	if err != nil && err != ErrNotFound {
		return err
	}

	if len(existing) > 0 {
		return ErrDuplicateKey
	}

	var data []byte
	data, err = document.Marshal()
	if err != nil {
		err = ErrMarshaling(err)
		return err
	}

	var previousData string
	previousData, _, err = tx.Set(key, string(data), nil)
	if err != nil {
		return err
	}
	if len(previousData) > 0 {
		return ErrDuplicateKey
	}

	// Insert secondary keys
	if len(t.Secondary) > 0 {
		for _, index := range t.Secondary {
			secondaryKey := index.Extract(document)
			if len(secondaryKey) > 0 {
				existing, err := tx.Get(secondaryKey)
				if err != nil && err != ErrNotFound {
					return err
				}
				if len(existing) > 0 {
					return ErrDuplicateKey
				}

				// Point to primary key which contains the record.
				_, _, err = tx.Set(secondaryKey, key, nil)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

//
//
//
func (t *Table) Update(tx *Tx, document Serializable) (previous Serializable, replaced bool, err error) {
	previous = nil
	replaced = false
	err = nil

	if t.Pk == nil {
		err = ErrPKIndexMissing
		t.Error = err
		return
	}

	if document == nil {
		err = ErrNilDocument
		return
	}

	key := t.Pk.Extract(document)
	if len(key) == 0 {
		err = ErrPkMissing
		return
	}

	var data []byte
	data, err = document.Marshal()
	if err != nil {
		err = ErrMarshaling(err)
		return
	}

	var previousData string
	previousData, replaced, err = tx.Set(key, string(data), nil)
	if err != nil {
		return
	}

	if replaced && len(previousData) > 0 {
		previous = t.Factory()
		err = previous.Unmarshal([]byte(previousData))
		if err != nil {
			err = ErrMarshaling(err)
		}
	}

	if replaced {
		if len(t.Secondary) > 0 {
			for _, index := range t.Secondary {
				newKey := index.Extract(document)
				insertNewKey := len(newKey) > 0

				if previous != nil {
					oldKey := index.Extract(previous)

					if oldKey != newKey && len(oldKey) > 0 {
						_, err = tx.Delete(oldKey)
						if err != nil {
							return
						}
					}
				}

				if insertNewKey {
					// Check if the key already exists.
					var existing string
					existing, err = tx.Get(newKey)
					if err != nil && err != ErrNotFound {
						return
					}
					if len(existing) > 0 {
						err = ErrDuplicateKey
						return
					}

					// Point to primary key which contains the record.
					_, _, err = tx.Set(newKey, key, nil)
					if err != nil {
						return
					}
				}

			}
		}
	}
	return
}

func (t *Table) Delete(tx *Tx, document Serializable) (previous Serializable, err error) {
	previous = nil
	err = nil

	if t.Pk == nil {
		err = ErrPKIndexMissing
		t.Error = err
		return
	}

	key := t.Pk.Extract(document)

	var previousData string
	previousData, err = tx.Delete(key)
	if err != nil {
		return
	}

	if len(previousData) == 0 {
		return nil, ErrNotFound
	}

	previous = t.Factory()
	err = previous.Unmarshal([]byte(previousData))
	if err != nil {
		err = ErrUnMarshaling(err)
	}

	if len(t.Secondary) > 0 {
		for _, index := range t.Secondary {
			oldKey := index.Extract(previous)

			if len(oldKey) > 0 {
				tx.Delete(oldKey)
			}
		}
	}
	return
}
