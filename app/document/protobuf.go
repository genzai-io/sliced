package document

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/genzai-io/sliced/app/table"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

//
type protobufDocument []byte

func (jd protobufDocument) Type() Type {
	return Protobuf
}

func (jd protobufDocument) Bytes() []byte {
	return jd
}

func (jd protobufDocument) String() string {
	return *(*string)(unsafe.Pointer(&jd))
}

func (jd protobufDocument) ToType(t Type) Document {
	if t == Protobuf {
		return jd
	}
	switch t {
	case JSON:

	}

	return nil
}

//func (pd protobufDocument)

func ProtobufProjector(projector ProtoProjection) Projector {
	switch projector.Fields() {
	case 0:
		return NilProjector

	case 1:
		return protobufProjector(projector.FieldAt(0).Number)

	default:
		names := make([]int32, projector.Fields())
		for i := 0; i < projector.Fields(); i++ {
			names[i] = projector.FieldAt(i).Number
		}
		return protobufMultiProjector(names)
	}
	return NilProjector
}

type protobufStringProjector int32

func (jp protobufStringProjector) Project(b []byte) table.Key {
	//return table.JSONToKey(gprotobuf.GetBytes(b, (string)(jp)))
	return nil
}

func (jp protobufStringProjector) ProjectString(s string) table.Key {
	//return table.JSONToKey(gprotobuf.Get(s, (string)(jp)))
	return nil
}

// Projects a single JSON path expression.
type protobufProjector int

func (jp protobufProjector) Project(b []byte) table.Key {
	//return table.JSONToKey(gprotobuf.GetBytes(b, (string)(jp)))
	return nil
}

func (jp protobufProjector) ProjectString(s string) table.Key {
	//return table.JSONToKey(gprotobuf.Get(s, (string)(jp)))
	return nil
}

// Projects an array of JSON path expression creating Composite Keys.
type protobufMultiProjector []int32

func (jp protobufMultiProjector) Project(b []byte) table.Key {
	//v := make([]table.Key, len(jp))
	for _, p := range jp {
		_ = p
		//v[i] = table.JSONToKey(gprotobuf.GetBytes(b, p))
	}
	return nil
}

func (jp protobufMultiProjector) ProjectString(s string) table.Key {
	//v := make([]table.Key, len(jp))
	for _, p := range jp {
		_ = p
		//v[i] = table.JSONToKey(gprotobuf.Get(s, p))
	}
	return nil
}

type ProtobufReader struct {
	b []byte
	i int
	n int

	wire  int
	field int
	t     descriptor.FieldDescriptorProto_Type
	v     uint64
	sv    []byte

	proto *MessageType
}

func NewProtobufReader(mt *MessageType, b []byte) *ProtobufReader {
	r := &ProtobufReader{
		b:     b,
		proto: mt,
	}
	return r
}

var nilMessageType = &MessageType{}

func (mt *MessageType) PBUFProject(b []byte, projector protobufProjector) (key table.Key) {
	fieldNum := uint64(projector)
	mt.PBUFKeyIterator(b, mt.FieldByNumber, func(entry *PBUFKey) bool {
		if entry.FieldNum == fieldNum {
			key = entry.Key
			return false
		}
		return true
	})
	return
}

func (mt *MessageType) PBUFKeysOf(buf []byte, fields []*FieldType) (keys []table.Key, err error) {
	keys = make([]table.Key, len(fields))

	if len(fields) == 0 {
		return
	}

	var (
		index      = 0
		field      = fields[0]
		prevNumber = 0
		outOfOrder = false
	)

	mt.PBUFKeyIterator(buf,
		func(number int) *FieldType {
			if number < prevNumber {
				outOfOrder = true
			}
			if field.Number == int32(number) {
				f := field
				field = nil
				return f
			}
			return nil
		},
		func(entry *PBUFKey) bool {
			// Detect out of order
			if outOfOrder {
				return false
			}

			keys[index] = entry.Key
			index++

			if len(fields) == index {
				return false
			}

			field = fields[index]
			return true
		},
	)

	if len(fields) == index {
		return
	}

	// Pull one key at a time for the remaining keys
	fields = fields[index:]
	for ; index < len(fields); index++ {
		keys[index] = mt.PBUFProject(buf, protobufProjector(index))
	}

	return
}

type PBUFKey struct {
	Wire     int
	FieldNum uint64
	Field    *FieldType
	Key      table.Key
}

