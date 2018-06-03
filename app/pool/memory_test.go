package pool

import (
	"fmt"
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/pbnjay/memory"
)

func Test_MaxMemory(t *testing.T) {
	fmt.Printf("System Memory %s\n", humanize.IBytes(memory.TotalMemory()))
}
