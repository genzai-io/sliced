package fs

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/mmap"
	"github.com/genzai-io/sliced/app/path"
	"github.com/genzai-io/sliced/app/pool/pbufio"
	"github.com/genzai-io/sliced/app/record"
	"github.com/genzai-io/sliced/proto/store"
)

var (
	ErrTruncate   = errors.New("truncate")
	ErrHasReaders = errors.New("has readers")
)

const (
	maxRecordHeaderBytes = binary.MaxVarintLen64*4 + 8
)

// Append-only optimized Topic Segment writer that continually extends a file
// by calling Truncate and mmap'ing the entire size of the file.
type SegmentWriter struct {
	sync.RWMutex

	// Drive file belongs to
	volume *Drive
	// Name of the file
	name string
	// OS file handle
	file *os.File

	// Flags
	closed bool // Flag to determine whether the file is closed
	err    []error
	fatal  []error

	// Positions
	writePos    int64 // The position of the next byte to be written
	syncPos     int64 // Logical length including non-synced
	truncatePos int64 // The threshold of when the writePos exceeds to fire off a truncate

	// The min position of all known tailers
	// Any regions below this offset can be remapped under the read-only b region
	tailerPos int64
	// Position to trigger a memmap
	remapPos int64

	// The expected size hint on creation
	expsz int64
	// Physical file size
	filesz int64

	// The amount of bytes to grow the physical file
	// This must be region aligned
	growsize   int64
	growthRate int64 // Bytes / second

	lastgrow  time.Time
	extending bool // A flag to determine whether next is being built

	// Stats
	extcount  int64
	extdur    time.Duration
	mmapcount int64
	mmapdur   time.Duration

	// mmap buffer protected by write lock
	b mmap.MMap

	// Header buffer
	muReaders sync.Mutex
	rcounter  int64
	readers   map[int64]*record.MMapReader
	hbuf      [maxRecordHeaderBytes]byte
	header    store.SegmentHeader
	stats     store.SegmentStats
}

func open2(
	volume *Drive,
	name string,
	expectedSize int64,
	header *store.SegmentHeader,
	stats *store.SegmentStats,
	mode os.FileMode,
) (
	*SegmentWriter,
	error,
) {

	// Open file
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, mode)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	_ = info

	// Are we expecting a file with records?
	if stats.Count > 0 {

	}

	return nil, nil
}

// Open a file for writing that was not closed cleanly. Likely of a result
// of a process or OS crash or kernel panic. We'll iterate until the last valid
// record and truncate to that record, then open the file as usual.
func recoverSegment(
	volume *Drive,
	name string,
	mode os.FileMode,
) (
	*SegmentWriter,
	error,
) {

	// Open file
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, mode)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	//size := info.Size()

	// Create a buffered reader of the file
	reader := pbufio.GetReader(file, 65536)
	defer pbufio.PutReader(reader)

	// Create an iterator using the buffered reader
	iterator, err := record.NewIterator(reader)
	if err != nil {
		return nil, err
	}

	_ = info

	var entry *record.Entry
	// Iterate until the end
	for entry, err = iterator.Next(); err != nil; {
		_ = entry
		_ = err
	}

	if err != nil {
		return nil, err
	}

	return createSegmentWriter(volume, name, 0, mode)
}

//
func openSegmentWriter(volume *Drive, segment *store.Segment, mode os.FileMode) (*SegmentWriter, error) {
	name := path.SegmentPath(segment)

	info, err := os.Stat(name)
	if err != nil {
		// Create segment if needed
		if err == os.ErrNotExist {
			return createSegmentWriter(volume, name, 0, mode)
		}

		return nil, err
	}

	expectedSize := segment.Stats.Size_
	fsize := info.Size()

	if expectedSize < fsize {

	}

	return nil, nil
}

