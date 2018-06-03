package fs

import (
	"errors"
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


	ErrFilled     = errors.New("filled")
	ErrEmptyWrite = errors.New("empty write")
)
