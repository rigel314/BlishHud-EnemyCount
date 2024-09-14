package main

import (
	"encoding/binary"
	"errors"
	"math"
	"reflect"
)

// Encode encodes the binary representation of data into buf according to
// the given byte order.
// It returns an error if buf is too small, otherwise the number of
// bytes written into buf.
func Encode(buf []byte, order binary.ByteOrder, data any) (int, error) {
	// Fast path for basic types and slices.
	// if n, _ := intDataSize(data); n != 0 {
	// 	if len(buf) < n {
	// 		return 0, errBufferTooSmall
	// 	}

	// 	encodeFast(buf, order, data)
	// 	return n, nil
	// }

	// Fallback to reflect-based encoding.
	v := reflect.Indirect(reflect.ValueOf(data))
	size := dataSize(v)
	if size < 0 {
		return 0, errors.New("binary.Encode: some values are not fixed-sized in type " + reflect.TypeOf(data).String())
	}

	if len(buf) < size {
		return 0, errBufferTooSmall
	}
	e := &encoder{order: order, buf: buf}
	e.value(v)
	return size, nil
}

// intDataSize returns the size of the data required to represent the data when encoded,
// and optionally a byte slice containing the encoded data if no conversion is necessary.
// It returns zero, nil if the type cannot be implemented by the fast path in Read or Write.
// func intDataSize(data any) (int, []byte) {
// 	switch data := data.(type) {
// 	case bool, int8, uint8, *bool, *int8, *uint8:
// 		return 1, nil
// 	case []bool:
// 		return len(data), nil
// 	case []int8:
// 		return len(data), nil
// 	case []uint8:
// 		return len(data), data
// 	case int16, uint16, *int16, *uint16:
// 		return 2, nil
// 	case []int16:
// 		return 2 * len(data), nil
// 	case []uint16:
// 		return 2 * len(data), nil
// 	case int32, uint32, *int32, *uint32:
// 		return 4, nil
// 	case []int32:
// 		return 4 * len(data), nil
// 	case []uint32:
// 		return 4 * len(data), nil
// 	case int64, uint64, *int64, *uint64:
// 		return 8, nil
// 	case []int64:
// 		return 8 * len(data), nil
// 	case []uint64:
// 		return 8 * len(data), nil
// 	case float32, *float32:
// 		return 4, nil
// 	case float64, *float64:
// 		return 8, nil
// 	case []float32:
// 		return 4 * len(data), nil
// 	case []float64:
// 		return 8 * len(data), nil
// 	case string:
// 		return 8 + len(data), nil
// 	case *string:
// 		if data == nil {
// 			return 0, nil
// 		} else {
// 			return 8 + len(*data), nil
// 		}
// 	}
// 	return 0, nil
// }

var errBufferTooSmall = errors.New("buffer too small")

// func encodeFast(bs []byte, order binary.ByteOrder, data any) {
// 	switch v := data.(type) {
// 	case *bool:
// 		if *v {
// 			bs[0] = 1
// 		} else {
// 			bs[0] = 0
// 		}
// 	case bool:
// 		if v {
// 			bs[0] = 1
// 		} else {
// 			bs[0] = 0
// 		}
// 	case []bool:
// 		for i, x := range v {
// 			if x {
// 				bs[i] = 1
// 			} else {
// 				bs[i] = 0
// 			}
// 		}
// 	case *int8:
// 		bs[0] = byte(*v)
// 	case int8:
// 		bs[0] = byte(v)
// 	case []int8:
// 		for i, x := range v {
// 			bs[i] = byte(x)
// 		}
// 	case *uint8:
// 		bs[0] = *v
// 	case uint8:
// 		bs[0] = v
// 	case []uint8:
// 		copy(bs, v)
// 	case *int16:
// 		order.PutUint16(bs, uint16(*v))
// 	case int16:
// 		order.PutUint16(bs, uint16(v))
// 	case []int16:
// 		for i, x := range v {
// 			order.PutUint16(bs[2*i:], uint16(x))
// 		}
// 	case *uint16:
// 		order.PutUint16(bs, *v)
// 	case uint16:
// 		order.PutUint16(bs, v)
// 	case []uint16:
// 		for i, x := range v {
// 			order.PutUint16(bs[2*i:], x)
// 		}
// 	case *int32:
// 		order.PutUint32(bs, uint32(*v))
// 	case int32:
// 		order.PutUint32(bs, uint32(v))
// 	case []int32:
// 		for i, x := range v {
// 			order.PutUint32(bs[4*i:], uint32(x))
// 		}
// 	case *uint32:
// 		order.PutUint32(bs, *v)
// 	case uint32:
// 		order.PutUint32(bs, v)
// 	case []uint32:
// 		for i, x := range v {
// 			order.PutUint32(bs[4*i:], x)
// 		}
// 	case *int64:
// 		order.PutUint64(bs, uint64(*v))
// 	case int64:
// 		order.PutUint64(bs, uint64(v))
// 	case []int64:
// 		for i, x := range v {
// 			order.PutUint64(bs[8*i:], uint64(x))
// 		}
// 	case *uint64:
// 		order.PutUint64(bs, *v)
// 	case uint64:
// 		order.PutUint64(bs, v)
// 	case []uint64:
// 		for i, x := range v {
// 			order.PutUint64(bs[8*i:], x)
// 		}
// 	case *float32:
// 		order.PutUint32(bs, math.Float32bits(*v))
// 	case float32:
// 		order.PutUint32(bs, math.Float32bits(v))
// 	case []float32:
// 		for i, x := range v {
// 			order.PutUint32(bs[4*i:], math.Float32bits(x))
// 		}
// 	case *float64:
// 		order.PutUint64(bs, math.Float64bits(*v))
// 	case float64:
// 		order.PutUint64(bs, math.Float64bits(v))
// 	case []float64:
// 		for i, x := range v {
// 			order.PutUint64(bs[8*i:], math.Float64bits(x))
// 		}
// 	case string:
// 		order.PutUint64(bs, uint64(len(v)))
// 		copy(bs[8:], ([]byte)(v))
// 	case *string:
// 		if v == nil {
// 			break
// 		}
// 		order.PutUint64(bs, uint64(len(*v)))
// 		copy(bs[8:], ([]byte)(*v))
// 	}
// }

