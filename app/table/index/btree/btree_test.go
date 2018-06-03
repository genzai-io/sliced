package btree

import (
	"testing"
	"unsafe"
	"github.com/armon/go-radix"
	"fmt"
	"strconv"
)

func BenchmarkBTree_Clone(b *testing.B) {
	by := []byte("HI")
	for i := 0; i < b.N; i++ {
		toString(by)
	}
}

func BenchmarkBTree_Clone2(b *testing.B) {
	by := []byte("HI")
	for i := 0; i < b.N; i++ {
		castString(by)
	}
}

func toString(b []byte) string {
	return string(b)
}

func castString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func BenchmarkPut(b *testing.B) {
	val := []byte("Some value")
	tree := New(64, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		tree.ReplaceOrInsert(&TreeItem{Key: key, Value: val})
	}
}

func BenchmarkGet(b *testing.B) {
	val := []byte("Some value")
	tree := New(64, nil)

	for i := 0; i < 200000; i++ {
		key := fmt.Sprintf("%d", i)
		tree.ReplaceOrInsert(&TreeItem{Key: key, Value: val})
	}

	item := &TreeItem{Key: "100000"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get(item)
	}
}

func BenchmarkRadixGet(b *testing.B) {
	val := []byte("Some value")
	tree := radix.New()

	for i := 0; i < 200000; i++ {
		key := fmt.Sprintf("%d", i)
		tree.Insert(key, val)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get("100000")
	}
}

func BenchmarkRadixPut(b *testing.B) {
	val := []byte("Some value")
	tree := radix.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		tree.Insert(key, val)
	}
}

func BenchmarkHashPut(b *testing.B) {
	val := []byte("Some value")
	tree := make(map[string]*TreeItem)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree[fmt.Sprintf("%d", i)] = &TreeItem{Key: "", Value: val}
	}
}

func BenchmarkHashGet(b *testing.B) {
	val := []byte("Some value")
	tree := make(map[string]*TreeItem)

	for i := 0; i < 200000; i++ {
		tree[fmt.Sprintf("%d", i)] = &TreeItem{Key: "", Value: val}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item, ok := tree["100000"]
		if !ok {

		} else if item != nil {

		}
	}
}

func BenchmarkBTree_Ascend(b *testing.B) {

}
