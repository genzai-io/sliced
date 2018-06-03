package fs

import (
	"fmt"
	"os"
	"testing"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/proto/store"
	"github.com/rs/zerolog"
)

func TestFile(t *testing.T) {
	os.Remove("dat.txt")

	volume := newDrive(store.Drive{})
	err := volume.Start()
	if err != nil {
		volume.Logger.Panic().Err(err)
	}

	file, err := volume.Create("dat.txt", 0, 0755)
	if err != nil {
		volume.Logger.Panic().Err(err)
	}

	wrote, err := file.Write([]byte("hello"))
	if err != nil {
		volume.Logger.Panic().Err(err)
	}

	volume.Logger.Info().Msgf("wrote: %d", wrote)

	for i := 0; i < 100000; i++ {
		//wrote, err = file.Write([]byte("this is something that is decently long to cause overflows"))
		wrote, err = file.Write([]byte("this is something that is decently long to cause overflows"))
		if err != nil {
			volume.Logger.Panic().Err(err)
		}
	}

	slice := string(file.b[file.writePos-100 : file.writePos-1])
	moved.Logger.Info().Msg(slice)

	err = file.Close()
	if err != nil {
		volume.Logger.Panic().Err(err)
	}
}

func BenchmarkFile(b *testing.B) {
	volume := newDrive(store.Drive{})
	err := volume.Start()
	if err != nil {
		volume.Logger.Panic().Err(err)
	}

	os.Remove(fmt.Sprintf("dat-%d.txt", b.N))

	file, err := volume.Create(fmt.Sprintf("dat-%d.txt", b.N), 0, 0755)
	//file, err := volume.Create(fmt.Sprintf("dat-%d.txt", b.N), int64(b.N * 10), 0755)
	if err != nil {
		volume.Logger.Panic().Err(err)
	}

	wrote, err := file.Write([]byte("hello"))
	if err != nil {
		volume.Logger.Panic().Err(err)
	}

	volume.Logger = volume.Logger.Level(zerolog.ErrorLevel)
	volume.Logger.Info().Msgf("wrote: %d", wrote)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//wrote, err = file.Write([]byte("this is something that is decently long to cause overflows"))
		wrote, err = file.Write([]byte("10000022"))
		if err != nil {
			volume.Logger.Panic().Err(err)
		}
	}
	b.StopTimer()

	err = file.Close()
	if err != nil {
		volume.Logger.Panic().Err(err)
	}

	moved.Logger.Info().Msgf("%s: grow count: %d in %s", file.name, file.extcount, file.extdur)
	moved.Logger.Info().Msgf("%s: mmap count: %d in %s", file.name, file.mmapcount, file.mmapdur)
}
