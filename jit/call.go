package jit

import (
	"reflect"
	"unsafe"
)

func call()

func FuncOf(a Alloc) func() {
	entry := call

	fn := &struct {
		entry unsafe.Pointer
		code  unsafe.Pointer
	}{**(**unsafe.Pointer)(unsafe.Pointer(&entry)), a.Addr()}

	return *(*func())(unsafe.Pointer(&fn))
}

func SetFunc(dst interface{}, a Alloc) {
	if v := reflect.ValueOf(dst); v.Type().Kind() != reflect.Ptr ||
		v.Elem().Type().Kind() != reflect.Func {
		panic("dst must be a pointer fo func")
	}

	ival := *(*struct {
		typ unsafe.Pointer
		val unsafe.Pointer
	})(unsafe.Pointer(&dst))

	*(*func())(ival.val) = FuncOf(a)
}
