package document

import (
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/common/spmap"
)

var Nil = &nilValue{}

type Value interface {
	Type() moved.DataType

	Marshal() []byte
}

func UnmarshalRaw(b []byte) Value {
	if len(b) == 0 {
		return Nil
	}

	switch moved.DataType(b[0]) {
	case moved.Nil:
		return Nil
	}

	return Nil
}

type nilValue struct{}
var nilBytes = []byte {'0'}

func (nilValue) Type() moved.DataType {
	return moved.Nil
}

func (nilValue) Marshal() []byte {
	return nilBytes
}

//
//
//
type PBUFMessage []byte

func (p PBUFMessage) Type() moved.DataType {
	return moved.Protobuf
}

func (p PBUFMessage) Marshal() []byte {
	return p
}

// List
type List []Value

func (l List) Type() moved.DataType {
	return moved.List
}


type Map spmap.Map
