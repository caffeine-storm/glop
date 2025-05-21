package linux_test

import (
	"runtime"
	"testing"

	"github.com/go-gl-legacy/gl"
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
	t.Run("can create window", func(t *testing.T) {
		toRunUnderGLContext := make(chan func())
		success := make(chan bool)
		ack := make(chan bool)
		go func() {
			runtime.LockOSThread()
			sysObj := linux.New()

			hdl := sysObj.CreateWindow(0, 0, 64, 64)

			success <- hdl != nil

			for fn := range toRunUnderGLContext {
				fn()
				ack <- true
			}
		}()

		if !<-success {
			t.Fatalf("sysObj.CreateWindow failed!")
		}

		t.Run("GL context has the right version", func(t *testing.T) {
			toRunUnderGLContext <- func() {
				major := gl.GetInteger(gl.MAJOR_VERSION)
				minor := gl.GetInteger(gl.MINOR_VERSION)
				t.Logf("glversion: %d.%d", major, minor)
				if major != 4 || minor < 5 {
					t.Logf("bad glversion: %d.%d", major, minor)
					t.Fail()
				}
			}
			<-ack
			toRunUnderGLContext <- func() {
				profile := gl.GetInteger(gl.CONTEXT_PROFILE_MASK)
				if profile != gl.CONTEXT_CORE_PROFILE_BIT {
					t.Logf("bad profile mask: %d", profile)
					t.Fail()
				}
			}
			<-ack
			close(toRunUnderGLContext)
		})
	})
}
