package mmap

import (
	"os"
	"testing"
)

func TestMMap_Lock(t *testing.T) {
	file, err := os.OpenFile("m.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = file.Truncate(int64(1))
	if err != nil {
		t.Fatal(err)
	}

	m, err := MapRegion(file, os.Getpagesize(), RDWR, RANDOM, 0)
	//m, err := MapRegion(file, os.Getpagesize(), RDWR, 0, 0)
	//m, err := Map(file, RDWR, ANON)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Lock()
	if err != nil {
		t.Fatal(err)
	}

	m[0] = '1'
	m[1] = '2'
	m[2] = '3'

	file.Truncate(int64(3))

	err = m.Flush()
	if err != nil {
		t.Fatal(err)
	}
	err = m.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	err = m.Unmap()
	if err != nil {
		t.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkMap(b *testing.B) {
	file, err := os.OpenFile("m.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		b.Fatal(err)
	}

	regionSize := os.Getpagesize()

	err = file.Truncate(int64(regionSize))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, err := MapRegion(file, regionSize, RDWR, 0, 0)
		if err != nil {
			b.Fatal(err)
		}
		m.Unmap()
	}
}

func BenchmarkMlock(b *testing.B) {
	file, err := os.OpenFile("m.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		b.Fatal(err)
	}

	regionSize := os.Getpagesize()

	err = file.Truncate(int64(regionSize))
	if err != nil {
		b.Fatal(err)
	}

	m, err := MapRegion(file, regionSize, RDWR, 0, 0)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lock()
		m.Unlock()
	}

	m.Unmap()
}

func BenchmarkTruncate(b *testing.B) {
	file, err := os.OpenFile("m.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		b.Fatal(err)
	}

	regionSize := os.Getpagesize()

	err = file.Truncate(int64(regionSize))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = file.Truncate(int64(regionSize))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStat(b *testing.B) {
	file, err := os.OpenFile("m.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		b.Fatal(err)
	}

	regionSize := os.Getpagesize()

	err = file.Truncate(int64(regionSize))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = file.Stat()
		if err != nil {
			b.Fatal(err)
		}
	}
}
