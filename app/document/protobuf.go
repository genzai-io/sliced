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

var (
	ErrProtobufField    = errors.New("protobuf bad field key")
	ErrProtobufWireType = errors.New("protobuf bad wiretype")
	ErrProtobufVarint   = errors.New("protobuf bad varint value")
	ErrProtobuf32bit    = errors.New("protobuf bad 32-bit value")
	ErrProtobuf64bit    = errors.New("protobuf bad 64-bit value")
	ErrProtobufLength   = errors.New("protobuf bad length-delimited value")
)

//
type PBUFKey struct {
	Wire     int
	FieldNum uint64
	Field    *FieldType
	Key      table.Key
}

//
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

			entry.Key = pbufVarintToKey(entry.Field.ProtobufType, v)

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
				entry.Field.Message.PBUFKeyIterator(vb, entry.Field.Message.FieldByNumber, func(entry *PBUFKey) bool {
					return true
				})
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

// Gets multiple fields from a protobuf serialized message.
// It expects the protobuf message to be sorted in ascending field number order and
// expects the specified fields to be ordered the same way. It will discover if either
// are out of order and will degrade the algorithm to scan the document for each remaining
// field it hasn't projected.
func (mt *MessageType) PBUFGet(buf []byte, fields []*FieldType, keys []table.Key) ([]table.Key, error) {
	keys = keys[:len(fields)]
	if len(fields) == 0 {
		return keys, nil
	}

	var (
		index      = 0
		l          = len(buf)
		fieldidx   = 0
		field      = fields[0]
		prevNumber = int32(0)
	)

	for index < l {
		var key uint64

		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return keys, ErrProtobufField
			}
			if index >= l {
				return keys, io.ErrUnexpectedEOF
			}
			b := buf[index]
			index++
			key |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}

		wire := int(key & 7)
		fieldNum := int32(key >> 3)

		if fieldNum <= 0 {
			return keys, fmt.Errorf("protobuf illegal tag %d (wire type %d)", fieldNum, key)
		}

		// Is it out of order?
		// Do a get on each field. This is much slower, but it guarantees to not miss any fields.
		if prevNumber > fieldNum {
			var err error
			for i, f := range fields {
				if keys[i], err = f.PBUFGet(buf); err != nil {
					return keys, err
				}
			}
			return keys, nil
		}

		prevNumber = fieldNum

	Match:
		if fieldNum != field.Number {
			// Is it missing from the message?
			if fieldNum > field.Number {
				keys[fieldidx] = table.SkipKey
				fieldidx++
				if fieldidx == len(fields) {
					return keys, nil
				}
				field = fields[fieldidx]
				prevNumber = fieldNum
				goto Match
			}

			// Skip over the value
			switch wire {
			case 0: // varint
				key = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return keys, ErrProtobufVarint
					}
					if index >= l {
						return keys, io.ErrUnexpectedEOF
					}
					b := buf[index]
					index++
					key |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}

			case 5: // 32-bit
				index += 4
				if index >= l {
					return keys, ErrProtobuf32bit
				}

			case 1: // 64-bit
				index += 8
				if index >= l {
					return keys, ErrProtobuf64bit
				}

			case 2: // length-delimited
				key = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return keys, ErrProtobufLength
					}
					if index >= l {
						return keys, io.ErrUnexpectedEOF
					}
					b := buf[index]
					index++
					key |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						index += int(key)
						if index >= l {
							return keys, io.ErrUnexpectedEOF
						}
						break
					}
				}

			default:
				return keys, ErrProtobufWireType
			}
			continue
		}

		// Break out the value from the buffer based on the wire type
		switch wire {
		case 0: // varint
			key = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return keys, ErrProtobufVarint
				}
				if index >= l {
					return keys, io.ErrUnexpectedEOF
				}
				b := buf[index]
				index++
				key |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}

			keys[fieldidx] = pbufVarintToKey(field.ProtobufType, key)

		case 5: // 32-bit
			index += 4
			if index >= l {
				return keys, io.ErrUnexpectedEOF
			}
			keys[fieldidx] = table.IntKey(uint64(buf[index]) |
				uint64(buf[index+1])<<8 |
				uint64(buf[index+2])<<16 |
				uint64(buf[index+3])<<24)

		case 1: // 64-bit
			index += 8
			if index >= l {
				return keys, io.ErrUnexpectedEOF
			}
			keys[fieldidx] = table.IntKey(uint64(buf[index]) |
				uint64(buf[index+1])<<8 |
				uint64(buf[index+2])<<16 |
				uint64(buf[index+3])<<24 |
				uint64(buf[index+4])<<32 |
				uint64(buf[index+5])<<40 |
				uint64(buf[index+6])<<48 |
				uint64(buf[index+7])<<56)

		case 2: // length-delimited
			key = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return keys, ErrProtobufLength
				}
				if index >= l {
					return keys, io.ErrUnexpectedEOF
				}
				b := buf[index]
				index++
				key |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					end := int(key) + index
					if end >= l {
						return keys, io.ErrUnexpectedEOF
					}
					keys[fieldidx] = table.StringKey(buf[index:end])
					index = end
					break
				}
			}

		default:
			return keys, ErrProtobufWireType
		}

		fieldidx++
		if fieldidx == len(fields) {
			return keys, nil
		}
		field = fields[fieldidx]
		// Are the fields not sorted by Number?
		if field.Number < fieldNum {
			var err error
			// Get the remaining one by one
			for ; fieldidx < len(fields); fieldidx++ {
				if keys[fieldidx], err = fields[fieldidx].PBUFGet(buf); err != nil {
					return keys, err
				}
			}
			return keys, nil
		}
	}

	return keys, nil
}

