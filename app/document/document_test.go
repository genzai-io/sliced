package document

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"testing"
	"unsafe"

	"github.com/genzai-io/sliced/proto/store"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

)

var (
	ErrInvalidLengthStore = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowStore   = fmt.Errorf("proto: integer overflow")
)

func encodeVarintStore(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}

func decode(dAtA []byte) error {
	buf := dAtA
	l := len(dAtA)
	iNdEx := 0

	var v uint64
	//var n int
	var vb []byte

	for len(buf) > 0 {
		// Parse the key
		key, n := binary.Uvarint(buf)
		if n <= 0 {
			return errors.New("bad protobuf field key")
		}
		buf = buf[n:]
		wireType := int(key & 7)
		fieldNum := key >> 3
		if wireType == 4 {
			return fmt.Errorf("proto: Roller: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Roller: illegal tag %d (wire type %d)", fieldNum, key)
		}

		fmt.Println(fmt.Sprintf("\nField: %d\nWireType: %d", fieldNum, wireType))

		// Break out the value from the buffer based on the wire type
		switch wireType {
		case 0: // varint
			v, n = binary.Uvarint(buf)
			if n <= 0 {
				return errors.New("bad protobuf varint value")
			}
			buf = buf[n:]

			i64 := *(*int64)(unsafe.Pointer(&v))
			fl := *(*float64)(unsafe.Pointer(&v))
			fmt.Println(fmt.Sprintf("Varint unsigned: %d", v))
			fmt.Println(fmt.Sprintf("Varint signed: %d", i64))
			fmt.Println(fmt.Sprintf("Varint float: %d", fl))
			fmt.Println(fmt.Sprintf("Varint zigzag: %d", int64((v >> 1) ^ uint64((int64(v&1)<<63)>>63))))

		case 5: // 32-bit
			if len(buf) < 4 {
				return errors.New("bad protobuf 32-bit value")
			}
			v = uint64(buf[0]) |
				uint64(buf[1])<<8 |
				uint64(buf[2])<<16 |
				uint64(buf[3])<<24
			buf = buf[4:]

			fmt.Println(fmt.Sprintf("32-bit: %s", v))

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

			fmt.Println(fmt.Sprintf("64-bit: %s", v))

		case 2: // length-delimited
			v, n = binary.Uvarint(buf)
			if n <= 0 || v > uint64(len(buf)-n) {
				return errors.New("bad protobuf length-delimited value")
			}
			vb = buf[n : n+int(v) : n+int(v)]
			buf = buf[n+int(v):]

			fmt.Println(fmt.Sprintf("String: %s", string(vb)))

		default:
			return errors.New("unknown protobuf wire-type")
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func decodeValue(buf []byte, wireType int) error {

	return nil
}

func Test(t *testing.T) {
	//buffer := proto.NewBuffer([]byte(""))
	//buffer.EncodeVarint(1)

	topic := &store.Topic{}
	topic.Name = "orders"
	topic.Id = 1003
	topic.QueueID = -200
	topic.RollerID = "standard"

	data, err := topic.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = decode(data)

	topic2 := &store.Topic{}
	err = topic2.Unmarshal(data)

	//d, _ := topic2.Descriptor()
	//fmt.Println(string(d))

	fileDescriptor, descriptorProto := descriptor.ForMessage(topic2)


	gz, _ := topic2.Descriptor()
	extractFile(gz)


	fmt.Println(fileDescriptor)
	fmt.Println(descriptorProto)

	fields := descriptorProto.Field
	for _, field := range fields {
		fmt.Println(*field.Name)
		fmt.Println(*field.Type)
	}


	//fmt.Println(fieldDescriptorProto)
	//fmt.Println(f2)
}