//
//
//
func createSegmentWriter(volume *Drive, name string, expectedSize int64, mode os.FileMode) (*SegmentWriter, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, mode)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	if expectedSize > 0 {
		// Page align
		expectedSize = (expectedSize/MapSize1 + 1) * MapSize1
	}

	initialSize := expectedSize

	if initialSize < MapSize1 {
		initialSize = MapSize1
	}

	growSize := MapSize1
	fsize := info.Size()

	if fsize > initialSize {
		if fsize <= MapSize1 {
			initialSize = MapSize2
			growSize = MapSize2
		} else if fsize <= MapSize2 {
			initialSize = MapSize3
			growSize = MapSize3
		} else if fsize <= MapSize3 {
			initialSize = MapSize4
			growSize = MapSize4
		} else if fsize <= MapSize4 {
			initialSize = MapSize5
			growSize = MapSize5
		} else {
			initialSize = MapSize6 * (fsize/MapSize6 + 1)
			growSize = MapSize5
		}
	}

	if fsize > initialSize {
		initialSize = MapSize6 * (fsize/MapSize6 + 1)
	}

	if fsize != initialSize {
		err = file.Truncate(initialSize)
		if err != nil {
			file.Close()
			return nil, err
		}
	}

	// First mapping
	b, err := mmap.MapRegion(file, int(initialSize), mmap.RDWR, mmap.SEQUENTIAL, 0)
	if err != nil {
		file.Close()
		return nil, err
	}

	w := &SegmentWriter{
		name:      name,
		volume:    volume,
		file:      file,
		expsz:     expectedSize,
		filesz:    initialSize,
		extending: false,

		b:           b,
		writePos:    0,
		growsize:    growSize,
		truncatePos: initialSize - (growSize / 2),
		lastgrow:    time.Now(),
	}

	// Do we need to iterate through the file to find the end position
	if info.Size() > 0 {

	}

	return w, nil
}

//
//
//
func (f *SegmentWriter) Close() (err error) {
	f.Lock()
	if !f.closed {
		f.closed = true

		if f.writePos < f.filesz {
			f.volume.Logger.Debug().Msgf("%s: truncating file down to exact size from %s to %s", f.name, humanize.IBytes(uint64(f.filesz)), humanize.IBytes(uint64(f.writePos)))
			err = f.file.Truncate(f.writePos)
		}

		// Sync if needed
		if f.syncPos < f.writePos {
			err = mmap.Fdatasync(f.file)
		}

		// Close all the readers
		f.muReaders.Lock()
		for k, v := range f.readers {
			v.Closed = true
			delete(f.readers, k)
		}
		f.muReaders.Unlock()
	}
	f.Unlock()

	return nil
}

// Append a single record
func (f *SegmentWriter) append(r *record.Record) (int, error) {
	// Ignore nil records
	if r == nil {
		return 0, nil
	}

	var expect int
	var hsize = 0

	hbuf := f.hbuf

	// Write timestamp
	ts := int64(r.ID.Epoch) - int64(f.header.Timestamp)
	expect = binary.PutVarint(hbuf[:], ts)
	hsize += expect

	// Write Seq
	expect = binary.PutUvarint(hbuf[hsize:], r.ID.Seq)
	hsize += expect

	// Write LogID
	logID := int64(r.LogID) - int64(f.header.LogID)
	expect = binary.PutVarint(hbuf[hsize:], logID)
	hsize += expect

	// Write Slot
	hbuf[hsize] = byte(r.Slot)
	hsize++
	hbuf[hsize] = byte(r.Slot >> 8)
	hsize++

	// Write body size
	bsize := len(r.Data)
	expect = binary.PutUvarint(hbuf[hsize:], uint64(bsize))
	hsize += expect

	// Is there a payload?
	if bsize > 0 {
		// Increase the body size
		f.stats.Body += uint64(bsize)
	}

	marked := f.writePos

	// Write header
	err := f.write(hbuf[:hsize])
	if err != nil {
		f.writePos = marked
		f.Unlock()
		return 0, err
	}

	// Write body
	bodyPos := f.writePos
	err = f.write(r.Data)
	if err != nil {
		f.writePos = marked
		f.Unlock()
		return 0, err
	}

	// Write terminator
	err = f.writeByte(record.End)
	if err != nil {
		f.writePos = marked
		f.Unlock()
		return 0, err
	}

	// Increase header size
	f.stats.Header += uint64(hsize)

	// Update First and last
	if f.stats.First == nil {
		// Set first r
		f.stats.First = &store.RecordPointer{
			Id:    &r.ID,
			LogID: r.LogID,
			Size_: uint32(bsize),
			Slot:  uint32(r.Slot),
			Pos:   int64(bodyPos),
		}
		// Create a new instance for last
		f.stats.Last = &store.RecordPointer{
			Id:    &r.ID,
			LogID: r.LogID,
			Size_: f.stats.First.Size_,
			Slot:  f.stats.First.Slot,
			Pos:   f.stats.First.Pos,
		}
	} else {
		f.stats.Last.Id = &r.ID
		f.stats.Last.LogID = r.LogID
		f.stats.Last.Size_ = uint32(bsize)
		f.stats.Last.Slot = uint32(r.Slot)
		f.stats.Last.Pos = int64(bodyPos)
	}

	if uint32(bsize) > f.stats.MaxBody {
		f.stats.MaxBody = uint32(bsize)
	}

	// Increment count
	f.stats.Count++

	return hsize + len(r.Data) + 1, nil
}

