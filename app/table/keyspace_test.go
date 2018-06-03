package table

import (
	"fmt"
	"testing"
)

func TestParsePrimaryKey(t *testing.T) {
	key := ParsePrimaryKey([]byte("orders:0001"))
	fmt.Println(key)
	fmt.Println(fmt.Sprintf("%s, $s", key.Type()))
}
