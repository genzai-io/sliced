package document

import (
	"fmt"
	"log"
	"testing"

	"github.com/valyala/fastjson"
)

func TestJsonVisitor(t *testing.T) {
	s := `{
		"obj": { "foo": 1234 },
		"arr": [ 23,4, "bar" ],
		"str": "foobar"
	}`

	var p fastjson.Parser
	v, err := p.Parse(s)
	if err != nil {
		log.Fatalf("cannot parse json: %s", err)
	}
	o, err := v.Object()
	if err != nil {
		log.Fatalf("cannot obtain object from json value: %s", err)
	}

	o.Visit(func(k []byte, v *fastjson.Value) {
		fmt.Println(v.Type())
		//fmt.Println(string(k))
		//fmt.Println(v.String())

		switch string(k) {
		case "obj":
			if o, err = v.Object(); err == nil {
				o.Visit(func(key []byte, v *fastjson.Value) {
					fmt.Println(string(key) + " -> " + v.String())
				})
			}
			//fmt.Printf("object %s\n", v)
		case "arr":
			fmt.Println("Array ->")
			if a, e := v.Array(); e == nil {
				for _, element := range a {
					fmt.Println(element.String())
				}
			}

			//fmt.Printf("array %s\n", v)
		case "str":
			fmt.Printf("string %s\n", v)
		}
	})
}
