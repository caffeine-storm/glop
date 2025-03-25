package system_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system"
)

type stubSystem struct{}

func (*stubSystem) Startup() {
}

func (*stubSystem) Think() int64 {
	return 7
}

func (*stubSystem) CreateWindow(x, y, width, height int) system.NativeWindowHandle {
	return "stub handle"
}

func (*stubSystem) HideCursor(bool) {
}

func (*stubSystem) GetWindowDims() (x, y, dx, dy int) {
	return
}

func (*stubSystem) SetWindowSize(width, height int) {}

func (*stubSystem) SwapBuffers() {}
func (*stubSystem) GetInputEvents() []gin.EventGroup {
	return nil
}

func (*stubSystem) EnableVSync(bool) {}

var _ system.System = (*stubSystem)(nil)

func GivenASystem() system.System {
	return &stubSystem{}
}

func TestSystem(t *testing.T) {
	t.Run("CreateWindow", func(t *testing.T) {
		t.Run("returns a native id", func(t *testing.T) {
			sys := GivenASystem()
			windowHandle := sys.CreateWindow(0, 0, 42, 42)
			if windowHandle == nil {
				t.Fatalf("window handles should not be nil")
			}
		})
	})
}