//
//
//
func (f *SegmentWriter) Append(record *record.Record) (n int, err error) {
	// Obtain a write lock
	f.Lock()
	// Check if closed
	if f.closed {
		f.Unlock()
		return 0, os.ErrClosed
	}

	n, err = f.append(record)
	if err != nil {
		f.Unlock()
		return 0, err
	}

	// Release write lock
	f.Unlock()

	return
}

//
//
//
func (f *SegmentWriter) writeByte(b byte) (err error) {
	if f.closed {
		f.Unlock()
		return os.ErrClosed
	}

	available := f.filesz - f.writePos
	if available <= 0 {
		f.volume.Logger.Info().Msgf("%s: filled %s", f.name, humanize.IBytes(uint64(f.writePos)))
		f.volume.Logger.Debug().Msgf("%s: extending...", f.name)

		// Apply backpressure to the write by waiting for the truncate to finish
		err = f.truncate(int(1))

		// Did truncate fail?
		if err != nil {
			f.Unlock()

			f.volume.Logger.Error().Msgf("%s: truncate() returned an error", f.name)
			f.volume.Logger.Error().Err(err)
			return
		}

		available = f.filesz - f.writePos
		if available <= 0 {
			f.Unlock()
			return ErrEmptyWrite
		}
	}

	f.b[f.writePos] = b
	f.writePos++

	return nil
}

//
//
//
func (f *SegmentWriter) write(p []byte) (error) {
	if f.closed {
		return os.ErrClosed
	}

	remaining := int64(len(p))
	var size int64
	var wrote int64
	var n int

	for len(p) > 0 {
		size = f.filesz
		available := size - f.writePos

		if remaining > available {
			dst := f.b[f.writePos:]
			copy(dst, p[:available])
			wrote = available
		} else {
			copy(f.b[f.writePos:f.writePos+remaining], p)
			wrote = remaining
		}

		// Did any write occur?
		if wrote == 0 {
			return ErrEmptyWrite
		}

		// Increase position by amount wrote
		f.writePos += wrote
		n += int(wrote)

		// Deduct remaining
		remaining -= wrote

		// Did we fill up the current tail?
		if remaining > 0 {
			f.volume.Logger.Info().Msgf("%s: filled %s", f.name, humanize.IBytes(uint64(f.writePos)))
			f.volume.Logger.Debug().Msgf("%s: extending...", f.name)

			// Apply backpressure to the write by waiting for the truncate to finish
			err := f.truncate(int(remaining))

			// Did truncate fail?
			if err != nil {
				f.Unlock()

				f.volume.Logger.Error().Msgf("%s: truncate() returned an error", f.name)
				f.volume.Logger.Error().Err(err)
				return err
			}
		} else {
			// Exit loop
			return nil
		}

		// Slice remaining bytes
		p = p[wrote:]
	}

	return nil
}

