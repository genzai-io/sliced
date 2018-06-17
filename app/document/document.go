package document

import (
	"strconv"

	"github.com/genzai-io/sliced/proto/store"
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

}
