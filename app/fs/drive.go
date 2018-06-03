package fs

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/genzai-io/sliced"
	store_pb "github.com/genzai-io/sliced/proto/store"
	"github.com/genzai-io/sliced/common/service"
)

type Drive struct {
	service.BaseService

	model store_pb.Drive

	// Context
	ctx    context.Context
	cancel context.CancelFunc

	// Mutexes
	mu          sync.Mutex
	muTruncate  sync.Mutex
	muSync      sync.Mutex
	muIncubator sync.Mutex
	muFilled    sync.Mutex
	muMap       sync.Mutex
	wg          sync.WaitGroup

	// Channels
	chTruncate chan *SegmentWriter
	chClose    chan *SegmentWriter

	// Memory metrics
	mapped      int64 // Number of bytes memory-mapped
	mappedWrite int64 // Number of bytes memory-mapped for writing
	mappedRead  int64 // Number of bytes memory-mapped for read only
	free        int64

	// Truncate metrics
	truncateOps int64
	truncateDur time.Duration

	// mmap metrics
	mmapOps int64
	mmapDur time.Duration
	// remap mmap metrics
	remapOps int64
	remapDur time.Duration

	// Maximum number of bytes the drive can process per second
	Bandwidth     int64
	BandwidthReal int64

	// Counter of bytes written
	written int64
}

func newDrive(model store_pb.Drive) *Drive {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Drive{
		model:      model,
		ctx:        ctx,
		cancel:     cancel,
		//chTruncate: make(chan *SegmentWriter, 16),
		//chClose:    make(chan *SegmentWriter, 16),
	}

	p.BaseService = *service.NewBaseService(moved.Logger, "drive", p)

	return p
}

func (p *Drive) OnStart() error {
	p.wg.Add(1)
	go p.allocator()

	p.wg.Add(1)
	go p.worker()

	return nil
}

func (p *Drive) OnStop() {
	p.cancel()
	p.wg.Wait()
}

func (d *Drive) Statfs() (*store_pb.DriveStats, error) {
	info, err := Statfs(d.model.Mount)
	if err != nil {
		return nil, err
	}

	d.mu.Lock()
	d.model.Stats = info
	d.mu.Unlock()
	return info, nil
}

func (p *Drive) Create(name string, expectedSize int64, mode os.FileMode) (*SegmentWriter, error) {
	var file *SegmentWriter
	var err error

	//p.muMap.Lock()
	//existing, ok := p.writeMap[name]
	//if ok && existing != nil {
	//	file = existing
	//} else {
	//	file, err = open(p, name, expectedSize, mode)
	//	if err != nil {
	//		p.muMap.Unlock()
	//		return nil, err
	//	}
	//
	//	//file.truncate(0)
	//	p.writeMap[name] = file
	//}
	//p.muMap.Unlock()

	return file, err
}

func (p *Drive) Read(name string, mode os.FileMode) {

}

func (p *Drive) close(file *SegmentWriter) {

}

func (p *Drive) allocator() {
	defer p.wg.Done()

//loop:
//	for {
//		select {
//		// Listen for cancel
//		case <-p.ctx.Done():
//			break loop
//
//		case file, ok := <-p.chTruncate:
//			if ok && file != nil {
//				if err := file.Truncate(0); err != nil {
//					p.Logger.Error().Msgf("%s truncate() error", file.name)
//					p.Logger.Error().Err(err)
//				}
//			}
//		}
//	}
}

func (p *Drive) worker() {
	defer p.wg.Done()

//	set := make([]*SegmentWriter, 0, 256)
//loop:
//	for {
//		select {
//		// Listen for cancel
//		case <-p.ctx.Done():
//			break loop
//
//			// Sync to disk every second
//		case <-time.After(time.Second):
//			set = set[:]
//
//			p.muSync.Lock()
//			// Iterate through list.
//			//for e := p.writelist.Front(); e != nil; e = e.Next() {
//			//	writer, ok := e.Value.(*SegmentWriter)
//			//	if ok {
//			//		set = append(set, writer)
//			//	}
//			//}
//			p.muSync.Unlock()
//
//			for i, file := range set {
//				err := file.sync()
//				set[i] = nil
//
//				if err != nil {
//					p.Logger.Error().Err(err)
//				}
//			}
//		}
//	}
}
