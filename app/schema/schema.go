package schema

import (
	"fmt"
	"unsafe"

	"github.com/genzai-io/sliced/common/spmap"
	"github.com/genzai-io/sliced/proto/schema"
)

type Versions struct {
}

type Schema struct {
	Model schema.Schema

	Packages *spmap.Map
	Stores   *spmap.Map
	Types    *spmap.Map
	Enums    *spmap.Map
}

func NewSchema() *Schema {
	s := &Schema{}
	return s
}

func New(model *schema.Schema) *Schema {
	s := &Schema{
		Packages: spmap.New(&spmap.Options{}),
		Stores:   spmap.New(&spmap.Options{}),
		Types:    spmap.New(&spmap.Options{}),
		Enums:    spmap.New(&spmap.Options{}),
	}
	if model == nil {
		return s
	}

	if model.Packages != nil {
		for _, v := range model.Packages {
			pkg := newPackage(s, nil, v)

			// Add all sub-types
			pkg.Types.Scan(func(key string, value unsafe.Pointer) bool {
				message := (*Message)(value)
				s.Types.Set(message.Path, value)
				return true
			})
			pkg.Enums.Scan(func(key string, value unsafe.Pointer) bool {
				enum := (*Enum)(value)
				s.Enums.Set(enum.Path, value)
				return true
			})
			pkg.Packages.Scan(func(key string, value unsafe.Pointer) bool {
				pkg := (*Package)(value)
				s.Packages.Set(pkg.Path, value)
				return true
			})
		}
	}
	return s
}

//
type Package struct {
	Schema *Schema
	Parent *Package
	Path   string
	Name   string
	Model  schema.Package

	Types    *spmap.Map // Map of all types from this level down
	Enums    *spmap.Map
	Packages *spmap.Map
}

func newPackage(s *Schema, parent *Package, model *schema.Package) *Package {
	if s == nil {
		s = &Schema{}
	}
	if model == nil {
		model = &schema.Package{}
	}

	p := &Package{
		Schema:   s,
		Parent:   parent,
		Name:     model.Name,
		Model:    *model,
		Types:    spmap.New(&spmap.Options{}),
		Enums:    spmap.New(&spmap.Options{}),
		Packages: spmap.New(&spmap.Options{}),
	}

	if parent != nil {
		p.Path = fmt.Sprintf("%s.%s", parent.Path, p.Name)
	} else {
		p.Path = fmt.Sprintf("%s.%s", s.Model.Name, p.Name)
	}

	if model.Packages != nil {
		for k, v := range model.Packages {
			child := newPackage(s, p, v)
			p.Packages.Set(k, unsafe.Pointer(child))

			// Add all sub-types
			child.Types.Scan(func(key string, value unsafe.Pointer) bool {
				message := (*Message)(value)
				p.Types.Set(message.Path, value)
				return true
			})
			child.Enums.Scan(func(key string, value unsafe.Pointer) bool {
				enum := (*Enum)(value)
				p.Enums.Set(enum.Path, value)
				return true
			})
			child.Packages.Scan(func(key string, value unsafe.Pointer) bool {
				pkg := (*Package)(value)
				p.Packages.Set(pkg.Path, value)
				return true
			})
		}
	}

	if model.Enums != nil {
		for k, v := range model.Enums {
			child := newPackage(s, p, v)
			p.Packages.Set(k, unsafe.Pointer(child))
		}
	}

	return p
}

//
type Message struct {
	Package *Package
	Parent  *Message
	Path    string
	Name    string
	Model   schema.Message

	FieldList []*Field

	Fields *spmap.Map
	Nested *spmap.Map
	Enums  *spmap.Map
}

func newMessage(model *schema.Message) *Message {
	m := &Message{}

	return m
}

//
//
//
type Field struct {
	Message     *Message
	Name        string
	TypeName    string
	TypeMessage *Message
	Model       schema.Field

	Dynamic bool
}

//
//
//
type Enum struct {
	Schema  *Schema
	Package *Package
	Type    *Message
	Path    string
	Name    string

	Model schema.Enum

	Options []*EnumOption
}

func newEnum(model *schema.Enum) *Enum {
	e := &Enum{}

	return e
}

func (e *Enum) IsNested() bool { return e.Package == nil && e.Type != nil }

//
//
//
type EnumOption struct {
	Enum  *Enum
	Model schema.EnumOption
}

func newEnumOption(enum *Enum, model *schema.EnumOption) *EnumOption {
	return &EnumOption{
		Enum:  enum,
		Model: *model,
	}
}

//
//
//
type Topic struct {
}
