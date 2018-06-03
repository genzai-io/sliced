package table

import (
	"unsafe"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/common/gjson"
	"github.com/genzai-io/sliced/common/sjson"
)

type DocumentFlags int

const (
	Compressed DocumentFlags = iota << 1
	LZ4X
	GZIPX
	Dirty
	Header      // Fast field updates
)

type DocumentContext struct {
}

//
//
type Document interface {
	Get(path string)

	Set(path string, value interface{})
}

//
type DocumentValue struct {
	Flags DocumentFlags
	Size  uint32
	Value []byte // Use a []byte for mutability
}

//
type JSONValue struct {
	DocumentValue
}

func (s JSONValue) Type() moved.DataType {
	return moved.JSON
}

//
func (j JSONValue) Get(path string) {
	if j.Flags&Header != 0 {
		// Check in header first.

	}

	if j.Flags&LZ4X != 0 {
		// Uncompress
	} else if j.Flags&GZIPX != 0 {
		// Uncompress
	}
	result := gjson.Get(*(*string)(unsafe.Pointer(&j.Value)), path)
	_ = result

	if j.Flags&LZ4X != 0 {
		// Compress
	} else if j.Flags&GZIPX != 0 {
		// Compress
	}
}

func (j JSONValue) Set(path string, value interface{}) {
	sjson.SetBytes(j.Value, path, value)
}

type DocumentString []byte

func (b DocumentString) getFromHeader(path string) (offset byte, length byte) {
	hlen := b[0]
	plen := byte(len(path))

	if hlen == 0 {
		return
	}

	var nameIndex byte
	var nameSize byte
	var valIndex byte
	var valSize byte
	var idx byte

LOOP:
	for offset := byte(1); offset < hlen; {
		nameIndex = offset + 1
		nameSize = b[offset]
		valSize = b[offset+1+nameSize]
		valIndex = offset + 2 + nameSize

		// Should we do a fast skip?
		if path[0] != b[nameIndex] || nameSize != plen {
			offset = valIndex + valSize
		} else {
			nameIndex += 1
			// Same lengths and same first character
			for idx = 1; idx < plen; nameIndex++ {
				if b[nameIndex] != path[idx] {
					offset = valIndex + valSize
					continue LOOP
				}
				idx++
			}

			// We have a match!
			return valIndex, valSize
		}
	}
	return 0, 0
}

func (b DocumentString) putInHeader(path string, val Key) {
	offset, length := b.getFromHeader(path)
	if offset == 0 {
		return
	}

	_ = length
}
