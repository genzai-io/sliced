// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package record

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"testing"
	"time"
	"unsafe"

	"github.com/google/btree"
	"github.com/tidwall/lotsa"

	store_pb "github.com/slice-d/genzai/proto/store"
)

func init() {
	//var seed int64 = 1520031745261947354
	seed := (time.Now().UnixNano())
	println("seed:", seed)
	rand.Seed(seed)
}

func TestTree_Insert(t *testing.T) {
	var tr Tree

	tr.Insert(store_pb.RecordID{100292, 1}, nil, 0)
	tr.Insert(store_pb.RecordID{100292, 0}, nil, 0)
	tr.Insert(store_pb.RecordID{5, 0}, nil, 0)
	tr.Insert(store_pb.RecordID{31, 0}, nil, 0)
	tr.Insert(store_pb.RecordID{16, 0}, nil, 0)
	tr.Insert(store_pb.RecordID{9, 0}, nil, 0)
	tr.Insert(store_pb.RecordID{5, 1}, nil, 0)

	tr.Scan(func(cell store_pb.RecordID, value unsafe.Pointer, extra uint64) bool {
		fmt.Println(cell)
		return true
	})
}

func random(N int, perm bool) []store_pb.RecordID {
	ints := make([]store_pb.RecordID, N)
	if perm {
		for i, x := range rand.Perm(N) {
			ints[i] = store_pb.RecordID{uint64(x), 0}
		}
	} else {
		m := make(map[store_pb.RecordID]bool)
		for len(m) < N {
			m[store_pb.RecordID{uint64(rand.Int63()), 0}] = true
		}
		var i int
		for k := range m {
			ints[i] = k
			i++
		}
	}
	return ints
}

func shuffle(ints []store_pb.RecordID) {
	for i := range ints {
		j := rand.Intn(i + 1)
		ints[i], ints[j] = ints[j], ints[i]
	}
}

func sortInts(ints []store_pb.RecordID) {
	sort.Slice(ints, func(i, j int) bool {
		return IsLess(ints[i], ints[j])
	})
}

func testEquals(t *testing.T, random, sorted []store_pb.RecordID) {
	t.Helper()
	sortInts(random)
	if len(sorted) != len(random) {
		t.Fatal("not equal")
	}
	for i := 0; i < len(sorted); i++ {
		if sorted[i] != random[i] {
			println(2)
			t.Fatal("not equal")
		}
	}
}

func TestRandom(t *testing.T) {
	start := time.Now()
	for time.Since(start) < time.Second {
		N := (rand.Int() % 10000)
		if N%2 == 1 {
			N++
		}
		ints := random(N, rand.Int()%2 == 0)
		var tr Tree
		for i := 0; i < N; i++ {
			tr.Insert(ints[i], nil, 0)
		}
		if tr.Len() != N {
			t.Fatalf("expected %v, got %v", N, tr.Len())
		}
		var all []store_pb.RecordID
		tr.Scan(func(cell store_pb.RecordID, data unsafe.Pointer, extra uint64) bool {
			all = append(all, cell)
			return true
		})
		testEquals(t, ints, all)
		if N > 0 {
			shuffle(ints)
			start := ints[len(ints)/2]
			var all []store_pb.RecordID
			tr.Range(start, func(cell store_pb.RecordID, data unsafe.Pointer, extra uint64) bool {
				all = append(all, cell)
				return true
			})
			sortInts(ints)
			var halved []store_pb.RecordID
			for i := 0; i < len(ints); i++ {
				if IsGreaterOrEqual(ints[i], start) {
					halved = ints[i:]
					break
				}
			}
			testEquals(t, halved, all)
		}
		shuffle(ints)
		for i := 0; i < len(ints)/2; i++ {
			tr.Remove(ints[i], nil)
		}
		if tr.Len() != N/2 {
			t.Fatalf("expected %v, got %v", N/2, tr.Len())
		}
		for i := len(ints) / 2; i < len(ints); i++ {
			tr.Remove(ints[i], nil)
		}
		if tr.Len() != 0 {
			t.Fatalf("expected %v, got %v", 0, tr.Len())
		}
	}
}
//
//// func TestExample(t *testing.T) {
//// 	var tr RecordTree
//
//// 	tr.Insert(10, nil, 0)
//// 	tr.Insert(5, nil, 0)
//// 	tr.Insert(31, nil, 0)
//// 	tr.Insert(16, nil, 0)
//// 	tr.Insert(9, nil, 0)
//
//// 	tr.Scan(func(cell uint64, value unsafe.Pointer, extra uint64) bool {
//// 		println(cell)
//// 		return true
//// 	})
//// }
//
type perfCtx struct {
	_insert func(cell store_pb.RecordID)
	_count  func() int
	_scan   func()
	_range  func(cell store_pb.RecordID, iter func(cell store_pb.RecordID) bool)
	_remove func(cell store_pb.RecordID)
}

type btreeItem store_pb.RecordID

func (v btreeItem) Less(v2 btree.Item) bool {
	return IsLess(store_pb.RecordID(v), store_pb.RecordID(v2.(btreeItem)))
}