//
//
//
func (f *SegmentWriter) Write(p []byte) (int, error) {
	// obtain a write lock
	// all writes are serialized
	f.Lock()

	if f.closed {
		f.Unlock()
		return 0, os.ErrClosed
	}

	err := f.write(p)

	// release write lock
	f.Unlock()

	if err != nil {
		return 0, err
	} else {
		return len(p), nil
	}
}

//
//
//
func (f *SegmentWriter) sync() error {
	f.RLock()
	syncPos := f.syncPos
	writePos := f.writePos
	f.RUnlock()

	if syncPos < writePos {
		if err := mmap.Fdatasync(f.file); err != nil {
			f.volume.Logger.Error().Msgf("%s: file.Sync() error")
			f.volume.Logger.Error().Err(err)

			f.Lock()
			f.err = append(f.err, err)
			f.Unlock()
		} else {
			f.Lock()
			f.syncPos = writePos
			f.Unlock()
		}
	}

	return nil
}

//
//
//
func (f *SegmentWriter) Truncate(needed int) error {
	f.Lock()
	defer f.Unlock()
	return f.truncate(needed)
}

//
//
// Truncates a file and memory maps a chunk of the tail of the file
func (f *SegmentWriter) truncate(needed int) error {
	// Was it canceled?
	if f.closed {
		return nil
	}

	if f.writePos < f.truncatePos {
		f.extending = false
		return nil
	}

	// Timing info
	now := time.Now()
	sinceLast := now.Sub(f.lastgrow)

	growsize := MapSize1

	// Let's determine how aggressive we need to be based on current demand.
	// We'll issue fewer truncate and mmap as a result.
	if sinceLast < AggressiveDuration {
		if f.filesz < MapSize1 {
			growsize = MapSize3
		} else if f.filesz < MapSize2 {
			growsize = MapSize4
		} else if f.filesz < MapSize3 {
			growsize = MapSize5
		} else if f.filesz < MapSize4 {
			growsize = MapSize6
		} else if f.filesz < MapSize5 {
			growsize = MapSize7
		} else {
			growsize = MapSize8
		}

		moved.Logger.Warn().Msgf("%s: aggressive growth detected growing by %s", f.name, humanize.IBytes(uint64(growsize)))
	} else {
		if f.filesz < MapSize1 {
			growsize = MapSize2
		} else if f.filesz < MapSize2 {
			growsize = MapSize3
		} else if f.filesz < MapSize3 {
			growsize = MapSize4
		} else if f.filesz < MapSize4 {
			growsize = MapSize5
		} else if f.filesz < MapSize5 {
			growsize = MapSize6
		} else if f.filesz < MapSize6 {
			growsize = MapSize7
		} else if f.filesz < MapSize7 {
			growsize = MapSize8
		} else {
			growsize = MapSize9
		}
		moved.Logger.Warn().Msgf("%s: normal growth detected growing by %s", f.name, humanize.IBytes(uint64(growsize)))
	}

	newsize := (f.filesz/growsize + 1) * growsize

	// Truncate file
	err := f.file.Truncate(newsize)
	if err != nil {
		f.volume.Logger.Error().Msgf("%s: file.Truncate() error", f.name)
		f.volume.Logger.Error().Err(err)
		return err
	}

	f.lastgrow = now
	f.filesz = newsize
	f.growsize = growsize
	f.truncatePos = newsize - (f.growsize / 2)

	f.extcount++
	f.extdur += time.Now().Sub(now)

	// Force a memmap?
	remapped := f.filesz > int64(len(f.b)) // || f.filesz >= f.remapPos
	if remapped {
		f.mmapcount++
		now = time.Now()
		f.volume.Logger.Warn().Msgf("%s: remapping since it grew over the mapped size", f.name)
		f.memmap()
		f.mmapdur += time.Now().Sub(now)
	}

	// Flag next as filled
	f.extending = false

	return nil
}

