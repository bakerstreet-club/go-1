package jsoniter

import (
	"testing"
	"github.com/json-iterator/go/require"
)

func Test_decode_one_field_struct(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		field1 string
	}
	obj := TestObject{}
	should.Nil(UnmarshalString(`{}`, &obj))
	should.Equal("", obj.field1)
	should.Nil(UnmarshalString(`{"field1": "hello"}`, &obj))
	should.Equal("hello", obj.field1)
}

func Test_decode_two_fields_struct(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		field1 string
		field2 string
	}
	obj := TestObject{}
	should.Nil(UnmarshalString(`{}`, &obj))
	should.Equal("", obj.field1)
	should.Nil(UnmarshalString(`{"field1": "a", "field2": "b"}`, &obj))
	should.Equal("a", obj.field1)
	should.Equal("b", obj.field2)
}

func Test_decode_three_fields_struct(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		field1 string
		field2 string
		field3 string
	}
	obj := TestObject{}
	should.Nil(UnmarshalString(`{}`, &obj))
	should.Equal("", obj.field1)
	should.Nil(UnmarshalString(`{"field1": "a", "field2": "b", "field3": "c"}`, &obj))
	should.Equal("a", obj.field1)
	should.Equal("b", obj.field2)
	should.Equal("c", obj.field3)
}

func Test_decode_four_fields_struct(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		field1 string
		field2 string
		field3 string
		field4 string
	}
	obj := TestObject{}
	should.Nil(UnmarshalString(`{}`, &obj))
	should.Equal("", obj.field1)
	should.Nil(UnmarshalString(`{"field1": "a", "field2": "b", "field3": "c", "field4": "d"}`, &obj))
	should.Equal("a", obj.field1)
	should.Equal("b", obj.field2)
	should.Equal("c", obj.field3)
	should.Equal("d", obj.field4)
}

func Test_decode_five_fields_struct(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		field1 string
		field2 string
		field3 string
		field4 string
		field5 string
	}
	obj := TestObject{}
	should.Nil(UnmarshalString(`{}`, &obj))
	should.Equal("", obj.field1)
	should.Nil(UnmarshalString(`{"field1": "a", "field2": "b", "field3": "c", "field4": "d", "field5": "e"}`, &obj))
	should.Equal("a", obj.field1)
	should.Equal("b", obj.field2)
	should.Equal("c", obj.field3)
	should.Equal("d", obj.field4)
	should.Equal("e", obj.field5)
}

func Test_decode_struct_with_optional_field(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		field1 *string
		field2 *string
	}
	obj := TestObject{}
	UnmarshalString(`{"field1": null, "field2": "world"}`, &obj)
	should.Nil(obj.field1)
	should.Equal("world", *obj.field2)
}

func Test_decode_struct_field_with_tag(t *testing.T) {
	should := require.New(t)
	type TestObject struct {
		Field1 string `json:"field-1"`
		Field2 string `json:"-"`
		Field3 int    `json:",string"`
	}
	obj := TestObject{Field2: "world"}
	UnmarshalString(`{"field-1": "hello", "field2": "", "Field3": "100"}`, &obj)
	should.Equal("hello", obj.Field1)
	should.Equal("world", obj.Field2)
	should.Equal(100, obj.Field3)
}