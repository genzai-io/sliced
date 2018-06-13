package table

import (
	"github.com/genzai-io/sliced/common/gjson"
)

type IndexOpts int

const (
	IncludeString   IndexOpts = 1 << iota
	IncludeInt
	IncludeFloat
	IncludeBool
	IncludeRect
	IncludeNil
	IncludeAny
	IncludeFloatAsInt
	CaseInsensitive
	SortDesc
)

// Project as single key from a value
type KeyProjector func(item *ValueItem) Key

//
type Indexer interface {
	// Parses raw RESP args
	ParseArgs(offset int, buf [][]byte) Key
	// Parse a single RESP arg
	ParseArg(arg []byte) Key
	// Number of fields in index
	Fields() int
	// Retrieve an IndexField at a position
	FieldAt(pos int) *IndexField
	// Factory function
	Index(index *Index, item *ValueItem) IndexItem
}

//
//
//
func ValueProjector(item *ValueItem) Key {
	l := len(item.Value)
	if l == 0 {
		return StringMin
	}

	if item.Value[0] == '{' {
		return StringKey(item.Value)
	} else {
		return ParseKey(item.Value)
	}
}

//
//
//
func RectProjector(item *ValueItem) Key {
	return ParseRect(item.Value)
}

//
//
//
func JSONProjector(path string) KeyProjector {
	return func(item *ValueItem) Key {
		result := gjson.Get(item.Value, path)
		switch result.Type {
		case gjson.String:
			return StringKey(result.Str)
		case gjson.Null:
			return Nil
		case gjson.JSON:
			return Nil
		case gjson.True:
			return True
		case gjson.False:
			return False
		case gjson.Number:
			return FloatKey(result.Num)
		default:
			return StringKey(result.Raw)
		}
	}
}

//
//
//
func JSONRectProjector(path string) KeyProjector {
	return func(item *ValueItem) Key {
		result := gjson.Get(item.Value, path)

		var rect Rect
		if result.Type == gjson.String {
			rect = ParseRect(result.Str)
		} else if result.Type == gjson.JSON {
			if result.Raw[0] == '[' {
				rect = ParseRect(result.Raw)
			} else {
				return SkipKey
			}
		} else {
			return SkipKey
		}

		if rect.Min == nil && rect.Max == nil {
			return SkipKey
		}
		return rect
	}
}

//
//
//
func JSONIndexer(path string, opts IndexOpts) *IndexField {
	return NewIndexer(path, opts, JSONProjector(path))
}

//
//
//
func JSONComposite(fields... *IndexField) *CompositeIndex {
	return &CompositeIndex{fields: fields}
}

//
//
//
func IndexString(desc bool) IndexOpts {
	o := IncludeString
	if desc {
		o |= SortDesc
	}
	return o
}

//
//
//
func IndexInt(desc bool) IndexOpts {
	o := IncludeInt | IncludeFloat | IncludeFloatAsInt
	if desc {
		o |= SortDesc
	}
	return o
}

//
//
//
func IndexFloat(desc bool) IndexOpts {
	o := IncludeInt | IncludeFloat
	if desc {
		o |= SortDesc
	}
	return o
}

//
//
//
func IndexAny(desc bool) IndexOpts {
	o := IncludeInt | IncludeFloat | IncludeString | IncludeNil
	if desc {
		o |= SortDesc
	}
	return o
}

//
//
//
func IndexSpatial() IndexOpts {
	return IncludeRect
}

func StringIndexer() *IndexField {
	return NewIndexer("", IndexString(false), ValueProjector)
}

//
//
//
func SpatialIndexer() *IndexField {
	return NewIndexer("", IndexSpatial(), RectProjector)
}

//
//
//
func JSONSpatialIndexer(path string) *IndexField {
	return NewIndexer(path, IndexSpatial(), JSONRectProjector(path))
}

//
//
//
func NewIndexer(
	name string,
	opts IndexOpts,
	projector KeyProjector,
) *IndexField {
	return &IndexField{
		name:      name,
		opts:      opts,
		projector: projector,
	}
}

// Meta data to describe the behavior of an index dimension
type IndexField struct {
	name      string
	length    int
	opts      IndexOpts
	projector KeyProjector
}

func (i *IndexField) ParseArgs(offset int, buf [][]byte) Key {
	if len(buf) < offset {
		return Nil
	}
	return i.ParseArg(buf[offset])
}

func (i *IndexField) ParseArg(buf []byte) Key {
	switch key := ParseKeyBytes(buf).(type) {
	case IntKey:
		if i.opts&IncludeInt != 0 {
			return key
		} else {
			return SkipKey
		}
	case StringKey:
		if i.opts&IncludeString != 0 {
			return key
		} else {
			return SkipKey
		}
	case FloatKey:
		if i.opts&IncludeFloat != 0 {
			return key
		} else {
			return SkipKey
		}
	default:
		return key
	}
}

func (i *IndexField) Fields() int {
	return 1
}

func (i *IndexField) FieldAt(index int) *IndexField {
	if index == 0 {
		return i
	} else {
		return nil
	}
}

