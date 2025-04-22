// package tls uses native code to expose Thread Local Storage primitives.
package tls

// #include "stdint.h"
// #include "stdlib.h"
// #include "threads.h"
//
// int makeflag(tss_t* out) {
//   return tss_create(out, NULL);
// }
//
// void setflag(tss_t* flag, uintptr_t val) {
//   tss_set(*flag, (void *)val);
// }
//
// uintptr_t getflag(tss_t* flag) {
//   void* val = tss_get(*flag);
//   return (uintptr_t)val;
// }
import "C"
import "fmt"

var flag C.tss_t

func init() {
	err := C.makeflag(&flag)
	if err != 0 {
		panic(fmt.Errorf("couldn't tss_create: %d", err))
	}
	ClearSentinel()
}

func SetSentinel() {
	C.setflag(&flag, 1)
}

func ClearSentinel() {
	C.setflag(&flag, 0)
}

func IsSentinelSet() bool {
	val := C.getflag(&flag)
	return val == 1
}
