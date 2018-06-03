package fs

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/proto/store"
	"github.com/slice-d/genzai/common/service"
)

var (
	ErrInvalidDriveKind = errors.New("invalid drive kind")
	ErrOutOfDiskSpace   = errors.New("out of disk space")

	// Amount of bytes to leave in each drive
	BufferSpace = 1024 * 1024 * 1024 // 1GB

	Drives *DriveService = nil
)

type DriveEvent struct {
	Timestamp time.Time
	Message   string
}

// Manages access to the locally mounted drives / volumes.
type DriveService struct {
	service.BaseService

	mu sync.RWMutex

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	drives map[string]*Drive
	hdd    DriveList
	ssd    DriveList
	nvm    DriveList

	removed []*Drive

	config map[string]*Drive

	// Store a slice of recent events
	events []*DriveEvent
}

func NewDriveService() *DriveService {
	ctx, cancel := context.WithCancel(context.Background())
	ds := &DriveService{
		ctx:    ctx,
		cancel: cancel,
		drives: make(map[string]*Drive),
	}

	ds.BaseService = *service.NewBaseService(moved.Logger, "drives", ds)

	return ds
}

func (d *DriveService) OnStart() error {
	d.syncConfig()

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		for {
			select {
			case <-d.ctx.Done():

			case <-time.After(time.Second * 5):
				d.statfs()
			}
		}
	}()

	return nil
}

func (d *DriveService) statfs() {
	d.mu.RLock()
	for _, drive := range d.drives {
		drive.Statfs()
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()
	sort.Sort(d.hdd)
	sort.Sort(d.ssd)
	sort.Sort(d.nvm)
}

func (d *DriveService) checkConfig() {
}

func (d *DriveService) syncConfig() {
	drives := moved.GetDrives()

	d.mu.Lock()
	defer d.mu.Unlock()

	changes := false

	// Check for removed
	for _, drive := range d.drives {
		if _, ok := drives[strings.ToLower(drive.model.Mount)]; !ok {
			// Drive removed
			d.removed = append(d.removed, drive)
			delete(d.drives, strings.ToLower(drive.model.Mount))
			changes = true

			d.Logger.Warn().Msgf("\"%s\" %s drive removed", drive.model.Mount, drive.model.Kind)
		}
	}

	// Look for new drives
	for _, model := range drives {
		existing, ok := d.drives[strings.ToLower(model.Mount)]
		// Was a new drive added?
		if existing == nil || !ok {
			d.drives[strings.ToLower(model.Mount)] = newDrive(*model)
			changes = true

			d.Logger.Warn().Msgf("\"%s\" %s drive added", model.Mount, model.Kind)
		} else {
			if existing.model.Kind != model.Kind {
				changes = true
				existing.model.Kind = model.Kind
				d.Logger.Warn().Msgf("configuring drive \"%s\" from kind %s to %s", model.Mount, existing.model.Kind, model.Kind)
			}
			if existing.model.Working != model.Working {
				changes = true
				existing.model.Working = model.Working
				d.Logger.Warn().Msgf("configuring drive \"%s\" from kind working=%s to working=%s", model.Mount, existing.model.Working, model.Working)
			}
		}
	}

	if changes {
		d.hdd = d.hdd[:]
		d.ssd = d.ssd[:]
		d.nvm = d.nvm[:]

		for _, drive := range d.drives {
			switch drive.model.Kind {
			case store.Drive_HDD:
				d.hdd = append(d.hdd, drive)
			case store.Drive_SSD:
				d.ssd = append(d.ssd, drive)
			case store.Drive_NVME:
				d.nvm = append(d.nvm, drive)
			}
		}

		sort.Sort(d.hdd)
		sort.Sort(d.ssd)
		sort.Sort(d.nvm)
	}
}

// Picks a drives to put a new file onto.
func (d *DriveService) Pick(kind store.Drive_Kind, size uint64) (*Drive, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	switch kind {
	case store.Drive_HDD:
		if drive := d.pick(d.hdd, size); drive != nil {
			return drive, nil
		}
		if drive := d.pick(d.ssd, size); drive != nil {
			return drive, nil
		}
		if drive := d.pick(d.nvm, size); drive != nil {
			return drive, nil
		}

	case store.Drive_SSD:
		if drive := d.pick(d.ssd, size); drive != nil {
			return drive, nil
		}
		if drive := d.pick(d.nvm, size); drive != nil {
			return drive, nil
		}
		if drive := d.pick(d.hdd, size); drive != nil {
			return drive, nil
		}

	case store.Drive_NVME:
		if drive := d.pick(d.nvm, size); drive != nil {
			return drive, nil
		}
		if drive := d.pick(d.ssd, size); drive != nil {
			return drive, nil
		}
		if drive := d.pick(d.hdd, size); drive != nil {
			return drive, nil
		}

	default:
		return nil, ErrInvalidDriveKind
	}

	return nil, ErrOutOfDiskSpace
}

func (d *DriveService) pick(drives []*Drive, size uint64) *Drive {
	for _, d := range drives {
		if d.model.Stats.Avail > size {
			return d
		}
	}
	return nil
}

type DriveList []*Drive

// Len is the number of elements in the collection.
func (d DriveList) Len() int {
	return len(d)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (d DriveList) Less(i, j int) bool {
	ii := d[i]
	jj := d[j]
	return ii.model.Stats.Used < jj.model.Stats.Used
}

// Swap swaps the elements with indexes i and j.
func (d DriveList) Swap(i, j int) {
	ii := d[i]
	jj := d[j]
	d[i] = jj
	d[j] = ii
}

// We want to evenly fill all available drives.
func (d DriveList) PickNext() {

}
