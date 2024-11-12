package debug_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

type pixel struct {
	r, g, b, a byte
}

func TestDrawRect(t *testing.T) {
	width, height := 50, 50
	buffer := &bytes.Buffer{}

	rendertest.WithGlForTest(width, height, func(sys system.System, queue render.RenderQueueInterface) {
		queue.Queue(func(render.RenderQueueState) {
			debug.DrawRectNdc(-1, -1, 1, 1)
			sys.SwapBuffers()
			debug.ScreenShotRgba(width, height, buffer)
		})
		queue.Purge()
	})

	rgbaBytes := buffer.Bytes()
	if len(rgbaBytes) != width*height*4 {
		panic(fmt.Errorf("wrong number of bytes, expected %d got %d", width*height*4, len(rgbaBytes)))
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			// Each pixel is 4 bytes of {r, g, b, a} starting at x*4 + y*50*4
			// The whole screen should be {255, 0, 0, 255}
			idx := x*4 + y*50*4
			r, g, b, a := rgbaBytes[idx], rgbaBytes[idx+1], rgbaBytes[idx+2], rgbaBytes[idx+3]
			px := pixel{
				r: r,
				g: g,
				b: b,
				a: a,
			}
			expected := pixel{
				r: 255,
				g: 0,
				b: 0,
				a: 255,
			}
			if px != expected {
				fmt.Printf("bytes: %v\n", rgbaBytes)
				t.Fatalf("pixel mismatch at (%d, %d): %+v", x, y, px)
			}
		}
	}
}

/* DANGER WILL ROBINSON! XXX: this has been crashing windows when running on
* WSL. Run at your PERIL!
* func TestDrawManyRects(t *testing.T) {
* for i := 0; i < 500; i++ {
*   TestDrawRect(t)
* }
}
*/
