package record

import (
	"errors"
	"time"

	"github.com/genzai-io/sliced/proto/store"
)

var (
	ErrExpectedHeader = errors.New("expected header")
	ErrExpectedEnd    = errors.New("expected end byte")
	ErrParseEpoch     = errors.New("parse epoch")
	ErrParseSeq       = errors.New("parse seq")
	ErrParseLogID     = errors.New("parse log id")
	ErrParseSlot      = errors.New("parse slot")
	ErrParseLength    = errors.New("parse length")
	ErrParseBody      = errors.New("parse body")
	ErrParseEnd       = errors.New("parse end")
	ErrNilRecordID    = errors.New("nil record id")

	ErrExpectedHeaderOrEnd = errors.New("expected header or end")
	ErrNilEntry            = errors.New("nil entry")
	ErrBadHeader           = errors.New("bad header")
	ErrCorrupted           = errors.New("corrupted")
	ErrWriteOutOfSpace     = errors.New("write out of space")
	ErrStop                = errors.New("stop")
)

const (
	Magic  = uint32(0xBAEEBAEF)
	Header = byte('[')
	End    = byte('\n')
)

type IDFactory struct {
	// Last ID used
	last store.RecordID
}

func NewIDFactory() *IDFactory {
	return &IDFactory{last: store.RecordID{uint64(time.Now().UnixNano() / 1000000), 0}}
}

func (i *IDFactory) Next() store.RecordID {
	ts := uint64(time.Now().UnixNano() / 1000000)
	if ts > i.last.Epoch {
		return store.RecordID{ts, 0}
	} else {
		return store.RecordID{i.last.Epoch, i.last.Seq + 1}
	}
}

func (i *IDFactory) NextSequence() store.RecordID {
	i.last = store.RecordID{i.last.Epoch, i.last.Seq + 1}
	return i.last
}

func IsEqual(from, to store.RecordID) bool {
	return from == to
}

func IsLess(from, to store.RecordID) bool {
	if from.Epoch < to.Epoch {
		return true
	} else if from.Epoch > to.Epoch {
		return false
	} else {
		return from.Seq < to.Seq
	}
}

func IsGreater(from, to store.RecordID) bool {
	if from.Epoch > to.Epoch {
		return true
	} else if from.Epoch < to.Epoch {
		return false
	} else {
		return from.Seq > to.Seq
	}
}

func IsGreaterOrEqual(from, to store.RecordID) bool {
	if from.Epoch > to.Epoch {
		return true
	} else if from.Epoch < to.Epoch {
		return false
	} else {
		return from.Seq >= to.Seq
	}
}

// Record in an excerpt file
type Record struct {
	ID    store.RecordID
	LogID uint64
	Slot  uint16
	Data  []byte
}

//
type Entry struct {
	Record

	pos       int64
	hsize     int
	bsize     int
	remaining int
	read      func(buf []byte) (int, error)
}

func (c *Entry) BodyPos() int64 {
	return c.pos + int64(c.hsize)
}

func (c *Entry) Read(buf []byte) (int, error) {
	if c.read == nil {
		return 0, nil
	}
	return c.read(buf)
}

func (c *Entry) Remaining() int {
	return c.remaining
}