func (mt *MessageType) PBUFKeyIterator(buf []byte, fieldFn func(number int) *FieldType, fn func(entry *PBUFKey) bool) error {
	l := len(buf)
	iNdEx := 0

	if mt == nil {
		mt = nilMessageType
	}

	var entry PBUFKey

	var key uint64
	var n int
	var v uint64
	//var n int
	var vb []byte

	for len(buf) > 0 {
		// Parse the key
		key, n = binary.Uvarint(buf)
		if n <= 0 {
			return errors.New("bad protobuf field key")
		}
		buf = buf[n:]
		entry.Wire = int(key & 7)
		entry.FieldNum = key >> 3
		if entry.Wire == 4 {
			return fmt.Errorf("proto: wiretype end group for non-group")
		}
		if entry.FieldNum <= 0 {
			return fmt.Errorf("proto: illegal tag %d (wire type %d)", entry.FieldNum, key)
		}

		entry.Field = mt.FieldByNumber(int(entry.FieldNum))

		// Skip over the value if field number not found.
		if entry.Field == nil {
			// Break out the value from the buffer based on the wire type
			switch entry.Wire {
			case 0: // varint
				v, n = binary.Uvarint(buf)
				if n <= 0 {
					return errors.New("bad protobuf varint value")
				}
				buf = buf[n:]

			case 5: // 32-bit
				if len(buf) < 4 {
					return errors.New("bad protobuf 32-bit value")
				}
				buf = buf[4:]

			case 1: // 64-bit
				if len(buf) < 8 {
					return errors.New("bad protobuf 64-bit value")
				}
				buf = buf[8:]

			case 2: // length-delimited
				v, n = binary.Uvarint(buf)
				if n <= 0 || v > uint64(len(buf)-n) {
					return errors.New("bad protobuf length-delimited value")
				}
				buf = buf[n+int(v):]

			default:
				return errors.New("unknown protobuf wire-type")
			}
			continue
		}

		// Break out the value from the buffer based on the wire type
		switch entry.Wire {
		case 0: // varint
			v, n = binary.Uvarint(buf)
			if n <= 0 {
				return errors.New("bad protobuf varint value")
			}
			buf = buf[n:]

			entry.Key = protobufVarint(entry.Field.ProtobufType, v)

		case 5: // 32-bit
			if len(buf) < 4 {
				return errors.New("bad protobuf 32-bit value")
			}
			v = uint64(buf[0]) |
				uint64(buf[1])<<8 |
				uint64(buf[2])<<16 |
				uint64(buf[3])<<24
			buf = buf[4:]

			entry.Key = table.IntKey(v)

		case 1: // 64-bit
			if len(buf) < 8 {
				return errors.New("bad protobuf 64-bit value")
			}
			v = uint64(buf[0]) |
				uint64(buf[1])<<8 |
				uint64(buf[2])<<16 |
				uint64(buf[3])<<24 |
				uint64(buf[4])<<32 |
				uint64(buf[5])<<40 |
				uint64(buf[6])<<48 |
				uint64(buf[7])<<56
			buf = buf[8:]

			entry.Key = table.IntKey(v)

		case 2: // length-delimited
			v, n = binary.Uvarint(buf)
			if n <= 0 || v > uint64(len(buf)-n) {
				return errors.New("bad protobuf length-delimited value")
			}
			vb = buf[n : n+int(v) : n+int(v)]
			buf = buf[n+int(v):]

			if entry.Field.ProtobufType == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				//field.Descriptor.TypeName
				fmt.Println("Message")
			}

			entry.Key = table.StringKey(string(vb))

		default:
			return errors.New("unknown protobuf wire-type")
		}

		if !fn(&entry) {
			return nil
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func decodeZigzag(v uint64) int64 {
	return int64((v >> 1) ^ uint64((int64(v&1)<<63)>>63))
}

func (r *ProtobufReader) toKey() table.Key {
	switch r.t {
	// 0 is reserved for errors.
	// Order is weird for historical reasons.
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return table.FloatKey(*(*float64)(unsafe.Pointer(&r.v)))
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return table.FloatKey(*(*float64)(unsafe.Pointer(&r.v)))
		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT64 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		x := int64(r.v >> 1)
		if r.v&1 != 0 {
			x = ^x
		}
		return table.FloatKey(*(*float64)(unsafe.Pointer(&r.v)))

	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return table.IntKey(r.v)

		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT32 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		x := int64(r.v >> 1)
		if r.v&1 != 0 {
			x = ^x
		}
		return table.IntKey(*(*float64)(unsafe.Pointer(&r.v)))

	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		x := int64(r.v >> 1)
		if r.v&1 != 0 {
			x = ^x
		}
		return table.IntKey(*(*float64)(unsafe.Pointer(&r.v)))

	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		x := int64(r.v >> 1)
		if r.v&1 != 0 {
			x = ^x
		}
		return table.IntKey(*(*float64)(unsafe.Pointer(&r.v)))

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if r.v != 0 {
			return table.True
		} else {
			return table.False
		}

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return table.StringKey(r.sv)

		// Tag-delimited aggregate.
		// Group type is deprecated and not supported in proto3. However, Proto3
		// implementations should still be able to parse the group wire format and
		// treat group fields as unknown fields.
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return table.Nil

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return table.Nil

		// New in version 2.
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return table.StringKey(r.sv)

	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return table.IntKey(r.v)

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
	}

	return table.Nil
}

func protobufVarint(t descriptor.FieldDescriptorProto_Type, raw uint64) table.Key {
	switch t {
	// 0 is reserved for errors.
	// Order is weird for historical reasons.
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return table.FloatKey(*(*float64)(unsafe.Pointer(&raw)))
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return table.FloatKey(*(*float64)(unsafe.Pointer(&raw)))
		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT64 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		//x := int64(raw >> 1)
		//if raw&1 != 0 {
		//	x = ^x
		//}
		return table.IntKey(raw)

	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return table.IntKey(raw)

		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT32 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		x := int64(raw >> 1)
		if raw&1 != 0 {
			x = ^x
		}
		return table.IntKey(*(*float64)(unsafe.Pointer(&raw)))

	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		x := int64(raw >> 1)
		if raw&1 != 0 {
			x = ^x
		}
		return table.IntKey(x)

	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		x := int64(raw >> 1)
		if raw&1 != 0 {
			x = ^x
		}
		return table.IntKey(x)

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if raw != 0 {
			return table.True
		} else {
			return table.False
		}

	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return table.IntKey(raw)

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return table.IntKey(raw)

	case descriptor.FieldDescriptorProto_TYPE_SFIXED32, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return table.IntKey(decodeZigzag(raw))

	case descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SINT64:
		return table.IntKey(decodeZigzag(raw))
	}

	return table.SkipKey
}
