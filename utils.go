package storage

import (
	"reflect"
	"unsafe"
)

func bytes2String(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}

	sh := &reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&bs[0])),
		Len:  len(bs),
	}
	return *(*string)(unsafe.Pointer(sh))
}
