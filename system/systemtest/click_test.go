package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

func WithTestWindow(dx, dy int, fn func(wdw any, driver systemtest.Driver)) {
	rendertest.WithGlForTest(dx, dy, func(sys system.System, queue render.RenderQueueInterface) {
		wdw := systemtest.NewTestWindow(sys, queue)
		driver := systemtest.NewDriver(wdw)
		queue.Queue(func(st render.RenderQueueState) {
			fn(wdw, driver)
		})
		queue.Purge()
	})
}

func TestE2EClickHelper(t *testing.T) {
	WithTestWindow(64, 64, func(wdw systemtest.Window, driver systemtest.Driver) {
		driver.Click(10, 42)

		driver.ProcessFrame()

		// Check that gin saw it.
		clickCount := gin.In().GetKeyById(gin.AnyMouseLButton).FramePressCount()

		if clickCount <= 0 {
			t.Fatalf("didn't see a click!")
		}
	})
}