func (i *IndexField) Key(index *Index, item *ValueItem) Key {
	// Project a key from the value
	val := i.projector(item)

	// Should we skip?
	if val == SkipKey {
		return nil
	}

	// The general logic is duplicated with "K()"
	// Only done to save a type assertion for the most
	// likely path of a single field index.
	switch key := val.(type) {
	default:
		return key
	case IntKey:
		if i.opts&IncludeInt != 0 {
			if i.opts&SortDesc != 0 {
				return IntDescKey(key)
			} else {
				return key
			}
		} else {
			return nil
		}
	case IntDescKey:
		if i.opts&IncludeInt != 0 {
			return key
		} else {
			return nil
		}
	case FloatKey:
		if i.opts&IncludeFloat != 0 {
			if i.opts&IncludeFloatAsInt != 0 {
				if i.opts&SortDesc != 0 {
					return IntDescKey(key)
				} else {
					return IntKey(key)
				}
			} else {
				if i.opts&SortDesc != 0 {
					return FloatDescKey(key)
				} else {
					return key
				}
			}
		} else {
			return nil
		}
	case FloatDescKey:
		if i.opts&IncludeFloat != 0 {
			if i.opts&IncludeFloatAsInt != 0 {
				if i.opts&SortDesc != 0 {
					return IntDescKey(key)
				} else {
					return IntKey(key)
				}
			} else {
				if i.opts&SortDesc != 0 {
					return FloatDescKey(key)
				} else {
					return key
				}
			}
		} else {
			return nil
		}
	case StringKey:
		if i.opts&IncludeString != 0 {
			// Truncate if necessary
			if i.length > 0 && len(key) > i.length {
				key = key[:i.length]
			}
			if i.opts&SortDesc != 0 {
				if i.opts&CaseInsensitive != 0 {
					return StringCIDescKey(key)
				} else {
					return StringDescKey(key)
				}
			} else {
				if i.opts&CaseInsensitive != 0 {
					return StringCIKey(key)
				} else {
					return key
				}
			}
		} else {
			return nil
		}

	case NilKey:
		if i.opts&IncludeNil != 0 {
			return key
		} else {
			return nil
		}
	case Rect:
		if i.opts&IncludeRect != 0 {
			return key
		} else {
			return nil
		}
	}
}

func (i *IndexField) Index(index *Index, item *ValueItem) IndexItem {
	// Project a key from the value
	val := i.projector(item)

	// Should we skip?
	if val == SkipKey {
		return nil
	}

	// The general logic is duplicated with "Extract()"
	// Only done to save a type assertion for the most
	// likely path of a single field index.
	switch key := val.(type) {
	default:
		return nil
	case IntKey:
		if i.opts&IncludeInt != 0 {
			if i.opts&SortDesc != 0 {
				return &intDescItem{
					indexItem: indexItem{
						idx:   index,
						value: item,
					},
					key: IntDescKey(key),
				}
			} else {
				return &intItem{
					indexItem: indexItem{
						idx:   index,
						value: item,
					},
					key: key,
				}
			}
		} else {
			return nil
		}
	case IntDescKey:
		if i.opts&IncludeInt != 0 {
			return &intDescItem{
				indexItem: indexItem{
					idx:   index,
					value: item,
				},
				key: key,
			}
		} else {
			return nil
		}
	case FloatKey:
		if i.opts&IncludeFloat != 0 {
			if i.opts&IncludeFloatAsInt != 0 {
				if i.opts&SortDesc != 0 {
					return &intDescItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: IntDescKey(key),
					}
				} else {
					return &intItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: IntKey(key),
					}
				}
			} else {
				if i.opts&SortDesc != 0 {
					return &floatDescItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: FloatDescKey(key),
					}
				} else {
					return &floatItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: key,
					}
				}
			}
		} else {
			return nil
		}
	case StringKey:
		if i.opts&IncludeString != 0 {
			// Truncate if necessary
			if i.length > 0 && len(key) > i.length {
				key = key[:i.length]
			}
			if i.opts&SortDesc != 0 {
				if i.opts&CaseInsensitive != 0 {
					return &stringCIDescItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: StringCIDescKey(key),
					}
				} else {
					return &stringDescItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: StringDescKey(key),
					}
				}
			} else {
				if i.opts&CaseInsensitive != 0 {
					return &stringCIItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: StringCIKey(key),
					}
				} else {
					return &stringItem{
						indexItem: indexItem{
							idx:   index,
							value: item,
						},
						key: key,
					}
				}
			}
		} else {
			return nil
		}

	case NilKey:
		if i.opts&IncludeNil != 0 {
			return &nilItem{
				indexItem: indexItem{
					idx:   index,
					value: item,
				},
			}
		} else {
			return nil
		}
	case Rect:
		if i.opts&IncludeRect != 0 {
			return &rectItem{
				indexItem: indexItem{
					idx:   index,
					value: item,
				},
				key: key,
			}
		} else {
			return nil
		}
	}
}

//
//
//
type CompositeIndex struct {
	fields []*IndexField
}

func (i *CompositeIndex) ParseArgs(offset int, buf [][]byte) Key {
	switch len(i.fields) {
	default:
		return SkipKey
	case 1:
		return i.fields[0].ParseArgs(offset, buf)
	case 2:
		return Key2{
			i.fields[0].ParseArgs(offset, buf),
			i.fields[2].ParseArgs(offset+1, buf),
		}
	}
}

func (i *CompositeIndex) ParseArg(buf []byte) Key {
	return Nil
}

func (i *CompositeIndex) Fields() int {
	return len(i.fields)
}

func (i *CompositeIndex) FieldAt(index int) *IndexField {
	if index < 0 || index > len(i.fields) {
		return nil
	}
	return i.fields[index]
}

func (i *CompositeIndex) Index(index *Index, item *ValueItem) IndexItem {
	l := len(i.fields)
	if l == 0 {
		return nil
	}
	switch len(i.fields) {
	case 0:
		return nil
	case 1:
		return i.fields[0].Index(index, item)
	case 2:
		key := i.fields[0].Key(index, item)
		if key == nil || key == SkipKey {
			return nil
		}
		key2 := i.fields[1].Key(index, item)
		if key2 == nil || key2 == SkipKey {
			return nil
		}
		return &key2Item{
			indexItem: indexItem{
				idx:   index,
				value: item,
			},
			key: Key2{key, key2},
		}

	default:
		return nil
	}
}
