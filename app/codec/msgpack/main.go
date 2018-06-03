package main

import (
	"github.com/vmihailenco/msgpack"
	"bytes"
	"fmt"
	"github.com/gogo/protobuf/proto"
)

func main() {
	proto.Unmarshal()
	//b, err := msgpack.Marshal([]map[string]interface{}{
	//	{"id": 1, "attrs": map[string]interface{}{"phone": 12345}},
	//	{"id": 2, "attrs": map[string]interface{}{"phone": 54321}},
	//})
	b, err := msgpack.Marshal(map[string]interface{}{
		"id": 1, "name": map[string]interface{} {"first": "Mike", "last": "Wilkins"}, "attrs": map[string]interface{}{"phone": 12345},
	})
	if err != nil {
		panic(err)
	}

	dec := msgpack.NewDecoder(bytes.NewBuffer(b))
	values, err := dec.Query("attrs.phone")
	if err != nil {
		panic(err)
	}
	fmt.Println("phones are", values)

	dec.Reset(bytes.NewBuffer(b))
	values, err = dec.Query("attrs.phone")
	if err != nil {
		panic(err)
	}
	fmt.Println("2nd phone is", values[0])

	dec.Reset(bytes.NewBuffer(b))
	values, err = dec.Query("name.last")
	if err != nil {
		panic(err)
	}
	fmt.Println("id", values)
}
