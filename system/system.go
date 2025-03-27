package system

import (
	"github.com/runningwild/glop/gin"
)

type NativeWindowHandle interface{}

type System interface {
	// Call after runtime.LockOSThread(), *NOT* in an init function.
	Startup()

	// Call System.Think() every frame. Returns the 'horizon'.
	Think() int64

	CreateWindow(x, y, width, height int) NativeWindowHandle
	// TODO: implement this:
	// DestroyWindow(NativeWindowHandle)

	// Hides/Unhides the cursor. A hidden cursor is invisible and its position is
	// locked. It should still generate mouse move events.
	HideCursor(bool)

	GetWindowDims() (x, y, dx, dy int)
	SetWindowSize(width, height int)

	SwapBuffers()
	GetInputEvents() []gin.EventGroup

	EnableVSync(bool)

	// These probably shouldn't be here, probably always want to do the Think()
	// approach
	//  Run()
	//  Quit()

	// --- helpful features in system objects that aren't really native features.

	// Attach a Mouse listener to the underlying input delegate.
	AddMouseListener(func(gin.MouseEvent))
}

// This is the interface implemented by any operating system that glop
// supports. The gos package on that OS should export a function called
// NewSystemInterface() which takes no parameters and returns an object that
// implements the system.Os interface.
type Os interface {
	// This is properly called after runtime.LockOSThread(), not in an init
	// function
	Startup()

	// Think() is called on a regular basis and always from main thread. Returns
	// the an event horizon; a timestamp that can be compared with
	// GetInputEvent's timestamps.
	Think() int64

	// Create a window with the appropriate dimensions and bind an OpenGl contxt
	// to it.  Currently glop only supports a single window, but this function
	// could be called more than once since a window could be destroyed so it can
	// be recreated at different dimensions or in full sreen mode.
	CreateWindow(x, y, width, height int) NativeWindowHandle

	// TODO: implement this:
	// DestroyWindow(NativeWindowHandle)

	// Hides/Unhides the cursor. A hidden cursor is invisible and its position is
	// locked. It should still generate mouse move events.
	HideCursor(bool)

	GetWindowDims() (x, y, dx, dy int)
	SetWindowSize(width, height int)

	// Swap the OpenGl buffers on this window
	SwapBuffers()

	// Returns all of the events in the order that they happened since the last
	// call to this function. The events do not have to be in order according to
	// KeyEvent.Timestamp, but they will be sorted according to this value. The
	// timestamp returned is the event horizon, no future events will have a
	// timestamp less than or equal to it.
	GetInputEvents() ([]gin.OsEvent, int64)

	EnableVSync(bool)

	// These probably shouldn't be here, probably always want to do the Think()
	// approach
	//  Run()
	//  Quit()
}

type sysObj struct {
	os       Os
	input    *gin.Input
	events   []gin.EventGroup
	start_ms int64
}

func Make(os Os, input *gin.Input) System {
	return &sysObj{
		os:    os,
		input: input,
	}
}

func (sys *sysObj) Startup() {
	sys.os.Startup()
	sys.start_ms = 14 // TODO(tmckee): get an initial timestamp from native call to GlopInit()
}

func (sys *sysObj) Think() int64 {
	sys.os.Think()
	events, horizon := sys.os.GetInputEvents()
	for i := range events {
		events[i].Timestamp -= sys.start_ms
	}
	sys.events = sys.input.Think(horizon-sys.start_ms, events)
	return horizon
}

func (sys *sysObj) CreateWindow(x, y, width, height int) NativeWindowHandle {
	return sys.os.CreateWindow(x, y, width, height)
}

func (sys *sysObj) HideCursor(hide bool) {
	sys.os.HideCursor(hide)
}

func (sys *sysObj) GetWindowDims() (int, int, int, int) {
	return sys.os.GetWindowDims()
}

func (sys *sysObj) SetWindowSize(width, height int) {
	sys.os.SetWindowSize(width, height)
}

func (sys *sysObj) SwapBuffers() {
	sys.os.SwapBuffers()
}

func (sys *sysObj) GetInputEvents() []gin.EventGroup {
	return sys.events
}

func (sys *sysObj) AddMouseListener(fn func(gin.MouseEvent)) {
	sys.input.AddMouseListener(fn)
}

func (sys *sysObj) EnableVSync(enable bool) {
	sys.os.EnableVSync(enable)
}
