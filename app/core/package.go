package core

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/pbnjay/memory"
)

var (
	regionSpace = memory.TotalMemory() / 2
	regionFree  = regionSpace
	regionUsed  = int64(0)
	regionWrite = int64(0)

	// Page Size
	OSPageSize = int64(os.Getpagesize())
	PageSize   = OSPageSize
	RegionSize = int64(PageSize * 1)

	MapSize1  = RegionSize
	MapSize2  = RegionSize * 8
	MapSize3  = RegionSize * 16
	MapSize4  = RegionSize * 32
	MapSize5  = RegionSize * 64
	MapSize6  = RegionSize * 128
	MapSize7  = RegionSize * 256
	MapSize8  = RegionSize * 512
	MapSize9  = RegionSize * 1024
	MapSize10 = RegionSize * 2048

	AggressiveDuration = time.Millisecond * 100

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

func IsParseError(err error) bool {
	return err == ErrParseEpoch || err == ErrParseSeq || err == ErrParseLogID || err == ErrParseSlot || err == ErrParseLength || err == ErrParseBody || err == ErrParseEnd
}

func init() {
	// Let's pump up the page from 4kb
	if PageSize < 65536 {
		PageSize = 65536
	}
}

func ReadUint16(r io.ByteReader) (uint16, error) {
	var err error
	var b, b2 byte
	b, err = r.ReadByte()
	if err != nil {
		return 0, err
	}
	b2, err = r.ReadByte()
	if err != nil {
		return 0, err
	}

	return uint16(b) | uint16(b2)<<8, nil
}
