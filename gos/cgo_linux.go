package gos

// #cgo LDFLAGS: -Llinux/lib -lglop -lX11 -lGL
// #include "linux/include/glop.h"
import "C"

import (
  "github.com/runningwild/glop/system"
  "github.com/runningwild/glop/gin"
  "unsafe"
)

type linuxSystemObject struct {
  horizon int64
}

var (
  linux_system_object linuxSystemObject
)

// Call after runtime.LockOSThread(), *NOT* in an init function
func (linux *linuxSystemObject) Startup() {
  C.GlopInit()
}

func GetSystemInterface() system.Os {
  return &linux_system_object
}

func (linux *linuxSystemObject) Run() {
  panic("Not implemented on linux")
}

func (linux *linuxSystemObject) Quit() {
  panic("Not implemented on linux")
}

func (linux *linuxSystemObject) CreateWindow(x,y,width,height int) {
  C.GlopCreateWindow(unsafe.Pointer(&(([]byte("linux window"))[0])), C.int(x), C.int(y), C.int(width), C.int(height))
}

func (linux *linuxSystemObject) SwapBuffers() {
  C.GlopSwapBuffers()
}

func (linux *linuxSystemObject) Think() {
  C.GlopThink()
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO: Adjust timestamp on events so that the oldest timestamp is newer than the
//       newest timestemp from the events from the previous call to GetInputEvents
//       Actually that should be in system
func (linux *linuxSystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
  var first_event *C.GlopKeyEvent
  cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
  var length C.int
  var horizon C.longlong
  C.GlopGetInputEvents(cp, unsafe.Pointer(&length), unsafe.Pointer(&horizon))
  linux.horizon = int64(horizon)
  c_events := (*[1000]C.GlopKeyEvent)(unsafe.Pointer(first_event))[:length]
  events := make([]gin.OsEvent, length)
  for i := range c_events {
    // TODO(tmckee): we should make this work; otherwise, we never get the
    // right mouse position.
    // wx,wy := linux.rawCursorToWindowCoords(int(c_events[i].cursor_x), int(c_events[i].cursor_y))
    keyId := gin.KeyId{
      Device: gin.DeviceId{
        // TODO(tmckee): we need to inspect the 'index' or 'device' to know
        // device type; right now, mouse events get labled as keyboard events
        // :(
        Type: gin.DeviceTypeKeyboard,
        Index: gin.DeviceIndex(c_events[i].device),
      },
      Index: gin.KeyIndex(c_events[i].index),
    }
    events[i] = gin.OsEvent{
      KeyId     : keyId,
      Press_amt : float64(c_events[i].press_amt),
      Timestamp : int64(c_events[i].timestamp),
      // X : wx,
      // Y : wy,
    }
  }
  return events, linux.horizon
}

func (linux *linuxSystemObject) HideCursor(hide bool) {
}

func (linux *linuxSystemObject) rawCursorToWindowCoords(x,y int) (int,int) {
  wx,wy,_,wdy := linux.GetWindowDims()
  return x - wx, wy + wdy - y
}

func (linux *linuxSystemObject) GetCursorPos() (int,int) {
  var x,y C.int
  C.GlopGetMousePosition(&x, &y)
  return linux.rawCursorToWindowCoords(int(x), int(y))
}

func (linux *linuxSystemObject) GetWindowDims() (int,int,int,int) {
  var x,y,dx,dy C.int
  C.GlopGetWindowDims(&x, &y, &dx, &dy)
  return int(x), int(y), int(dx), int(dy)
}

func (linux *linuxSystemObject) EnableVSync(enable bool) {
  var _enable C.int
  if enable {
    _enable = 1
  }
  C.GlopEnableVSync(_enable)
}
