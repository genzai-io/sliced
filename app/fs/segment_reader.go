package fs

import (
	"os"
	"sync"

	"github.com/slice-d/genzai/app/mmap"
	"github.com/slice-d/genzai/app/record"
	"github.com/slice-d/genzai/proto/store"
)

// Optimized for immutable segment files.
type SegmentReader struct {
	sync.RWMutex

	model   *store.Segment
	file    *os.File
	closed  bool
	b       mmap.MMap
	count   int64
	readers map[int64]*record.MMapReader
}

func NewSegmentReader(model *store.Segment, name string, size int64, mode os.FileMode) (*SegmentReader, error) {
	file, err := os.OpenFile(name, os.O_RDONLY, mode)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	_ = info

	// We advise sequential access since the files are logs.
	b, err := mmap.MapRegion(file, int(size), mmap.RDONLY, mmap.SEQUENTIAL, 0)
	if err != nil {
		file.Close()
		return nil, err
	}

	r := &SegmentReader{
		b:       b,
		readers: make(map[int64]*record.MMapReader),
	}

	return r, nil
}

func (r *SegmentReader) ReadAt(b []byte, pos int64) (int, error) {
	r.RLock()

	r.RUnlock()

	return 0, nil
}

func (r *SegmentReader) Cursor() (*record.MMapReader, error) {
	r.Lock()
	defer r.Unlock()

	if r.closed {
		return nil, os.ErrClosed
	}

	r.count++
	id := r.count
	reader := &record.MMapReader{
		RWMutex: &r.RWMutex,
		B:       r.b,
		I:       0,
	}
	r.readers[id] = reader

	return reader, nil
}

func (r *SegmentReader) Close() error {
	r.Lock()
	defer r.Unlock()

	if r.closed {
		return os.ErrClosed
	}

	if len(r.readers) > 0 {
		return ErrHasReaders
	}

	return nil
}
