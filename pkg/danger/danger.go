package danger

import (
	"reflect"
	"unsafe"
)

// // These method were implemented as a low-allocation, high-performance way to convert bytes<->string
// // When running on go 1.178.8, the following error is reported
// //      fatal error: checkptr: converted pointer straddles multiple allocations
// // For now, falling back to the classic conversion approach, but would be good to revisit

// func BytesToString(bytes []byte) (s string) {
// 	return string(bytes)
// }

// func StringToBytes(s string) (b []byte) {
// 	return []byte(s)
// }

// Original implemnentations below:

// BytesToString turns a []byte into a string with 0 MemAllocs and 0 MemBytes.
// This is an unsafe operation and may lead to problems if the bytes passed in
// are changed while the string is used. No checking whether bytes are valid
// UTF-8 data is performed.
func BytesToString(bytes []byte) (s string) {
	if len(bytes) == 0 {
		return s
	}
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sh.Data = uintptr(unsafe.Pointer(&bytes[0]))
	sh.Len = len(bytes)
	return s
}

// StringToBytes turns a string into a []byte with 0 MemAllocs and 0 MemBytes.
// This is an unsafe operation and will lead to problems if the underlying bytes
// are changed.
func StringToBytes(s string) (b []byte) {
	if len(s) == 0 {
		return b
	}
	const max = 0x7fff0000 // 2147418112
	if len(s) > max {
		panic("string too large")
	}
	bytes := (*[max]byte)(
		unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data),
	)
	return bytes[:len(s):len(s)]
}
