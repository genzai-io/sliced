package btrdb

import (
	"fmt"
	"sync"
)

type Schema struct {
	sync.Mutex

	nameMap map[string]*TableIndex

	Tables []*Table
	Sealed bool
	Errors []error
}

func (s *Schema) Add(tables ...*Table) error {
	s.Lock()
	defer s.Unlock()

	if s.Sealed {
		return ErrSealed
	}

	if len(tables) > 0 {
		for _, table := range tables {
			s.Tables = append(s.Tables, table)
		}
	}

	return nil
}

func (s *Schema) addIndex(index *TableIndex) error {
	if s.nameMap == nil {
		s.nameMap = make(map[string]*TableIndex)
	}

	_, ok := s.nameMap[index.Name]
	if ok {
		return fmt.Errorf("index name '%s' was already used", index.Name)
	}
	s.nameMap[index.Name] = index

	return nil
}

func (s *Schema) Create(tx *Tx) error {
	s.Lock()
	defer s.Unlock()

	if len(s.Tables) == 0 {
		return ErrEmptySchema
	}

	for _, table := range s.Tables {
		if err := s.addIndex(table.Pk); err != nil {
			return err
		}
		if len(table.Secondary) > 0 {
			for _, index := range table.Secondary {
				if err := s.addIndex(index); err != nil {
					return err
				}
			}
		}

		if err := table.Build(tx); err != nil {
			s.Errors = append(s.Errors, err)
		}
	}

	s.Sealed = true

	return nil
}
