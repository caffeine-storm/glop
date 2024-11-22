package linux

// #cgo LDFLAGS: -lX11 -lGL
// #include "include/glop.h"
// #include "stdlib.h"
import "C"
import (
	"runtime"
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
	return int64(C.GlopThink())
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO(tmckee): use a montonic clock for the timestamps
func (linux *SystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
	var first_event *C.struct_GlopKeyEvent
	var length C.size_t
	var horizon C.int64_t

	pinner := runtime.Pinner{}
	defer pinner.Unpin()
	pinner.Pin(&first_event)
	pinner.Pin(&length)
	pinner.Pin(&horizon)

	C.GlopGetInputEvents(&first_event, &length, &horizon)

	defer C.free(unsafe.Pointer(first_event))
	linux.horizon = int64(horizon)

	events := make([]gin.OsEvent, length)
	i := 0
	event_iterator := first_event
	for chunk := 0; chunk < int(length)/64; chunk++ {
		c_events := (*[64]C.struct_GlopKeyEvent)(unsafe.Pointer(event_iterator))[:]
		for _, c_event := range c_events {
			// TODO(tmckee): we should make this work; otherwise, we never get the
			// right mouse position.
			// wx,wy := linux.rawCursorToWindowCoords(int(c_events[i].cursor_x), int(c_events[i].cursor_y))
			keyId := gin.KeyId{
				Device: gin.DeviceId{
					// TODO(tmckee): we need to inspect the 'index' or 'device' to know
					// device type; right now, mouse events get labled as keyboard events
					// :(
					Type:  gin.DeviceTypeKeyboard,
					Index: gin.DeviceIndex(c_event.device),
				},
				Index: gin.KeyIndex(c_event.index),
			}
			events[i] = gin.OsEvent{
				KeyId:     keyId,
				Press_amt: float64(c_event.press_amt),
				Timestamp: int64(c_event.timestamp),
				// X : wx,
				// Y : wy,
			}
			i++
		}
		event_iterator = (*C.struct_GlopKeyEvent)(unsafe.Pointer(uintptr(unsafe.Pointer(event_iterator)) + 64*C.sizeof_struct_GlopKeyEvent))
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
