package rendertest_test

import (
	"fmt"
	"image"
	"testing"

	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDrawRect(t *testing.T) {
	Convey("rendertest.DrawRect should work", t, DrawRectSpec)
}

func DrawRectSpec() {
	width, height := 50, 50
	var result *image.RGBA

	rendertest.WithGlForTest(width, height, func(sys system.System, queue render.RenderQueueInterface) {
		queue.Queue(func(render.RenderQueueState) {
			rendertest.BlankAndDrawRectNdc(-1, -1, 1, 1)
			result = debug.ScreenShotRgba(width, height)
		})
		queue.Purge()

		if len(result.Pix) != width*height*4 {
			panic(fmt.Errorf("wrong number of bytes, expected %d got %d", width*height*4, len(result.Pix)))
		}

		Convey("Should see red pixels", func() {
			So(queue, rendertest.ShouldLookLikeFile, "red")
		})
	})
}

/* DANGER WILL ROBINSON! XXX: this has been crashing windows when running on
* WSL if we don't cache XWindow instances/glxContexts. Run at your PERIL!
* func TestDrawManyRects(t *testing.T) {
* for i := 0; i < 500; i++ {
*   TestDrawRect(t)
* }
}
*/