//
//
//
func (f *SegmentWriter) close(err error) {
	f.closed = true

	f.b = nil

	// Close all the readers
	f.muReaders.Lock()
	for k, v := range f.readers {
		v.Closed = true
		v.B = nil
		delete(f.readers, k)
	}
	f.muReaders.Unlock()
}

//
//
//
func (f *SegmentWriter) memmap() {
	if f.filesz < f.remapPos {
		return
	}

	// We map over the filesz of the file relative to it's filesz
	// while capping a max growth rate.
	mapgrowsize := MapSize1
	if f.filesz < MapSize1 {
		mapgrowsize = MapSize2
	} else if f.filesz < MapSize2 {
		mapgrowsize = MapSize3
	} else if f.filesz < MapSize3 {
		mapgrowsize = MapSize4
	} else if f.filesz < MapSize4 {
		mapgrowsize = MapSize5
	} else if f.filesz < MapSize5 {
		mapgrowsize = MapSize6
	} else if f.filesz < MapSize6 {
		mapgrowsize = MapSize7
	} else if f.filesz < MapSize7 {
		mapgrowsize = MapSize8
	} else {
		mapgrowsize = MapSize9
	}

	mapsize := (f.filesz/mapgrowsize + 1) * mapgrowsize
	remapPos := mapsize - (mapgrowsize / 2)

	// mmap
	b, err := mmap.MapRegion(f.file, int(mapsize), mmap.RDWR, mmap.SEQUENTIAL, 0)

	if err != nil {
		// Let's unmap current first
		err = f.b.Unmap()
		if err != nil {
			f.fatal = append(f.fatal, err)
			f.close(err)
			return
		}

		// Try mmap now
		b, err = mmap.MapRegion(f.file, int(mapsize), mmap.RDWR, mmap.SEQUENTIAL, 0)
		if err != nil {
			f.fatal = append(f.fatal, err)
			f.close(err)
			return
		}

		if err != nil {
			f.err = append(f.err, err)
			f.close(err)
			return
		}

		f.b = b
	} else {
		// Unmap
		err = f.b.Unmap()
		if err != nil {
			f.err = append(f.err, err)
			f.volume.Logger.Error().Msgf("%s: unmap failed but a new mapping succeeded")
			f.volume.Logger.Error().Err(err)
		}

		// Change to new map
		f.b = b

		// Update memmap position
		f.remapPos = remapPos
	}

	// Update mapped buffer on all readers
	// Since it's an append-only file the read index can remain the same
	for _, v := range f.readers {
		v.B = f.b
	}
}

//
//
//
func (f *SegmentWriter) ReadAt(b []byte, pos int64) (int, error) {
	l := int64(len(b))
	if l == 0 {
		return 0, os.ErrInvalid
	}

	// since the file is append-only then we can read up to the current write position
	f.RLock()
	if f.closed {
		f.RUnlock()
		return 0, os.ErrClosed
	}
	size := f.writePos
	expectedSize := l + pos
	if expectedSize > size {
		available := expectedSize - size
		if available > 0 {
			copy(b, f.b[pos:pos+available])
		}
		f.RUnlock()
		return int(available), io.EOF
	} else {
		copy(b, f.b[pos:pos+l])
		f.RUnlock()
		return len(b), nil
	}
}

//
//
//
func (r *SegmentWriter) Size() int64 {
	r.RLock()
	size := r.writePos
	r.RUnlock()
	return size
}

//
//
//
func (f *SegmentWriter) OpenReader() (*record.MMapReader, error) {
	f.RLock()
	f.muReaders.Lock()
	defer f.muReaders.Unlock()
	defer f.RUnlock()

	if f.closed {
		return nil, os.ErrClosed
	}

	reader, err := record.NewMappedReader(&f.RWMutex, &f.header, f.b)
	if err != nil {
		return nil, err
	}

	f.rcounter++
	id := f.rcounter
	f.readers[id] = reader

	return reader, nil
}
