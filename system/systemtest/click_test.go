package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	"github.com/runningwild/glop/system/systemtest"
)

func WithTestWindow(dx, dy int, fn func(window systemtest.Window)) {
	rendertest.WithGlForTest(dx, dy, func(sys system.System, queue render.RenderQueueInterface) {
		window := systemtest.NewTestWindow(sys, queue)
		queue.Queue(func(st render.RenderQueueState) {
			fn(window)
		})
		queue.Purge()
	})
}

func WithTestWindowDriver(dx, dy int, fn func(driver systemtest.Driver)) {
	WithTestWindow(dx, dy, func(window systemtest.Window) {
		fn(window.NewDriver())
	})
}

func TestE2EClickHelper(t *testing.T) {
	WithTestWindowDriver(64, 64, func(driver systemtest.Driver) {
		driver.Click(10, 42)

		driver.ProcessFrame()

		// Check that gin saw it.
		clickCount := gin.In().GetKeyById(gin.AnyMouseLButton).FramePressCount()

		if clickCount <= 0 {
			t.Fatalf("didn't see a click!")
		}
	})
}
