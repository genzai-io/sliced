package table

import (
	"fmt"

	"github.com/slice-d/genzai"
	"github.com/slice-d/genzai/app/table/index/btree"
)

// Composite keys have 2 or more keys and the keys cannot be other composite keys
//
// Composite Extract with 2 Elements
type Key2 struct {
	_1 Key
	_2 Key
}

func (k Key2) CanIndex() bool {
	return k._1.CanIndex() && k._2.CanIndex()
}
func (k Key2) Keys() int { return 2 }
func (k Key2) KeyAt(index int) Key {
	switch index {
	case 0:
		return k._1
	case 1:
		return k._2
	}
	return SkipKey
}
func (k Key2) Type() moved.DataType {
	return moved.Any
}
func (k Key2) Match(pattern string) bool {
	return k._1.Match(pattern)
}
func (k Key2) Less(than btree.Item, ctx interface{}) bool {
	switch t := than.(type) {
	case *key2Item:
		switch k._1.Compare(t.key._1) {
		case -1:
			return true
		case 1:
			return false
		}
		switch k._2.Compare(t.key._2) {
		case -1:
			return true
		case 1:
			return false
		default:
			return false
		}
	case Key2:
		if k._1.LessThan(t._1) {
			return true
		}
		return k._2.LessThan(t._2)
	case *Key2:
		if k._1.LessThan(t._1) {
			return true
		}
		return k._2.LessThan(t._2)
		//case Key3:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case *Key3:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case Key4:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case *Key4:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case Key5:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case *Key5:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
	default:
		return false
	}
	return false
}
func (k Key2) LessThan(than Key) bool {
	switch t := than.(type) {
	case Key2:
		if k._1.LessThan(t._1) {
			return true
		}
		return k._2.LessThan(t._2)
	case *Key2:
		if k._1.LessThan(t._1) {
			return true
		}
		return k._2.LessThan(t._2)
		//case Key3:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case *Key3:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case Key4:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case *Key4:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case Key5:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
		//case *Key5:
		//	if k._1.LessThan(t._1) {
		//		return true
		//	}
		//	return k._2.LessThan(t._2)
	default:
		return false
	}
}
func (k Key2) LessThanItem(than btree.Item, item *ValueItem) bool {
	switch to := than.(type) {
	case *key2Item:
		switch k._1.Compare(to.key._1) {
		case -1:
			return true
		case 1:
			return false
		}
		switch k._2.Compare(to.key._2) {
		case -1:
			return true
		case 1:
			return false
		default:
			if item == nil {
				return to.value != nil
			} else if to.value == nil {
				return true
			} else {
				return item.Key.LessThan(to.value.Key)
			}
		}
	case Key2:
		if k._1.LessThan(to._1) {
			return true
		}
		return k._2.LessThan(to._2)
	case *Key2:
		if k._1.LessThan(to._1) {
			return true
		}
		return k._2.LessThan(to._2)
		//case Key3:
		//	if k._1.LessThan(to._1) {
		//		return true
		//	}
		//	return k._2.LessThan(to._2)
		//case *Key3:
		//	if k._1.LessThan(to._1) {
		//		return true
		//	}
		//	return k._2.LessThan(to._2)
		//case Key4:
		//	if k._1.LessThan(to._1) {
		//		return true
		//	}
		//	return k._2.LessThan(to._2)
		//case *Key4:
		//	if k._1.LessThan(to._1) {
		//		return true
		//	}
		//	return k._2.LessThan(to._2)
		//case Key5:
		//	if k._1.LessThan(to._1) {
		//		return true
		//	}
		//	return k._2.LessThan(to._2)
		//case *Key5:
		//	if k._1.LessThan(to._1) {
		//		return true
		//	}
		//	return k._2.LessThan(to._2)

	case *nilItem:
		return false
	case *trueItem:
		return k._1.LessThan(True)
	case *falseItem:
		return k._1.LessThan(False)
	case *intItem:
		return k._1.LessThan(to.key)
	case *intDescItem:
		return k._1.LessThan(to.key)
	case *floatItem:
		return k._1.LessThan(to.key)
	case *floatDescItem:
		return k._1.LessThan(to.key)
	case *stringItem:
		return k._1.LessThan(to.key)
	case *stringDescItem:
		return k._1.LessThan(to.key)
	case *stringCIItem:
		return k._1.LessThan(to.key)
	case *stringCIDescItem:
		return k._1.LessThan(to.key)
	}
	return false
}
func (k Key2) Compare(key Key) int {
	return 1
}
func (k Key2) String() string {
	return fmt.Sprintf("(%s, %s)", k._1, k._2)
}

//
//
//
type Key3 struct {
	_1 Key
	_2 Key
	_3 Key
}

//
//
//
type Key4 struct {
	_1 Key
	_2 Key
	_3 Key
	_4 Key
}

//
//
//
type Key5 struct {
	_1 Key
	_2 Key
	_3 Key
	_4 Key
	_5 Key
}
