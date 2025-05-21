package linux_test

import (
	"runtime"
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gos/linux"
)

type stubCoordser struct {
	x, y int
}

func (m *stubCoordser) RawCursorToWindowCoords(x, y int) (int, int) {
	return m.x, m.y
}

func StubCoordser(x, y int) linux.RawCursorToWindowCoordser {
	return &stubCoordser{
		x: x,
		y: y,
	}
}

func GivenAClickAt(x, y int) *linux.NativeKeyEvent {
	return linux.NewNativeMouseEvent(x, y)
}

func TestNativeToGin(t *testing.T) {
	evt := GivenAClickAt(42, 1812)

	coordser := StubCoordser(1, 2)
	ginEvent := linux.NativeToGin(coordser, evt)

	if ginEvent.X != 1 || ginEvent.Y != 2 {
		t.Fatalf("Expected the 'raw' coords to be (1, 2), got (%d, %d)", ginEvent.X, ginEvent.Y)
	}

	if ginEvent.KeyId.Device.Type != gin.DeviceTypeMouse {
		t.Fatalf("a click event should be attributed to a mouse")
	}
}

func TestGlopCreateWindowHandle(t *testing.T) {
	success := make(chan bool)
	go func() {
		runtime.LockOSThread()
		sysObj := linux.New()

		hdl := sysObj.CreateWindow(0, 0, 64, 64)

		success <- hdl != nil
	}()

	if !<-success {
		t.Fatalf("sysObj.CreateWindow failed!")
	}
}