func TestPerf(t *testing.T) {
	t.Run("CellTree", func(t *testing.T) {
		println("-- celltree --")
		var tr Tree
		ctx := perfCtx{
			_insert: func(cell store_pb.RecordID) { tr.Insert(cell, nil, 0) },
			_count:  func() int { return tr.Len() },
			_scan: func() {
				tr.Scan(func(cell store_pb.RecordID, data unsafe.Pointer, extra uint64) bool {
					return true
				})
			},
			_range: func(cell store_pb.RecordID, iter func(cell store_pb.RecordID) bool) {
				tr.Range(cell, func(cell store_pb.RecordID, data unsafe.Pointer, extra uint64) bool {
					return iter(cell)
				})
			},
			_remove: func(cell store_pb.RecordID) { tr.Remove(cell, nil) },
		}
		testPerf(t, ctx)
	})
	t.Run("BTree", func(t *testing.T) {
		println("-- btree --")
		tr := btree.New(16)
		ctx := perfCtx{
			_insert: func(cell store_pb.RecordID) { tr.ReplaceOrInsert(btreeItem(cell)) },
			_count:  func() int { return tr.Len() },
			_scan: func() {
				tr.Ascend(func(item btree.Item) bool {
					return true
				})
			},
			_range: func(cell store_pb.RecordID, iter func(cell store_pb.RecordID) bool) {
				tr.AscendGreaterOrEqual(btreeItem(cell), func(item btree.Item) bool {
					return iter(store_pb.RecordID(item.(btreeItem)))
				})
			},
			_remove: func(cell store_pb.RecordID) { tr.Delete(btreeItem(cell)) },
		}
		testPerf(t, ctx)
	})
}
//
func testPerf(t *testing.T, ctx perfCtx) {
	N := 1024 * 1024
	ints := random(N, false)
	var ms1, ms2 runtime.MemStats
	defer func() {
		heapBytes := int(ms2.HeapAlloc - ms1.HeapAlloc)
		fmt.Printf("memory %13s bytes %s/entry \n",
			commaize(heapBytes), commaize(heapBytes/len(ints)))
		fmt.Printf("\n")
	}()
	runtime.GC()
	time.Sleep(time.Millisecond * 100)
	runtime.ReadMemStats(&ms1)

	var start time.Time
	var dur time.Duration
	output := func(tag string, N int) {
		dur = time.Since(start)
		fmt.Printf("%-8s %10s ops in %4dms %10s/sec\n",
			tag, commaize(N), int(dur.Seconds()*1000),
			commaize(int(float64(N)/dur.Seconds())))
	}

	/////////////////////////////////////////////
	start = time.Now()
	lotsa.Ops(N, 1, func(i, _ int) {
		ctx._insert(ints[i])
	})
	output("insert", N)
	runtime.GC()
	time.Sleep(time.Millisecond * 100)
	runtime.ReadMemStats(&ms2)
	if ctx._count() != N {
		t.Fatalf("expected %v, got %v", N, ctx._count())
	}
	/////////////////////////////////////////////
	shuffle(ints)
	start = time.Now()
	lotsa.Ops(100, 1, func(i, _ int) { ctx._scan() })
	output("scan", 100)
	/////////////////////////////////////////////
	sortInts(ints)
	start = time.Now()
	lotsa.Ops(N, 1, func(i, _ int) {
		var found bool
		ctx._range(ints[i], func(cell store_pb.RecordID) bool {
			if cell != ints[i] {
				t.Fatal("invalid")
			}
			found = true
			return false
		})
		if !found {
			t.Fatal("not found")
		}
	})
	output("range", N)
	/////////////////////////////////////////////
	shuffle(ints)
	start = time.Now()
	lotsa.Ops(N, 1, func(i, _ int) {
		ctx._remove(ints[i])
	})
	output("remove", N)
	if ctx._count() != 0 {
		t.Fatalf("expected %v, got %v", 0, ctx._count())
	}
}

func commaize(n int) string {
	s1, s2 := fmt.Sprintf("%d", n), ""
	for i, j := len(s1)-1, 0; i >= 0; i, j = i-1, j+1 {
		if j%3 == 0 && j != 0 {
			s2 = "," + s2
		}
		s2 = string(s1[i]) + s2
	}
	return s2
}

func TestPerfLongTime(t *testing.T) {
	if os.Getenv("PERFLONGTIME") != "1" {
		fmt.Printf("TestPerfLongTime disabled (PERFLONGTIME=1)\n")
		return
	}
	x := 0
	N := 1024 * 1024
	ints := random(N, false)
	var tr Tree
	var insops, remops int
	var ms1, ms2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&ms1)
	start := time.Now()
	var insdur, remdur time.Duration
	var xstart time.Time

	// insert all items

	for i := 0; i < len(ints); i++ {
		tr.Insert(ints[i], nil, 0)
		insops++
	}
	insdur += time.Since(start)
	// now delete every 4th item and rerandomize
	for {
		//opp := rand.Uint64()
		xstart = time.Now()
		for i := x; i < len(ints); i += 4 {
			tr.Remove(ints[i], nil)
			//ints[i] ^= opp TODO: Fix
			remops++
		}
		remdur += time.Since(xstart)
		xstart = time.Now()
		for i := x; i < len(ints); i += 4 {
			tr.Insert(ints[i], nil, 0)
			insops++
		}
		insdur += time.Since(xstart)
		if tr.Len() != N {
			t.Fatal("shit")
		}
		runtime.GC()
		runtime.ReadMemStats(&ms2)
		heapBytes := int(ms2.HeapAlloc - ms1.HeapAlloc)
		x = (x + 1) % 4
		dur := time.Since(start)

		fmt.Printf("\r  %10s ops %10s ins/sec %10s rem/sec (%s bytes/cell)\r",
			commaize(insops+remops),
			commaize(int(float64(insops)/insdur.Seconds())),
			commaize(int(float64(remops)/remdur.Seconds())),
			commaize(heapBytes/N),
		)
		if dur > time.Minute {
			break
		}
	}
	fmt.Printf("\n")
}
