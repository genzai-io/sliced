package record

import (
	"encoding/binary"
	"io"
	"os"
	"sync"

	"github.com/genzai-io/sliced/proto/store"
)

// A sequential optimized file reader that memory mmaps an entire file
type MMapReader struct {
	*sync.RWMutex // inherited mutex to protect buffer
	Closed bool   // Closed flag
	Epoch  int64
	LogID  uint64
	B      []byte // current buffer
	I      int64  // current reading index
}

func (r *MMapReader) readHeader() (*store.SegmentHeader, error) {
	var size int
	var hlen uint64
	var b = r.B

	if r.Closed {
		//r.RUnlock()
		return nil, os.ErrClosed
	}

	if r.I >= int64(len(r.B)) {
		//r.RUnlock()
		return nil, io.EOF
	}

	hlen, size = Uvarint(b)
	if size <= 0 {
		//r.RUnlock()
		return nil, io.ErrUnexpectedEOF
	}
	b = b[size:]

	if int(hlen) > len(b) {
		return nil, io.ErrUnexpectedEOF
	}

	hdr := b[:hlen]
	b = b[hlen:]

	// Parse header
	header := &store.SegmentHeader{}
	err := header.Unmarshal(hdr)
	if err != nil {
		return nil, err
	}

	// Build header
	r.Epoch = int64(header.Timestamp)
	r.LogID = header.LogID

	if len(b) < 1 {
		return header, io.ErrUnexpectedEOF
	}

	// Read 'End' byte
	if b[0] != End {
		return header, ErrExpectedEnd
	}

	// Increase position to right past 'End' byte
	r.I += int64(size+len(hdr)) + 1

	return header, nil
}

func (r *MMapReader) ReadAt(ptr *store.RecordPointer, entry *Entry) error {
	if ptr == nil || ptr.Id == nil {
		return ErrNilRecordID
	}
	if ptr == nil || entry == nil {
		return ErrNilEntry
	}
	r.RLock()
	defer r.RUnlock()

	if r.Closed {
		//r.RUnlock()
		return os.ErrClosed
	}

	if ptr.Pos + int64(ptr.Size_) >= int64(len(r.B)) {
		//r.RUnlock()
		return io.EOF
	}

	entry.ID = *ptr.Id
	entry.LogID = ptr.LogID
	entry.Slot = uint16(ptr.Slot)
	entry.bsize = int(ptr.Size_)
	entry.remaining = 0
	// Slice body
	entry.Data = r.B[ptr.Pos:entry.bsize]

	return nil
}

func (r *MMapReader) ReadEntry(entry *Entry) (int, error) {
	r.RLock()
	defer r.RUnlock()

	if r.Closed {
		//r.RUnlock()
		return 0, os.ErrClosed
	}

	if r.I >= int64(len(r.B)) {
		//r.RUnlock()
		return 0, io.EOF
	}

	var (
		length uint64
		size   int
		epoch  int64
		read   = 0
		b      = r.B
	)

	if entry == nil {
		//r.RUnlock()
		return 0, ErrNilEntry
	}

	// Reset headerSize
	read = 0

	// Read epoch offset
	epoch, size = binary.Varint(b)
	if size <= 0 {
		//r.RUnlock()
		return size, ErrParseEpoch
	}
	read += size
	b = b[size:]

	// Adjust
	entry.ID.Epoch = uint64(int64(r.Epoch) + epoch)

	// Read ID seq
	entry.ID.Seq, size = Uvarint(b)
	if size <= 0 {
		//r.RUnlock()
		return read, ErrParseSeq
	}

	read += size
	b = b[size:]

	// Read Log ID
	epoch, size = Varint(b)
	if size <= 0 {
		//r.RUnlock()
		return read, ErrParseLogID
	}

	read += size
	b = b[size:]

	// Adjust LogID
	entry.LogID = uint64(int64(r.LogID) + epoch)

	if len(b) < 2 {
		//r.RUnlock()
		return read, ErrParseSlot
	}

	// Read slot
	entry.Slot = uint16(b[0]) | uint16(b[1])<<8
	read += 2

	// Read body length
	length, size = Uvarint(b)
	if size <= 0 {
		//r.RUnlock()
		return read, ErrParseLength
	}
	read += size

	entry.hsize = read

	// Setup entry for reading the body
	entry.bsize = int(length)
	entry.remaining = 0

	if len(b) < entry.bsize+1 {
		//r.RUnlock()
		read += len(b)
		return read, ErrParseBody
	}

	read += entry.bsize + 1

	// Slice body
	entry.Data = b[:length]

	// Check for End terminator
	if b[length+1] != End {
		//r.RUnlock()
		return read, ErrParseEnd
	}
	//r.RUnlock()

	r.I += int64(read)

	return read, nil
}

// NewMappedReader returns a new MappedReader reading from b.
func NewMappedReader(mu *sync.RWMutex, header *store.SegmentHeader, b []byte) (*MMapReader, error) {
	if header == nil {
		m := &MMapReader{
			RWMutex: mu,
			Closed:  false,
			B:       b,
			I:       0,
			Epoch:   0,
			LogID:   0,
		}
		_, err := m.readHeader()
		if err != nil {
			return nil, err
		}
		return m, nil
	} else {
		m := &MMapReader{
			RWMutex: mu,
			Closed:  false,
			B:       b,
			I:       header.StartIndex,
			Epoch:   int64(header.Timestamp),
			LogID:   header.LogID,
		}

		if header.StartIndex <= 0 {
			m.I = 0
			_, err := m.readHeader()
			if err != nil {
				return nil, err
			}
		}

		return m, nil
	}
}
