package table

import (
	"fmt"
	"testing"
)

func TestKey2_Compare(t *testing.T) {
	key := Key2{StringKey("0"), IntKey(0)}
	key2 := Key2{StringKey("1"), IntKey(0)}

	fmt.Println(key2.LessThan(key))
}

func TestParse(t *testing.T) {
	testFloat(t, []byte("-0193."))
	testFloat(t, []byte("-00193."))
	testFloat(t, []byte("-193."))
	testFloat(t, []byte("-0193.0"))
	testFloat(t, []byte("-193.0"))
	testFloat(t, []byte("+0193."))
	testFloat(t, []byte("+193."))
	testFloat(t, []byte("+0193.0"))
	testFloat(t, []byte("+193.0"))
	testFloat(t, []byte("0193."))
	testFloat(t, []byte("193."))
	testFloat(t, []byte("193.0"))

	testInt(t, []byte("-0193"))
	testInt(t, []byte("-00193"))
	testInt(t, []byte("-193"))
	testInt(t, []byte("-01930"))
	testInt(t, []byte("-1930"))
	testInt(t, []byte("+0193"))
	testInt(t, []byte("+193"))
	testInt(t, []byte("+01930"))
	testInt(t, []byte("+1930"))
	testInt(t, []byte("0193"))
	testInt(t, []byte("193"))
	testInt(t, []byte("1930"))

	testNil(t, nil)
	testString(t, []byte("--1"))
}

func testFloat(t *testing.T, val []byte) {
	if !isFloat(val) {
		t.Fatal(fmt.Sprintln("value not float: %s", string(val)))
	}
}

func testInt(t *testing.T, val []byte) {
	if !isInt(val) {
		t.Fatal(fmt.Sprintln("value not int: %s", string(val)))
	}
}

func testString(t *testing.T, val []byte) {
	if !isString(val) {
		t.Fatal(fmt.Sprintln("value not string: %s", string(val)))
	}
}

func testNil(t *testing.T, val []byte) {
	if !isNil(val) {
		t.Fatal(fmt.Sprintln("value not nil: %s", string(val)))
	}
}

func isFloat(val []byte) bool {
	key := ParseKeyBytes(val)
	_, ok := key.(FloatKey)
	return ok
}

func isInt(val []byte) bool {
	key := ParseKeyBytes(val)
	_, ok := key.(IntKey)
	return ok
}

func isString(val []byte) bool {
	key := ParseKeyBytes(val)
	_, ok := key.(StringKey)
	return ok
}

func isNil(val []byte) bool {
	key := ParseKeyBytes(val)
	_, ok := key.(NilKey)
	return ok
}
