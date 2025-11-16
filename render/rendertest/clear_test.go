package rendertest_test

import (
	"image"
	"testing"

	"github.com/caffeine-storm/glop/debug"
	"github.com/caffeine-storm/glop/render"
	"github.com/caffeine-storm/glop/render/rendertest"
	"github.com/caffeine-storm/glop/render/rendertest/testbuilder"
)

func TestClearScreen(t *testing.T) {
	testbuilder.Run(func(queue render.RenderQueueInterface) {
		var imgResult *image.NRGBA
		queue.Queue(func(st render.RenderQueueState) {
			// draw some stuff
			render.WithColour(0, 1, 0, 1, func() {
				// green!
				rendertest.DrawRectNdc(-0.75, -0.75, 0.75, 0.75)

				// clear the screen
				rendertest.ClearScreen()

				imgResult = debug.ScreenShotNrgba(64, 64)
			})
		})
		queue.Purge()

		// expect all black
		badPixels := []any{}
		for y := imgResult.Bounds().Min.Y; y < imgResult.Bounds().Max.Y; y++ {
			for x := imgResult.Bounds().Min.X; x < imgResult.Bounds().Max.X; x++ {
				r, g, b, a := imgResult.At(x, y).RGBA()
				if r != 0 || g != 0 || b != 0 || a != 0xffff {
					t.Fail()
					badPixels = append(badPixels, []any{[]any{"xy", x, y}, []any{"rgba", r, g, b, a}})
				}
			}
		}

		if len(badPixels) > 0 {
			t.Log("expected black image", "bad pixels", badPixels)
		}
	})
}
