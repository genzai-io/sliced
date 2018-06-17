package document

import (
	"fmt"
	"time"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/common/gjson"
	"github.com/genzai-io/sliced/common/spinlock"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

var nilMessageType = &MessageType{}

type MessageVersions struct {
}

// Represents a single type of document or "prototype".
// If a protobuf descriptor is assigned, then protobuf messages
// are supported. Otherwise, a self-describing format is the only format
// supported with initial support for JSON. Projections are supported based
// on a "Path" expression:
//
//	name.first
//  address
//  address.city
//
// Path expressions are format agnostic. Both Protobuf and JSON are completely
// interchangeable given the descriptor is set.
//
// Multiple path expressions can be listed for a "Multi" value projection.
// Indexing is supported based on Projections that can be "keyed".
type MessageType struct {
	spinlock.Locker

	Version    string
	Timestamp  time.Time

	Dynamic    bool
	Parent     *MessageType
	FQN        string
	Path       string
	Name       string
	File       *ProtoFile
	Descriptor *descriptor.DescriptorProto

	Fields     []*FieldType
	FieldTable []*FieldType
	Indexes    []*FieldType

	FieldsByName map[string]*FieldType
	Nested       map[string]*MessageType
	Enums        map[string]*EnumType
}

func newProtoMap(proto *MessageType, file *ProtoFile, list []*descriptor.DescriptorProto, m map[string]*MessageType) {
	if len(list) == 0 {
		return
	}
	for _, t := range list {
		nested := NewProto(proto, file, t)
		m[nested.Name] = nested
	}
}

func NewProto(parent *MessageType, file *ProtoFile, d *descriptor.DescriptorProto) *MessageType {
	if d == nil {
		d = &descriptor.DescriptorProto{}
	}

	p := &MessageType{
		Parent:     parent,
		File:       file,
		Descriptor: d,

		Name:   *d.Name,
		Fields: make([]*FieldType, 0, len(d.Field)),

		FieldsByName: make(map[string]*FieldType),
		Nested:       make(map[string]*MessageType),
		Enums:        make(map[string]*EnumType),
	}

	if parent != nil {
		p.FQN = fmt.Sprintf("%s.%s", parent.FQN, p.Name)
		p.Path = fmt.Sprintf("%s.%s", parent.Path, p.Name)

	} else {
		p.FQN = fmt.Sprintf(".%s.%s", *file.Descriptor.Package, p.Name)
		p.Path = p.Name
	}

	file.Messages[p.FQN] = p
	file.Messages[p.Path] = p

	newProtoMap(p, file, d.NestedType, p.Nested)
	newEnumMap(p, file, d.EnumType, p.Enums)

	if len(d.Field) > 0 {
		var max int32
		for _, f := range d.Field {
			if *f.Number > max {
				max = *f.Number
			}
			field := NewField(p, f)

			p.Fields = append(p.Fields, field)
		}

		p.FieldTable = make([]*FieldType, max+1)
		for _, f := range p.Fields {
			p.FieldTable[f.Number] = f
			p.FieldsByName[f.Name] = f
		}
	}

	return p
}

func (p *MessageType) resolve() {
	p.Lock()
	defer p.Unlock()

	for _, f := range p.Fields {
		f.resolve()
	}
}

func (p *MessageType) FieldByNumber(number int) *FieldType {
	if number < 0 || number > len(p.FieldTable) {
		return nil
	}
	return p.FieldTable[number]
}

//func (p *MessageType) Get

func ToDocument(b []byte) Document {
	if len(b) == 0 {
		return nil
	}

	if b[0] == '{' {
		// Create JSON document
		//return jsonDocument(b)
	} else {
		// Create Protobuf document
		//return protobufDocument(b)
	}

	return nil
}

//
type FieldType struct {
	dynamic    bool
	Parent     *MessageType
	Descriptor *descriptor.FieldDescriptorProto
	Label      descriptor.FieldDescriptorProto_Label

	Name     string
	Number   int32
	JsonName string
	SqlName  string

	Message *MessageType
	Enum    *EnumType

	Type         moved.DataType
	JsonType     gjson.Type
	ProtobufType descriptor.FieldDescriptorProto_Type

	//JsonProjector     jsonProjector
	//ProtobufProjector protobufProjector
}

func NewField(parent *MessageType, descriptor *descriptor.FieldDescriptorProto) *FieldType {
	f := &FieldType{
		Parent:     parent,
		Descriptor: descriptor,
		Name:       *descriptor.Name,
		Number:     *descriptor.Number,
		JsonName:   *descriptor.JsonName,

		ProtobufType: *descriptor.Type,
		JsonType:     protobufTypeToJSONType(*descriptor.Type),
		Type:         pbufTypeToSlicedType(*descriptor.Type),

		//JsonProjector:     jsonProjector(*descriptor.JsonName),
		//ProtobufProjector: protobufProjector(*descriptor.Number),
	}

	if f.JsonName == "" {
		f.JsonName = f.Name
	}

	return f
}

func (f *FieldType) resolve() {
	switch f.ProtobufType {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		typeName := *f.Descriptor.TypeName
		f.Message, _ = f.Parent.File.Messages[typeName]

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		typeName := *f.Descriptor.TypeName
		f.Enum, _ = f.Parent.File.Enums[typeName]
	}
}

func (f *FieldType) IsDynamic() bool { return f.Descriptor == nil }

func (f *FieldType) IsRepeated() bool { return f.Descriptor.IsRepeated() }

//
type EnumType struct {
	Path   string
	FQN    string
	Proto  *MessageType
	File   *ProtoFile
	Name   string
	Values []*EnumValue
}

//
type EnumValue struct {
	Enum   *EnumType
	Index  int32
	Name   string
	Number int32
}

func newEnumMap(proto *MessageType, file *ProtoFile, list []*descriptor.EnumDescriptorProto, m map[string]*EnumType) {
	if len(list) == 0 {
		return
	}
	for _, e := range list {
		m[*e.Name] = NewEnum(proto, file, e)
	}
}

func NewEnum(proto *MessageType, file *ProtoFile, d *descriptor.EnumDescriptorProto) *EnumType {
	if d == nil {
		d = &descriptor.EnumDescriptorProto{}
	}

	enum := &EnumType{
		Proto: proto,
		File:  file,
		Name:  *d.Name,
	}

	if proto != nil {
		enum.FQN = fmt.Sprintf("%s.%s", proto.FQN, enum.Name)
		enum.Path = fmt.Sprintf("%s.%s", proto.Path, enum.Name)

	} else {
		enum.FQN = fmt.Sprintf(".%s.%s", *file.Descriptor.Package, enum.Name)
		enum.Path = enum.Name
	}

	file.Enums[enum.FQN] = enum
	file.Enums[enum.Path] = enum

	if len(d.Value) > 0 {
		enum.Values = make([]*EnumValue, len(d.Value))
		for i, v := range d.Value {
			enum.Values[i] = &EnumValue{
				Enum:   enum,
				Index:  int32(i),
				Name:   *v.Name,
				Number: *v.Number,
			}
		}
	}

	return enum
}

//
type ProtoProjection interface {
	Fields() int

	FieldAt(i int) FieldType
}

func protobufTypeToJSONType(t descriptor.FieldDescriptorProto_Type) gjson.Type {
	switch t {
	// 0 is reserved for errors.
	// Order is weird for historical reasons.
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return gjson.Number

	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return gjson.Number
		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT64 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return gjson.Number

	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return gjson.Number

		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT32 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return gjson.Number

	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return gjson.Number

	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return gjson.Number

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return gjson.Bool

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return gjson.String

		// Tag-delimited aggregate.
		// Group type is deprecated and not supported in proto3. However, Proto3
		// implementations should still be able to parse the group wire format and
		// treat group fields as unknown fields.
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return gjson.String

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return gjson.JSON

		// New in version 2.
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return gjson.String

	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return gjson.Number

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return gjson.Number
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return gjson.Number
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return gjson.Number
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return gjson.Number
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return gjson.Number
	}
	return gjson.Null
}

func pbufTypeToSlicedType(t descriptor.FieldDescriptorProto_Type) moved.DataType {
	switch t {
	// 0 is reserved for errors.
	// Order is weird for historical reasons.
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return moved.Float

	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return moved.Float32
		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT64 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return moved.Int

	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return moved.Uint

		// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT32 if
		// negative values are likely.
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return moved.Int

	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return moved.Int

	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return moved.Int32

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return moved.Bool

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return moved.String

		// Tag-delimited aggregate.
		// Group type is deprecated and not supported in proto3. However, Proto3
		// implementations should still be able to parse the group wire format and
		// treat group fields as unknown fields.
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return moved.String

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return moved.Protobuf

		// New in version 2.
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return moved.String

	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return moved.Uint32

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return moved.Int16
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return moved.Int32
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return moved.Int
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return moved.Int32
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return moved.Int
	}
	return moved.Nil
}
