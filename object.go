package moved

import (
	"errors"
)

var (
	ErrInvalidType = errors.New("expected string")
)

type DataType byte

const (
	Nil       DataType = 0 // Keyable
	String    DataType = 1 // Keyable
	Int       DataType = 2 // Keyable
	Float     DataType = 3 // Keyable
	Bool      DataType = 4 // Keyable
	Rect      DataType = 5 // Keyable
	Timestamp DataType = 6 // Keyable
	List      DataType = 7

	Struct DataType = 8

	// Message Formats
	JSON    DataType = 14
	MsgPack DataType = 15

	Any DataType = 0

	Composite DataType = 99

	// Data Structures

)
