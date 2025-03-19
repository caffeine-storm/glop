package systemtest_test

import (
	"fmt"
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
		expectedX, expectedY := 10, 42
		driver.Click(expectedX, expectedY)

		driver.ProcessFrame()

		// Check that gin saw it.
		lbuttonKey := gin.In().GetKeyById(gin.AnyMouseLButton)

		if lbuttonKey.FramePressCount() <= 0 {
			t.Fatalf("didn't see a click!")
		}

		actualX, actualY := lbuttonKey.Cursor().Point()

		if actualX != expectedX || actualY != expectedY {
			t.Fatalf("click co-ordinates didn't match! expected: %s actual %s",
				fmt.Sprintf("(%d, %d)", expectedX, expectedY),
				fmt.Sprintf("(%d, %d)", actualX, actualY))
		}
	})
}
