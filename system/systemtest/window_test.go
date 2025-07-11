package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system/systemtest"
)

func TestTestWindow(t *testing.T) {
	systemtest.WithTestWindow(64, 64, func(window systemtest.Window) {
		t.Run("exposes a queue", func(t *testing.T) {
			var _ render.RenderQueueInterface = window.GetQueue()
		})

		t.Run("exposes window dimensions", func(t *testing.T) {
			var _ gui.Dims = window.GetDims()
		})

		t.Run("runs testcase off of the render thread", func(t *testing.T) {
			render.MustNotBeOnRenderThread()
		})
	})
}
