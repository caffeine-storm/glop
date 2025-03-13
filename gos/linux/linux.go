package linux

// #cgo LDFLAGS: -lX11 -lGL
// #include "include/glop.h"
// #include "stdlib.h"
import "C"
import (
	"unsafe"

	"github.com/runningwild/glop/gin"
)

type SystemObject struct {
	horizon      int64
	windowHandle C.GlopWindowHandle // Handle to native per-window data
}

// Call after runtime.LockOSThread(), *NOT* in an init function
func (linux *SystemObject) Startup() {
	C.GlopInit()
}

func (linux *SystemObject) Run() {
	panic("Not implemented on linux")
}

func (linux *SystemObject) Quit() {
	panic("Not implemented on linux")
}

func (linux *SystemObject) CreateWindow(x, y, width, height int) {
	linux.windowHandle = C.GlopCreateWindow(unsafe.Pointer(&(([]byte("linux window"))[0])), C.int(x), C.int(y), C.int(width), C.int(height))
}

func (linux *SystemObject) SwapBuffers() {
	C.GlopSwapBuffers()
}

func (linux *SystemObject) Think() int64 {
	linux.horizon = int64(C.GlopThink())
	return linux.horizon
}

func GlopToGin(glopEvent *C.struct_GlopKeyEvent) gin.OsEvent {
	// TODO(tmckee): we should make this work; otherwise, we never get the
	// right mouse position.
	// wx,wy := linux.rawCursorToWindowCoords(int(glopEvent.cursor_x), int(glopEvent.cursor_y))
	keyId := gin.KeyId{
		Device: gin.DeviceId{
			// TODO(tmckee): we need to inspect the 'index' or 'device' to know
			// device type; right now, mouse events get labled as keyboard events
			// :(
			Type:  gin.DeviceTypeKeyboard,
			Index: gin.DeviceIndex(glopEvent.device),
		},
		Index: gin.KeyIndex(glopEvent.index),
	}
	return gin.OsEvent{
		KeyId:     keyId,
		Press_amt: float64(glopEvent.press_amt),
		Timestamp: int64(glopEvent.timestamp),
		// X : wx,
		// Y : wy,
	}
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO(tmckee): use a montonic clock for the timestamps
func (linux *SystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
	var firstEvent *C.struct_GlopKeyEvent
	var length C.size_t
	var horizon C.int64_t

	C.GlopGetInputEvents(&firstEvent, &length, &horizon)
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
			events[i] = GlopToGin(&eventChunk[j])
			i++
		}
	}

	// Typically, there'll be some non-full chunk of input that we still need to
	// process.
	eventChunk, _ = next64(eventIterator)
	for j := 0; j < int(length)%64; j++ {
		events[i] = GlopToGin(&eventChunk[j])
		i++
	}

	return events, linux.horizon
}

func (linux *SystemObject) HideCursor(hide bool) {
}

func (linux *SystemObject) rawCursorToWindowCoords(x, y int) (int, int) {
	wx, wy, _, wdy := linux.GetWindowDims()
	return x - wx, wy + wdy - y
}

func (linux *SystemObject) GetCursorPos() (int, int) {
	var x, y C.int
	C.GlopGetMousePosition(&x, &y)
	return linux.rawCursorToWindowCoords(int(x), int(y))
}

func (linux *SystemObject) GetWindowDims() (int, int, int, int) {
	var x, y, dx, dy C.int
	C.GlopGetWindowDims(&x, &y, &dx, &dy)
	return int(x), int(y), int(dx), int(dy)
}

func (linux *SystemObject) SetWindowSize(width, height int) {
	C.GlopSetWindowSize(C.int(width), C.int(height))
}

func (linux *SystemObject) EnableVSync(enable bool) {
	var _enable C.int
	if enable {
		_enable = 1
	}
	C.GlopEnableVSync(_enable)
}

func (linux *SystemObject) SetGlContext() {
	if linux.windowHandle.data == nil {
		// We haven't initialized a GL context yet; do nothing
		return
	}
	C.GlopSetGlContext(linux.windowHandle)
}
