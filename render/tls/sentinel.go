// package tls uses native code to expose Thread Local Storage primitives.
package tls

// #include "stdlib.h"
// #include "threads.h"
//
// int makeflag(tss_t* out) {
//   return tss_create(out, NULL);
// }
//
// void setflag(tss_t* flag, int val) {
//   int* valptr = malloc(sizeof(int));
//   *valptr = val;
//   tss_set(*flag, valptr);
// }
//
// int getflag(tss_t* flag) {
//   int* val = tss_get(*flag);
//   return *val;
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
