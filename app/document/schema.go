package document

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

//
//
//
type ProtoFile struct {
	Name string
	FQN  string

	Hash       string
	GzipHash   string
	Descriptor *descriptor.FileDescriptorProto
	Messages   map[string]*MessageType
	Enums      map[string]*EnumType
}

func NewProtoFile(hash, gzipHash string, descriptor *descriptor.FileDescriptorProto) *ProtoFile {
	file := &ProtoFile{
		Name:       *descriptor.Name,
		Hash:       hash,
		GzipHash:   gzipHash,
		Descriptor: descriptor,
		Messages:   make(map[string]*MessageType),
		Enums:      make(map[string]*EnumType),
	}

	newProtoMap(nil, file, descriptor.MessageType, file.Messages)
	newEnumMap(nil, file, descriptor.EnumType, file.Enums)

	for _, m := range file.Messages {
		m.resolve()
	}

	return file
}


