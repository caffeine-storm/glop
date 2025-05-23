package rendertest_test

import (
	"image"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

func TestClearScreen(t *testing.T) {
	rendertest.DeprecatedWithGlForTest(50, 50, func(_ system.System, queue render.RenderQueueInterface) {
		var imgResult *image.RGBA
		queue.Queue(func(st render.RenderQueueState) {
			// draw some stuff
			gl.Color4f(0, 1, 0, 1) // green!
			rendertest.DrawRectNdc(-0.75, -0.75, 0.75, 0.75)

			// clear the screen
			rendertest.ClearScreen()

			imgResult = debug.ScreenShotRgba(50, 50)
		})
		queue.Purge()

		// expect all black
		badPixels := []int{}
		for y := 0; y < 50; y++ {
			for x := 0; x < 50; x++ {
				r, g, b, a := imgResult.At(x, y).RGBA()
				if r != 0 || g != 0 || b != 0 || a != 0xffff {
					t.Fail()
					badPixels = append(badPixels, x, y, int(r), int(g), int(b), int(a))
				}
			}
		}

		if len(badPixels) > 0 {
			t.Log("expected black image", "bad pixels", badPixels)
		}
	})
}
