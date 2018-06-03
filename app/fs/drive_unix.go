package fs

import (
	"syscall"

	"github.com/slice-d/genzai/proto/store"
)

func Statfs(path string) (*store.DriveStats, error) {
	var stats syscall.Statfs_t

	if err := syscall.Statfs(path, &stats); err != nil {
		return nil, err
	}

	return &store.DriveStats{
		Size_: stats.Blocks * uint64(stats.Bsize),
		Avail: stats.Bavail * uint64(stats.Bsize),
		Used: (stats.Blocks * uint64(stats.Bsize)) - (stats.Bavail * uint64(stats.Bsize)),
	}, nil
}
