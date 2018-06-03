package pool

import (
	"os"

	"github.com/pbnjay/memory"
	_ "github.com/pbnjay/memory"
)

var MaxMemory = memory.TotalMemory()
var AvailMemory = int64(float64(MaxMemory) * 0.5)

var (
	PageSize  = os.Getpagesize()
	BlockSize = PageSize
)

func init() {
	if PageSize < 65536 {
		BlockSize = 65536
	} else {
		BlockSize = PageSize
	}
}
