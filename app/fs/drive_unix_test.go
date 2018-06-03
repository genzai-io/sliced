package fs

import (
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/dustin/go-humanize"
)

func TestStatfs(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	info, err := Statfs(path)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Size: %s\n", humanize.Bytes(info.Size_))
	fmt.Printf("Used: %s\n", humanize.Bytes(info.Used))
	fmt.Printf("Free: %s\n", humanize.Bytes(info.Avail))

	parsed, err := humanize.ParseBytes("4GB")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(parsed)
}

func TestStat(t *testing.T) {
	var stat syscall.Stat_t

	err := syscall.Stat("/", &stat)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(stat.Size)
}