// dataSize returns the number of bytes the actual data represented by v occupies in memory.
// For compound structures, it sums the sizes of the elements. Thus, for instance, for a slice
// it returns the length of the slice times the element size and does not count the memory
// occupied by the header. If the type of v is not acceptable, dataSize returns -1.
func dataSize(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		size := sizeof(v)
		if size >= 0 {
			return size * v.Len()
		}

	case reflect.Struct:
		size := sizeof(v)
		return size

	case reflect.String:
		return 8 + v.Len()

	default:
		if v.IsValid() {
			return sizeof(v)
		}
	}

	return -1
}

type coder struct {
	order  binary.ByteOrder
	buf    []byte
	offset int
}

type encoder coder

func (e *encoder) bool(x bool) {
	if x {
		e.buf[e.offset] = 1
	} else {
		e.buf[e.offset] = 0
	}
	e.offset++
}

func (e *encoder) uint8(x uint8) {
	e.buf[e.offset] = x
	e.offset++
}

func (e *encoder) uint16(x uint16) {
	e.order.PutUint16(e.buf[e.offset:e.offset+2], x)
	e.offset += 2
}

func (e *encoder) uint32(x uint32) {
	e.order.PutUint32(e.buf[e.offset:e.offset+4], x)
	e.offset += 4
}

func (e *encoder) uint64(x uint64) {
	e.order.PutUint64(e.buf[e.offset:e.offset+8], x)
	e.offset += 8
}

func (e *encoder) int8(x int8) { e.uint8(uint8(x)) }

func (e *encoder) int16(x int16) { e.uint16(uint16(x)) }

func (e *encoder) int32(x int32) { e.uint32(uint32(x)) }

func (e *encoder) int64(x int64) { e.uint64(uint64(x)) }

func (e *encoder) string(x string) {
	e.order.PutUint64(e.buf[e.offset:e.offset+8], uint64(len(x)))
	e.offset += 8
	e.offset += copy(e.buf[e.offset:], ([]byte)(x))
}

func (e *encoder) value(v reflect.Value) {
	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			e.value(v.Index(i))
		}

	case reflect.Struct:
		t := v.Type()
		l := v.NumField()
		for i := 0; i < l; i++ {
			// see comment for corresponding code in decoder.value()
			if v := v.Field(i); v.CanSet() || t.Field(i).Name != "_" {
				e.value(v)
			} else {
				e.skip(v)
			}
		}

	case reflect.Slice:
		l := v.Len()
		for i := 0; i < l; i++ {
			e.value(v.Index(i))
		}

	case reflect.Bool:
		e.bool(v.Bool())

	case reflect.Int8:
		e.int8(int8(v.Int()))
	case reflect.Int16:
		e.int16(int16(v.Int()))
	case reflect.Int32:
		e.int32(int32(v.Int()))
	case reflect.Int64:
		e.int64(v.Int())

	case reflect.Uint8:
		e.uint8(uint8(v.Uint()))
	case reflect.Uint16:
		e.uint16(uint16(v.Uint()))
	case reflect.Uint32:
		e.uint32(uint32(v.Uint()))
	case reflect.Uint64:
		e.uint64(v.Uint())

	case reflect.Float32:
		e.uint32(math.Float32bits(float32(v.Float())))
	case reflect.Float64:
		e.uint64(math.Float64bits(v.Float()))

	case reflect.Complex64:
		x := v.Complex()
		e.uint32(math.Float32bits(float32(real(x))))
		e.uint32(math.Float32bits(float32(imag(x))))
	case reflect.Complex128:
		x := v.Complex()
		e.uint64(math.Float64bits(real(x)))
		e.uint64(math.Float64bits(imag(x)))

	case reflect.String:
		e.string(v.String())

	case reflect.Ptr:
		if !v.IsNil() {
			e.value(v.Elem())
		}
	}
}

func (e *encoder) skip(v reflect.Value) {
	n := dataSize(v)
	clear(e.buf[e.offset : e.offset+n])
	e.offset += n
}

// sizeof returns the size >= 0 of variables for the given type or -1 if the type is not acceptable.
func sizeof(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Array:
		if s := sizeof(v.Elem()); s >= 0 {
			return s * v.Len()
		}

	case reflect.Struct:
		sum := 0
		for i, n := 0, v.NumField(); i < n; i++ {
			s := sizeof(v.Field(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Bool,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return int(v.Type().Size())

	case reflect.String:
		return 8 + v.Len()

	case reflect.Ptr:
		if v.IsNil() {
			return 0
		}
		return sizeof(v.Elem())
	}

	return -1
}
