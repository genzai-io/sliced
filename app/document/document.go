package document

import (
	"strconv"

	"github.com/genzai-io/sliced/app/table"
	"github.com/genzai-io/sliced/proto/store"
)

type Type int

const (
	JSON     Type = iota // Body is in JSON format
	Protobuf             // Body is in Protocol buffers binary format
	Hash                 // Body is a Go map
	Struct               // Body is a Go struct
	Blob                 // Body is just a []byte
)

type GlobalID struct {
	DatabaseID uint32
	Epoch      uint64
	Slot       uint16
	Seq        uint32
}

type ID struct {
	Epoch uint64
	Slot  uint16
	Seq   uint32
}

type SecondaryValue struct {
	ID      ID
	LogID   uint64
	Key     table.Key
	Pointer table.Key
}

func (id ID) String() string {
	return strconv.FormatUint(id.Epoch, 10) +
		"." + strconv.FormatUint(uint64(id.Slot), 10) +
		"." + strconv.FormatUint(uint64(id.Seq), 10)
}

type RecordHeader struct {
	ID    store.RecordID
	Slot  uint16
	LogID uint64
}

type Record struct {
	RecordHeader

	Indexes  []Projector
	Document Document
}

type Document interface {
	Type() Type

	Bytes() []byte

	String() string
}

//
type VisitorContext struct {
	Number int
	Name   string
}

//
type FieldVisitor interface {
	Begin()

	End()
}
