package document

import (
	"unsafe"

	"github.com/genzai-io/sliced/app/table"
	"github.com/genzai-io/sliced/common/gjson"
)

//
type jsonDocument []byte

func (jd jsonDocument) Type() Type {
	return JSON
}

func (jd jsonDocument) Bytes() []byte {
	return jd
}

func (jd jsonDocument) String() string {
	return *(*string)(unsafe.Pointer(&jd))
}

func (jd jsonDocument) Projector(projector ProtoProjection) Projector {
	switch projector.Fields() {
	case 0:
		return NilProjector

	case 1:
		return jsonProjector(projector.FieldAt(0).JsonName)

	default:
		names := make([]string, projector.Fields())
		for i := 0; i < projector.Fields(); i++ {
			names[i] = projector.FieldAt(i).JsonName
		}
		return jsonMultiProjector(names)
	}
	return NilProjector
}

// Projects a single JSON path expression.
type jsonProjector string

func (jp jsonProjector) Project(b []byte) table.Key {
	return table.JSONToKey(gjson.GetBytes(b, (string)(jp)))
}

func (jp jsonProjector) ProjectString(s string) table.Key {
	return table.JSONToKey(gjson.Get(s, (string)(jp)))
}

// Projects an array of JSON path expression creating Composite Keys.
type jsonMultiProjector []string

func (jp jsonMultiProjector) Project(b []byte) table.Key {
	v := make([]table.Key, len(jp))
	for i, p := range jp {
		v[i] = table.JSONToKey(gjson.GetBytes(b, p))
	}
	return nil
}

func (jp jsonMultiProjector) ProjectString(s string) table.Key {
	v := make([]table.Key, len(jp))
	for i, p := range jp {
		v[i] = table.JSONToKey(gjson.Get(s, p))
	}
	return nil
}
