package pegomock

import (
	"reflect"
)

func EqBool(value bool) bool {
	RegisterMatcher(&EqMatcher{Value: value})
	return false
}

func AnyBool() bool {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((bool)(false))))
	return false
}

func AnyBoolSlice() []bool {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((bool)(false)))))
	return nil
}

func EqInt(value int) int {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyInt() int {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((int)(0))))
	return 0
}

func AnyIntSlice() []int {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((int)(0)))))
	return nil
}

func EqInt8(value int8) int8 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyInt8() int8 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((int8)(0))))
	return 0
}

func AnyInt8Slice() []int8 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((int8)(0)))))
	return nil
}

func EqInt16(value int16) int16 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyInt16() int16 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((int16)(0))))
	return 0
}

func AnyInt16Slice() []int16 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((int16)(0)))))
	return nil
}

func EqInt32(value int32) int32 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyInt32() int32 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((int32)(0))))
	return 0
}

func AnyInt32Slice() []int32 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((int32)(0)))))
	return nil
}

func EqInt64(value int64) int64 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyInt64() int64 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((int64)(0))))
	return 0
}

func AnyInt64Slice() []int64 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((int64)(0)))))
	return nil
}

func EqUint(value uint) uint {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyUint() uint {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((uint)(0))))
	return 0
}

func AnyUintSlice() []uint {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((uint)(0)))))
	return nil
}

func EqUint8(value uint8) uint8 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyUint8() uint8 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((uint8)(0))))
	return 0
}

func AnyUint8Slice() []uint8 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((uint8)(0)))))
	return nil
}

func EqUint16(value uint16) uint16 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyUint16() uint16 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((uint16)(0))))
	return 0
}

func AnyUint16Slice() []uint16 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((uint16)(0)))))
	return nil
}

func EqUint32(value uint32) uint32 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyUint32() uint32 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((uint32)(0))))
	return 0
}

func AnyUint32Slice() []uint32 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((uint32)(0)))))
	return nil
}

func EqUint64(value uint64) uint64 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyUint64() uint64 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((uint64)(0))))
	return 0
}

func AnyUint64Slice() []uint64 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((uint64)(0)))))
	return nil
}

func EqUintptr(value uintptr) uintptr {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyUintptr() uintptr {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((uintptr)(0))))
	return 0
}

func AnyUintptrSlice() []uintptr {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((uintptr)(0)))))
	return nil
}

func EqFloat32(value float32) float32 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyFloat32() float32 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((float32)(0))))
	return 0
}

func AnyFloat32Slice() []float32 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((float32)(0)))))
	return nil
}

func EqFloat64(value float64) float64 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyFloat64() float64 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((float64)(0))))
	return 0
}

func AnyFloat64Slice() []float64 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((float64)(0)))))
	return nil
}

func EqComplex64(value complex64) complex64 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyComplex64() complex64 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((complex64)(0))))
	return 0
}

func AnyComplex64Slice() []complex64 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((complex64)(0)))))
	return nil
}

func EqComplex128(value complex128) complex128 {
	RegisterMatcher(&EqMatcher{Value: value})
	return 0
}

func AnyComplex128() complex128 {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((complex128)(0))))
	return 0
}

func AnyComplex128Slice() []complex128 {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((complex128)(0)))))
	return nil
}

func EqString(value string) string {
	RegisterMatcher(&EqMatcher{Value: value})
	return ""
}

func AnyString() string {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((string)(""))))
	return ""
}

func AnyStringSlice() []string {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((string)("")))))
	return nil
}

func EqInterface(value interface{}) interface{} {
	RegisterMatcher(&EqMatcher{Value: value})
	return nil
}

func AnyInterface() interface{} {
	RegisterMatcher(NewAnyMatcher(reflect.TypeOf((*(interface{}))(nil)).Elem()))
	return nil
}

func AnyInterfaceSlice() []interface{} {
	RegisterMatcher(NewAnyMatcher(reflect.SliceOf(reflect.TypeOf((*(interface{}))(nil)).Elem())))
	return nil
}
	