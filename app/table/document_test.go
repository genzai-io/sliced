package table

import (
	"fmt"
	"testing"

	"github.com/genzai-io/sliced/app/codec/gjson"
)

func TestJSONValue_Get(t *testing.T) {
	document := DocumentString(make([]byte, 0, 64))
	document = append(document,
		16,
		2, 'h', 'i',
		1, 1,
		9, 'n', 'a', 'm', 'e', '.', 'l', 'a', 's', 't',
		5, 'S', 'm', 'i', 't', 'h',
	)

	offset, length := document.getFromHeader("name.last")
	fmt.Println(offset, "  ", length)
	fmt.Println(string(document[offset:offset+length]))
}

func TestJSONValue_Get2(t *testing.T) {
	document := DocumentString(make([]byte, 0, 64))
	document = append(document,
		16,
		2, 'h', 'i',
		1, 1,
		5, 'n', 'a', 'm', 'e', '.',
		255, 35, // Skip obj
		4, 'l', 'a', 's', 't',
		5, 'S', 'm', 'i', 't', 'h',
		5, 'f', 'i', 'r', 's', 't',
		5, 'F', 'r', 'a', 'n', 'k',
		255, 24,
		0,
	)

	offset, length := document.getFromHeader("name.last")
	fmt.Println(offset, "  ", length)
	fmt.Println(string(document[offset:offset+length]))
}

func BenchmarkJSONValue_Get(b *testing.B) {
	document := DocumentString(make([]byte, 0, 64))
	document = append(document,
		16,
		2, 'h', 'i',
		1, 1,
		9, 'a', 'a', 'm', 'e', '.', 'a', 'a', 's', 't',
		5, 'S', 'm', 'i', 't', 'h',
		9, 'a', 'a', 'm', 'e', '.', 'a', 'a', 's', 't',
		5, 'S', 'm', 'i', 't', 'h',
		9, 'n', 'a', 'm', 'e', '.', 'l', 'a', 's', 't',
		5, 'S', 'm', 'i', 't', 'h',
	)
	path := "name.last"
	document[0] = byte(len(document))

	for i := 0; i < b.N; i++ {
		offset, length := document.getFromHeader(path)
		_ = offset
		_ = length
	}
}

func BenchmarkJSONValue_GetJSON(b *testing.B) {
	json := `{"name":{"first":"John", "middle", "Richard", "last": "Smith"} "location":[10.2 34.0]}`
	path := "location"
	//ctx := &gjson.ParseContext{}

	for i := 0; i < b.N; i++ {
		//gjson.GetWithContext(json, path, ctx)
		//ctx.Reset()
		gjson.Get(json, path)
		//offset, length := document.getFromHeader(path)
		//_ = offset
		//_ = length
	}
}
