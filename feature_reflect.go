package jsoniter

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync/atomic"
	"unsafe"
)

/*
Reflection on type to create decoders, which is then cached
Reflection on value is avoided as we can, as the reflect.Value itself will allocate, with following exceptions
1. create instance of new value, for example *int will need a int to be allocated
2. append to slice, if the existing cap is not enough, allocate will be done using Reflect.New
3. assignment to map, both key and value will be reflect.Value
For a simple struct binding, it will be reflect.Value free and allocation free
*/

// Decoder works like a father class for sub-type decoders
type Decoder interface {
	decode(ptr unsafe.Pointer, iter *Iterator)
}

type DecoderFunc func(ptr unsafe.Pointer, iter *Iterator)
type ExtensionFunc func(typ reflect.Type, field *reflect.StructField) ([]string, DecoderFunc)

type funcDecoder struct {
	fun DecoderFunc
}

func (decoder *funcDecoder) decode(ptr unsafe.Pointer, iter *Iterator) {
	decoder.fun(ptr, iter)
}

var DECODERS unsafe.Pointer

var typeDecoders map[string]Decoder
var fieldDecoders map[string]Decoder
var extensions []ExtensionFunc

func init() {
	typeDecoders = map[string]Decoder{}
	fieldDecoders = map[string]Decoder{}
	extensions = []ExtensionFunc{}
	atomic.StorePointer(&DECODERS, unsafe.Pointer(&map[string]Decoder{}))
}

func addDecoderToCache(cacheKey reflect.Type, decoder Decoder) {
	retry := true
	for retry {
		ptr := atomic.LoadPointer(&DECODERS)
		cache := *(*map[reflect.Type]Decoder)(ptr)
		copy := map[reflect.Type]Decoder{}
		for k, v := range cache {
			copy[k] = v
		}
		copy[cacheKey] = decoder
		retry = !atomic.CompareAndSwapPointer(&DECODERS, ptr, unsafe.Pointer(&copy))
	}
}

func getDecoderFromCache(cacheKey reflect.Type) Decoder {
	ptr := atomic.LoadPointer(&DECODERS)
	cache := *(*map[reflect.Type]Decoder)(ptr)
	return cache[cacheKey]
}

// RegisterTypeDecoder can register a type for json object
func RegisterTypeDecoder(typ string, fun DecoderFunc) {
	typeDecoders[typ] = &funcDecoder{fun}
}

// RegisterFieldDecoder can register a type for json field
func RegisterFieldDecoder(typ string, field string, fun DecoderFunc) {
	fieldDecoders[fmt.Sprintf("%s/%s", typ, field)] = &funcDecoder{fun}
}

// RegisterExtension can register a custom extension
func RegisterExtension(extension ExtensionFunc) {
	extensions = append(extensions, extension)
}

// CleanDecoders cleans decoders registered
func CleanDecoders() {
	typeDecoders = map[string]Decoder{}
	fieldDecoders = map[string]Decoder{}
}

type optionalDecoder struct {
	valueType    reflect.Type
	valueDecoder Decoder
}

func (decoder *optionalDecoder) decode(ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		*((*unsafe.Pointer)(ptr)) = nil
	} else {
		if *((*unsafe.Pointer)(ptr)) == nil {
			// pointer to null, we have to allocate memory to hold the value
			value := reflect.New(decoder.valueType)
			decoder.valueDecoder.decode(unsafe.Pointer(value.Pointer()), iter)
			*((*uintptr)(ptr)) = value.Pointer()
		} else {
			// reuse existing instance
			decoder.valueDecoder.decode(*((*unsafe.Pointer)(ptr)), iter)
		}
	}
}

type mapDecoder struct {
	mapType      reflect.Type
	elemType     reflect.Type
	elemDecoder  Decoder
	mapInterface emptyInterface
}

func (decoder *mapDecoder) decode(ptr unsafe.Pointer, iter *Iterator) {
	// dark magic to cast unsafe.Pointer back to interface{} using reflect.Type
	mapInterface := decoder.mapInterface
	mapInterface.word = ptr
	realInterface := (*interface{})(unsafe.Pointer(&mapInterface))
	realVal := reflect.ValueOf(*realInterface).Elem()

	for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
		elem := reflect.New(decoder.elemType)
		decoder.elemDecoder.decode(unsafe.Pointer(elem.Pointer()), iter)
		// to put into map, we have to use reflection
		realVal.SetMapIndex(reflect.ValueOf(string([]byte(field))), elem.Elem())
	}
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  *struct{}
	word unsafe.Pointer
}

// ReadAny converts a json object in a Iterator instance to Any
func (iter *Iterator) ReadAny() (ret *Any) {
	valueType := iter.WhatIsNext()
	switch valueType {
	case String:
		return MakeAny(iter.ReadString())
	case Number:
		return iter.readNumber()
	case Null:
		return MakeAny(nil)
	case Bool:
		return MakeAny(iter.ReadBool())
	case Array:
		val := []interface{}{}
		for iter.ReadArray() {
			element := iter.ReadAny()
			if iter.Error != nil {
				return
			}
			val = append(val, element.val)
		}
		return MakeAny(val)
	case Object:
		val := map[string]interface{}{}
		for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
			element := iter.ReadAny()
			if iter.Error != nil {
				return
			}
			val[string([]byte(field))] = element.val
		}
		return MakeAny(val)
	default:
		iter.reportError("ReadAny", fmt.Sprintf("unexpected value type: %v", valueType))
		return MakeAny(nil)
	}
}

