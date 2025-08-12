package gloptest

import (
	"fmt"
	"reflect"
	"runtime"
)

func FileLineForClosure(fn any) (string, int) {
	reflected := reflect.ValueOf(fn)
	if reflected.Kind() != reflect.Func {
		panic(fmt.Errorf("can't get file+line info for non-function; given type: %T", fn))
	}
	up := uintptr(reflected.UnsafePointer())
	return runtime.FuncForPC(up).FileLine(up)
}
