package record

import (
	"bufio"
	"errors"
	"io"

	"github.com/genzai-io/sliced/proto/store"
)

// Iterates through a Segment formatted file using a non-mapped buffer.
type Iterator struct {
	reader *bufio.Reader

	code      byte
	epoch     int64
	timestamp int64
	logID     int64
	entry     Entry
	pos       int64

	data []byte

	err error
}

// Creates a new Iterator from a io.MappedReader
func NewIterator(reader *bufio.Reader) (*Iterator, error) {
	iterator := &Iterator{
		reader: reader,
	}

	// Read the header
	iterator.err = iterator.readHeader()
	if iterator.err != nil {
		return nil, iterator.err
	}

	return iterator, nil
}

// Creates a new Iterator from a io.MappedReader
func NewIteratorWithHeader(header *store.SegmentHeader, reader *bufio.Reader, data []byte) (*Iterator, error) {
	// Allocate iterator
	iterator := &Iterator{
		reader: reader,
	}

	// Set the header and skip needing to physically read it
	iterator.logID = int64(header.LogID)
	iterator.timestamp = int64(header.Timestamp)
	iterator.pos = header.StartIndex

	iterator.data = data

	return iterator, nil
}

// Reads a header starting at the first header byte until it reaches an "End" byte
func (it *Iterator) readHeader() error {
	var size int
	var hlen uint64
	size, hlen, it.err = ReadUvarint(it.reader)

	if it.err != nil {
		return it.err
	}

	hdr := make([]byte, hlen)
	_, it.err = it.reader.Read(hdr)
	if it.err != nil {
		return it.err
	}

	// Parse header
	header := store.SegmentHeader{}
	it.err = header.Unmarshal(hdr)
	if it.err != nil {
		return it.err
	}
	// Build header
	it.timestamp = int64(header.Timestamp)
	it.logID = int64(header.LogID)

	// Read 'End' byte
	it.code, it.err = it.reader.ReadByte()
	if it.err != nil {
		return it.err
	}
	if it.code != End {
		return ErrExpectedEnd
	}

	// Increase position to right past 'End' byte
	it.pos += int64(size+len(hdr)) + 1

	return nil
}

// Reads the next Entry and returns nil for Entry if EOF occurs.
func (it *Iterator) Next() (*Entry, error) {
	var length uint64
	var size int
	var entry = &it.entry
	var reader = it.reader

	// Discard unread from previous Entry
	if entry.remaining > 0 {
		// Skip to next entry
		_, it.err = reader.Discard(entry.remaining)

		if it.err != nil {
			return nil, it.err
		}
	}

	entry.pos = it.pos

	// Reset headerSize
	entry.hsize = 0

	// Read epoch offset
	size, it.epoch, it.err = ReadVarint(reader)
	if it.err != nil {
		if it.err == io.EOF {
			it.err = nil
			return nil, nil
		}
		return nil, it.err
	}
	// Adjust
	entry.ID.Epoch = uint64(it.timestamp + it.epoch)
	entry.hsize += size

	// Read ID seq
	size, entry.ID.Seq, it.err = ReadUvarint(reader)
	if it.err != nil {
		return nil, it.err
	}
	entry.hsize += size

	// Read Log ID
	size, it.epoch, it.err = ReadVarint(reader)
	if it.err != nil {
		return nil, it.err
	}
	// Adjust LogID
	entry.LogID = uint64(it.logID + it.epoch)
	entry.hsize += size

	// Read slot
	entry.Slot, it.err = ReadUint16(reader)
	if it.err != nil {
		return nil, it.err
	}
	entry.hsize += 2

	// Read body length
	size, length, it.err = ReadUvarint(reader)
	if it.err != nil {
		return nil, it.err
	}
	entry.hsize += size

	// Setup entry for reading the body
	entry.bsize = int(length)
	entry.remaining = entry.bsize

	// Move the position to the next entry
	it.pos += int64(entry.hsize + int(length))
	entry.pos += int64(entry.hsize)

	// Read next code
	it.code, it.err = reader.ReadByte()
	if it.err != nil {
		if it.err == io.EOF {
			it.err = nil
			return nil, nil
		}
		return nil, it.err
	}

	// Move position up 1
	it.pos++

	if it.code != End {
		return nil, ErrExpectedEnd
	}

	return entry, nil
}

var ErrOverflow = errors.New("binary: varint overflows a 64-bit integer")

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

// ReadUvarint reads an encoded unsigned integer from r and returns it as a uint64.
func ReadUvarint(r io.ByteReader) (int, uint64, error) {
	var x uint64
	var s uint
	for i := 0; ; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return i + 1, x, err
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return i + 1, x, ErrOverflow
			}
			return i + 1, x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}

// ReadVarint reads an encoded signed integer from r and returns it as an int64.
func ReadVarint(r io.ByteReader) (int, int64, error) {
	l, ux, err := ReadUvarint(r) // ok to continue in presence of error
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return l, x, err
}

func Uvarint(buf []byte) (uint64, int) {
	var x uint64
	var s uint
	for i, b := range buf {
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0, -(i + 1) // overflow
			}
			return x | uint64(b)<<s, i + 1
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, 0
}

// Varint decodes an int64 from buf and returns that value and the
// number of bytes read (> 0). If an error occurred, the value is 0
// and the number of bytes n is <= 0 with the following meaning:
//
// 	n == 0: buf too small
// 	n  < 0: value larger than 64 bits (overflow)
// 	        and -n is the number of bytes read
//
func Varint(buf []byte) (int64, int) {
	ux, n := Uvarint(buf) // ok to continue in presence of error
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, n
}
