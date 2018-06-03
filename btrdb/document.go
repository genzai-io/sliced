package btrdb

type DocumentType byte

const (
	JSON     DocumentType = 0
	Protobuf DocumentType = 1
)

type Serializable interface {
	Marshaler
	Unmarshaler
}

// Marshaler is the interface representing objects that can marshal themselves.
type Marshaler interface {
	Marshal() ([]byte, error)
}

// Unmarshaler is the interface representing objects that can
// unmarshal themselves.  The method should reset the receiver before
// decoding starts.  The argument points to data that may be
// overwritten, so implementations should not keep references to the
// buffer.
type Unmarshaler interface {
	Unmarshal([]byte) error
}

type FieldIterator interface {

}

type Parser interface {

}

type Factory func() Serializable

type Document struct {
	Type  DocumentType
	Value string
}