func (iter *Iterator) readNumber() (ret *Any) {
	strBuf := [8]byte{}
	str := strBuf[0:0]
	hasMore := true
	foundFloat := false
	foundNegative := false
	for hasMore {
		for i := iter.head; i < iter.tail; i++ {
			c := iter.buf[i]
			switch c {
			case '-':
				foundNegative = true
				str = append(str, c)
				continue
			case '+', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				str = append(str, c)
				continue
			case '.', 'e', 'E':
				foundFloat = true
				str = append(str, c)
				continue
			default:
				iter.head = i
				hasMore = false
				break
			}
			if !hasMore {
				break
			}
		}
		if hasMore {
			if !iter.loadMore() {
				break
			}
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		return
	}
	number := *(*string)(unsafe.Pointer(&str))
	if foundFloat {
		val, err := strconv.ParseFloat(number, 64)
		if err != nil {
			iter.Error = err
			return
		}
		return MakeAny(val)
	}
	if foundNegative {
		val, err := strconv.ParseInt(number, 10, 64)
		if err != nil {
			iter.Error = err
			return
		}
		return MakeAny(val)
	}
	val, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		iter.Error = err
		return
	}
	return MakeAny(val)
}

// Read converts an Iterator instance into go interface, same as json.Unmarshal
func (iter *Iterator) Read(obj interface{}) {
	typ := reflect.TypeOf(obj)
	cacheKey := typ.Elem()
	cachedDecoder := getDecoderFromCache(cacheKey)
	if cachedDecoder == nil {
		decoder, err := decoderOfType(typ)
		if err != nil {
			iter.Error = err
			return
		}
		cachedDecoder = decoder
		addDecoderToCache(cacheKey, decoder)
	}
	e := (*emptyInterface)(unsafe.Pointer(&obj))
	cachedDecoder.decode(e.word, iter)
}

type prefix string

func (p prefix) addTo(decoder Decoder, err error) (Decoder, error) {
	if err != nil {
		return nil, fmt.Errorf("%s: %s", p, err.Error())
	}
	return decoder, err
}

func decoderOfType(typ reflect.Type) (Decoder, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return prefix("ptr").addTo(decoderOfPtr(typ.Elem()))
	default:
		return nil, errors.New("expect ptr")
	}
}

func decoderOfPtr(typ reflect.Type) (Decoder, error) {
	typeName := typ.String()
	if typeName == "jsoniter.Any" {
		return &anyDecoder{}, nil
	}
	typeDecoder := typeDecoders[typeName]
	if typeDecoder != nil {
		return typeDecoder, nil
	}
	switch typ.Kind() {
	case reflect.String:
		return &stringDecoder{}, nil
	case reflect.Int:
		return &intDecoder{}, nil
	case reflect.Int8:
		return &int8Decoder{}, nil
	case reflect.Int16:
		return &int16Decoder{}, nil
	case reflect.Int32:
		return &int32Decoder{}, nil
	case reflect.Int64:
		return &int64Decoder{}, nil
	case reflect.Uint:
		return &uintDecoder{}, nil
	case reflect.Uint8:
		return &uint8Decoder{}, nil
	case reflect.Uint16:
		return &uint16Decoder{}, nil
	case reflect.Uint32:
		return &uint32Decoder{}, nil
	case reflect.Uint64:
		return &uint64Decoder{}, nil
	case reflect.Float32:
		return &float32Decoder{}, nil
	case reflect.Float64:
		return &float64Decoder{}, nil
	case reflect.Bool:
		return &boolDecoder{}, nil
	case reflect.Interface:
		return &interfaceDecoder{}, nil
	case reflect.Struct:
		return decoderOfStruct(typ)
	case reflect.Slice:
		return prefix("[slice]").addTo(decoderOfSlice(typ))
	case reflect.Map:
		return prefix("[map]").addTo(decoderOfMap(typ))
	case reflect.Ptr:
		return prefix("[optional]").addTo(decoderOfOptional(typ.Elem()))
	default:
		return nil, fmt.Errorf("unsupported type: %v", typ)
	}
}

func decoderOfOptional(typ reflect.Type) (Decoder, error) {
	decoder, err := decoderOfPtr(typ)
	if err != nil {
		return nil, err
	}
	return &optionalDecoder{typ, decoder}, nil
}

func decoderOfSlice(typ reflect.Type) (Decoder, error) {
	decoder, err := decoderOfPtr(typ.Elem())
	if err != nil {
		return nil, err
	}
	return &sliceDecoder{typ, typ.Elem(), decoder}, nil
}

func decoderOfMap(typ reflect.Type) (Decoder, error) {
	decoder, err := decoderOfPtr(typ.Elem())
	if err != nil {
		return nil, err
	}
	mapInterface := reflect.New(typ).Interface()
	return &mapDecoder{typ, typ.Elem(), decoder, *((*emptyInterface)(unsafe.Pointer(&mapInterface)))}, nil
}
