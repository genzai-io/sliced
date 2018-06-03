package btrdb

import (
	"encoding/binary"

	"github.com/genzai-io/sliced/common/gjson"
)

func Contains(tx *Tx, key string) bool {
	_, err := tx.Get(key)
	return err != ErrNotFound
}

func ChooseLess(val interface{}) func(a, b string) bool {
	switch v := val.(type) {
	case int, int32, uint32, int64, uint64:
		return IndexInt
	case string:
		return IndexString
	default:
		return IndexBinary
	}
}

func FormatInt(val uint64) string {
	b := make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(b, val)
	return string(b)
}

//
func FormatJSON(val gjson.Result) string {
	switch val.Type {
	case gjson.String:
		return val.Str

	case gjson.Number:
		b := make([]byte, 8, 8)
		binary.LittleEndian.PutUint64(b, uint64(val.Num))
		return string(b)

	case gjson.Null:
		return ""

	case gjson.JSON:
		return val.Raw

	default:
		return val.Raw
	}
}

func Format(val interface{}) string {
	switch v := val.(type) {
	case int:
		return FormatInt(uint64(v))
	case int64:
		return FormatInt(uint64(v))
	case uint64:
		return FormatInt(v)
	case float64:
		return FormatInt(uint64(v))
	case int32:
		return FormatInt(uint64(v))
	case uint32:
		return FormatInt(uint64(v))
	case float32:
		return FormatInt(uint64(v))

	case []byte:
		return string(v)

	case string:
		return v
	}
	return ""
}

func CastInt64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case uint32:
		return int64(v), true
	case int64:
		return int64(v), true
	case uint64:
		return int64(v), true
	}
	return 0, false
}

func IsInt(val interface{}) bool {
	switch val.(type) {
	case int, int32, uint32, int64, uint64:
		return true
	}
	return false
}
