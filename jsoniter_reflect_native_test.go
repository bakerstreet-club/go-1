package jsoniter

import (
	"testing"
	"fmt"
)

func Test_reflect_str(t *testing.T) {
	iter := ParseString(`"hello"`)
	str := ""
	iter.Read(&str)
	if str != "hello" {
		fmt.Println(iter.Error)
		t.Fatal(str)
	}
}

func Test_reflect_ptr_str(t *testing.T) {
	iter := ParseString(`"hello"`)
	var str *string
	iter.Read(&str)
	if *str != "hello" {
		t.Fatal(str)
	}
}

func Test_reflect_int(t *testing.T) {
	iter := ParseString(`123`)
	val := int(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_int8(t *testing.T) {
	iter := ParseString(`123`)
	val := int8(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_int16(t *testing.T) {
	iter := ParseString(`123`)
	val := int16(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_int32(t *testing.T) {
	iter := ParseString(`123`)
	val := int32(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_int64(t *testing.T) {
	iter := ParseString(`123`)
	val := int64(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_uint(t *testing.T) {
	iter := ParseString(`123`)
	val := uint(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_uint8(t *testing.T) {
	iter := ParseString(`123`)
	val := uint8(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_uint16(t *testing.T) {
	iter := ParseString(`123`)
	val := uint16(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_uint32(t *testing.T) {
	iter := ParseString(`123`)
	val := uint32(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_uint64(t *testing.T) {
	iter := ParseString(`123`)
	val := uint64(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_byte(t *testing.T) {
	iter := ParseString(`123`)
	val := byte(0)
	iter.Read(&val)
	if val != 123 {
		t.Fatal(val)
	}
}

func Test_reflect_float32(t *testing.T) {
	iter := ParseString(`1.23`)
	val := float32(0)
	iter.Read(&val)
	if val != 1.23 {
		fmt.Println(iter.Error)
		t.Fatal(val)
	}
}

func Test_reflect_float64(t *testing.T) {
	iter := ParseString(`1.23`)
	val := float64(0)
	iter.Read(&val)
	if val != 1.23 {
		fmt.Println(iter.Error)
		t.Fatal(val)
	}
}

func Test_reflect_bool(t *testing.T) {
	iter := ParseString(`true`)
	val := false
	iter.Read(&val)
	if val != true {
		fmt.Println(iter.Error)
		t.Fatal(val)
	}
}