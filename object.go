package moved

import (
	"errors"
)

var (
	ErrInvalidType = errors.New("expected string")
)

type DataType byte

const (
	Nil    DataType = 0 // Keyable
	String DataType = 1 // Keyable

	Int    DataType = 2 // Keyable
	Int32  DataType = 3
	Uint32 DataType = 4
	Int16  DataType = 5
	Uint16 DataType = 6
	Int8   DataType = 7
	Uint8  DataType = 8

	Float   DataType = 10 // Keyable
	Float32 DataType = 11 // Keyable

	Bool  DataType = 12 // Keyable
	Rect  DataType = 13 // Keyable
	Time  DataType = 14 // Keyable
	Bytes DataType = 15
	List  DataType = 18
	Map   DataType = 19

	Struct DataType = 9

	// Message Formats
	Protobuf DataType = 20
	JSON     DataType = 21
	MsgPack  DataType = 22
	CBOR     DataType = 23
	XML      DataType = 24

	Any DataType = 0

	Composite DataType = 99

	// Data Structures
)
