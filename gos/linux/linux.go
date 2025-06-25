package linux

// #cgo LDFLAGS: -lX11 -lGL
// #include "include/glop.h"
// #include "stdlib.h"
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/system"
)

type SystemObject struct {
	horizon      int64
	windowHandle C.GlopWindowHandle // Handle to native per-window data
}

func (linux *SystemObject) Startup() int64 {
	return int64(C.GlopInit())
}

func (linux *SystemObject) Run() {
	panic("Not implemented on linux")
}

func (linux *SystemObject) Quit() {
	panic("Not implemented on linux")
}

// Call after runtime.LockOSThread(), *NOT* in an init function
func (linux *SystemObject) CreateWindow(x, y, width, height int) system.NativeWindowHandle {
	linux.windowHandle = C.GlopCreateWindowHandle(C.CString("linux window"), C.int(x), C.int(y), C.int(width), C.int(height))
	return fmt.Sprintf("%d", C.GetNativeHandle(linux.windowHandle))
}

func (linux *SystemObject) SwapBuffers() {
	C.GlopSwapBuffers(linux.windowHandle)
}

func (linux *SystemObject) Think() int64 {
	linux.horizon = int64(C.GlopThink(linux.windowHandle))
	return linux.horizon
}

func nativeDeviceToGinDevice(n C.short) gin.DeviceType {
	switch n {
	case C.glopDeviceKeyboard:
		return gin.DeviceTypeKeyboard
	case C.glopDeviceMouse:
		return gin.DeviceTypeMouse
	case C.glopDeviceDerived:
		return gin.DeviceTypeDerived
		// gin.DeviceTypeController is not supported right now
	}

	panic(fmt.Errorf("nativeDeviceToGinDevice: got invalid value %d", n))
}

type NativeKeyEvent C.struct_GlopKeyEvent

func NewNativeMouseEvent(x, y int) *NativeKeyEvent {
	return &NativeKeyEvent{
		index:       0,
		device_type: C.glopDeviceMouse,
		cursor_x:    C.int(x),
		cursor_y:    C.int(y),
	}
}

type RawCursorToWindowCoordser interface {
	RawCursorToWindowCoords(x, y int) (int, int)
}

func NativeToGin(linux RawCursorToWindowCoordser, nativeEvent *NativeKeyEvent) gin.OsEvent {
	wx, wy := linux.RawCursorToWindowCoords(int(nativeEvent.cursor_x), int(nativeEvent.cursor_y))
	keyId := gin.KeyId{
		Device: gin.DeviceId{
			Type: nativeDeviceToGinDevice(nativeEvent.device_type),
			// TODO(#28): shouldn't we be indexing devices?
			Index: 0, // gin.DeviceIndex(nativeEvent.device_index),
		},
		Index: gin.KeyIndex(nativeEvent.index),
	}
	ret := gin.OsEvent{
		KeyId:     keyId,
		Press_amt: float64(nativeEvent.press_amt),
		Timestamp: int64(nativeEvent.timestamp),
		X:         wx,
		Y:         wy,
	}

	glog.TraceLogger().Trace("native to gin", "native", *nativeEvent, "ret", ret)

	return ret
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO(tmckee): use a montonic clock for the timestamps
func (linux *SystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
	var firstEvent *C.struct_GlopKeyEvent
	var length C.size_t
	var horizon C.int64_t

	if linux.windowHandle.data == nil {
		panic("can't call GetInputEvents before opening the window!")
	}

	C.GlopGetInputEvents(linux.windowHandle, &firstEvent, &length, &horizon)
	defer C.free(unsafe.Pointer(firstEvent))
	linux.horizon = int64(horizon)

	// Given a pointer to a C array, returns the same pointer co-erced to a 64
	// element array and a pointer to one-past that array. Useful for iterating
	// elements of a C array 64 elements at a time.
	next64 := func(itr *C.struct_GlopKeyEvent) (*[64]C.struct_GlopKeyEvent, *C.struct_GlopKeyEvent) {
		result := (*[64]C.struct_GlopKeyEvent)(unsafe.Pointer(itr))
		bounds := uintptr(64 * C.sizeof_struct_GlopKeyEvent)
		itr = (*C.struct_GlopKeyEvent)(unsafe.Pointer((uintptr(unsafe.Pointer(itr)) + bounds)))
		return result, itr
	}

	events := make([]gin.OsEvent, length)
	var eventChunk *[64]C.struct_GlopKeyEvent
	i := 0
	eventIterator := firstEvent
	for chunk := 0; chunk < int(length)/64; chunk++ {
		eventChunk, eventIterator = next64(eventIterator)
		for j := 0; j < 64; j++ {
			events[i] = NativeToGin(linux, (*NativeKeyEvent)(&eventChunk[j]))
			i++
		}
	}

	// Typically, there'll be some non-full chunk of input that we still need to
	// process.
	eventChunk, _ = next64(eventIterator)
	for j := 0; j < int(length)%64; j++ {
		events[i] = NativeToGin(linux, (*NativeKeyEvent)(&eventChunk[j]))
		i++
	}

	return events, linux.horizon
}

func (linux *SystemObject) HideCursor(hide bool) {
}

func (linux *SystemObject) RawCursorToWindowCoords(x, y int) (int, int) {
	return x, y
}

func (linux *SystemObject) GetWindowDims() (int, int, int, int) {
	var x, y, dx, dy C.int
	C.GlopGetWindowDims(linux.windowHandle, &x, &y, &dx, &dy)
	return int(x), int(y), int(dx), int(dy)
}

func (linux *SystemObject) SetWindowSize(width, height int) {
	C.GlopSetWindowSize(linux.windowHandle, C.int(width), C.int(height))
}

func (linux *SystemObject) EnableVSync(enable bool) {
	var _enable C.int
	if enable {
		_enable = 1
	}
	C.GlopEnableVSync(_enable)
}

// TODO(tmckee)(clean): this isn't used; remove it!
func (linux *SystemObject) SetGlContext() {
	if linux.windowHandle.data == nil {
		// We haven't initialized a GL context yet; do nothing
		return
	}
	C.GlopSetGlContext(linux.windowHandle)
}

func New() *SystemObject {
	ret := &SystemObject{}
	ret.Startup()
	return ret
}
