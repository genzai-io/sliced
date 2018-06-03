package table

import (
	"math"
	"strconv"
	"unsafe"
	"encoding/binary"
	"reflect"
)

func IntToString(n int) string {
	buf := make([]byte, 4, 4)
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)
	return string(buf)
}

func Int32ToString(n int32) string {
	buf := make([]byte, 4, 4)
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)
	return string(buf)
}

func UInt32ToString(n uint32) string {
	buf := make([]byte, 4, 4)
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)
	return string(buf)
}

func Int64ToString(n int64) string {
	buf := make([]byte, 8, 8)
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	return string(buf)
}

func UInt64ToString(n uint64) string {
	buf := make([]byte, 8, 8)
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	return string(buf)
}

func EncodeInt64(num string) string {
	buf := make([]byte, 8, 8)
	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		n = 0
	}
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	return string(buf)
}

func EncodeFloat(num string) string {
	buf := make([]byte, 8, 8)
	nu, err := strconv.ParseFloat(num, 64)
	if err != nil {
		nu = 0.0
	}
	n := math.Float64bits(nu)
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	return string(buf)
}

func EncodeFloat64(num string) string {
	buf := make([]byte, 8, 8)
	nu, err := strconv.ParseFloat(num, 64)
	if err != nil {
		nu = 0.0
	}
	//Endian.PutUint64(buf, uint64(nu))
	n := math.Float64bits(nu)
	Endian.PutUint64(buf, n)
	//buf[0] = byte(n >> 56)
	//buf[1] = byte(n >> 48)
	//buf[2] = byte(n >> 40)
	//buf[3] = byte(n >> 32)
	//buf[4] = byte(n >> 24)
	//buf[5] = byte(n >> 16)
	//buf[6] = byte(n >> 8)
	//buf[7] = byte(n)
	return string(buf)
}

func StrAsFloat64(str string) float64 {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&str))
	return math.Float64frombits(*(*uint64)(unsafe.Pointer(hdr.Data)))
}

func FloatToString(num float64) string {
	buf := make([]byte, 8, 8)
	n := math.Float64bits(num)
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	return string(buf)
}

type SmartType uintptr

var Endian binary.ByteOrder
var bigEndian bool

func IsBigEndian() bool {
	return bigEndian
}

func IsLittleEndian() bool {
	return !bigEndian
}

func init() {
	if getEndian() {
		Endian = binary.BigEndian
		bigEndian = true
	} else {
		Endian = binary.LittleEndian
		bigEndian = false
	}
}

//
const INT_SIZE int = int(unsafe.Sizeof(0))

//true = big endian, false = little endian
func getEndian() (ret bool) {
	var i int = 0x1
	bs := (*[INT_SIZE]byte)(unsafe.Pointer(&i))
	if bs[0] == 0 {
		return true
	} else {
		return false
	}

}
