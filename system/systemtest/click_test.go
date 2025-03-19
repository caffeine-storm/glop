package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	"github.com/runningwild/glop/system/systemtest"
)

func WithTestWindow(dx, dy int, fn func(wdw systemtest.Window)) {
	rendertest.WithGlForTest(dx, dy, func(sys system.System, queue render.RenderQueueInterface) {
		wdw := systemtest.NewTestWindow(sys, queue)
		queue.Queue(func(st render.RenderQueueState) {
			fn(wdw)
		})
		queue.Purge()
	})
}

func TestE2EClickHelper(t *testing.T) {
	WithTestWindow(64, 64, func(wdw systemtest.Window) {
		driver := wdw.NewDriver()
		driver.Click(10, 42)

		driver.ProcessFrame()

		// Check that gin saw it.
		clickCount := gin.In().GetKeyById(gin.AnyMouseLButton).FramePressCount()

		if clickCount <= 0 {
			t.Fatalf("didn't see a click!")
		}
	})
}