// Gets a single value
func (mt *FieldType) PBUFGet(buf []byte) (table.Key, error) {
	index := 0
	l := len(buf)

	for index < l {
		var key uint64

		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return table.SkipKey, ErrProtobufField
			}
			if index >= l {
				return table.SkipKey, io.ErrUnexpectedEOF
			}
			b := buf[index]
			index++
			key |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}

		wire := int(key & 7)
		fieldNum := int32(key >> 3)

		if fieldNum <= 0 {
			return table.SkipKey, fmt.Errorf("proto: illegal tag %d (wire type %d)", fieldNum, key)
		}

		// Skip over the value if field number not found.
		if mt.Number != fieldNum {
			// Break out the value from the buffer based on the wire type
			switch wire {
			case 0: // varint
				key = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return table.SkipKey, ErrProtobufVarint
					}
					if index >= l {
						return table.SkipKey, io.ErrUnexpectedEOF
					}
					b := buf[index]
					index++
					key |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}

			case 5: // 32-bit
				index += 4
				if index >= l {
					return table.SkipKey, ErrProtobuf32bit
				}

			case 1: // 64-bit
				index += 8
				if index >= l {
					return table.SkipKey, ErrProtobuf64bit
				}

			case 2: // length-delimited
				key = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return table.SkipKey, ErrProtobufLength
					}
					if index >= l {
						return table.SkipKey, io.ErrUnexpectedEOF
					}
					b := buf[index]
					index++
					key |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						index += int(key)
						if index >= l {
							return table.SkipKey, io.ErrUnexpectedEOF
						}
						break
					}
				}

			default:
				return table.SkipKey, ErrProtobufWireType
			}
			continue
		}

		// Break out the value from the buffer based on the wire type
		switch wire {
		case 0: // varint
			key = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return table.SkipKey, ErrProtobufField
				}
				if index >= l {
					return table.SkipKey, io.ErrUnexpectedEOF
				}
				b := buf[index]
				index++
				key |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}

			return pbufVarintToKey(mt.ProtobufType, key), nil

		case 5: // 32-bit
			index += 4
			if index >= l {
				return table.SkipKey, io.ErrUnexpectedEOF
			}
			return table.IntKey(uint64(buf[index]) |
				uint64(buf[index+1])<<8 |
				uint64(buf[index+2])<<16 |
				uint64(buf[index+3])<<24), nil

		case 1: // 64-bit
			index += 8
			if index >= l {
				return table.SkipKey, io.ErrUnexpectedEOF
			}
			return table.IntKey(uint64(buf[index]) |
				uint64(buf[index+1])<<8 |
				uint64(buf[index+2])<<16 |
				uint64(buf[index+3])<<24 |
				uint64(buf[index+4])<<32 |
				uint64(buf[index+5])<<40 |
				uint64(buf[index+6])<<48 |
				uint64(buf[index+7])<<56), nil

		case 2: // length-delimited
			key = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return table.SkipKey, ErrProtobufLength
				}
				if index >= l {
					return table.SkipKey, io.ErrUnexpectedEOF
				}
				b := buf[index]
				index++
				key |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					end := int(key) + index
					if end >= l {
						return table.SkipKey, io.ErrUnexpectedEOF
					}
					return table.StringKey(buf[index:end]), nil
				}
			}

		default:
			return table.SkipKey, ErrProtobufWireType
		}
	}

	return table.SkipKey, nil
}

//
func decodeZigzag(v uint64) int64 {
	return int64((v >> 1) ^ uint64((int64(v&1)<<63)>>63))
}

func pbufVarintToKey(t descriptor.FieldDescriptorProto_Type, raw uint64) table.Key {
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
