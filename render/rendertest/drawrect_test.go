package rendertest_test

import (
	"fmt"
	"image"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDrawRect(t *testing.T) {
	Convey("rendertest.DrawRect should work", t, DrawRectSpec)
}

var width, height = 64, 64

func RunTest(withTex2d bool) {
	testbuilder.New().WithSize(width, height).WithQueue().Run(func(queue render.RenderQueueInterface) {
		var result *image.RGBA
		queue.Queue(func(render.RenderQueueState) {
			if withTex2d {
				gl.Enable(gl.TEXTURE_2D)
			} else {
				gl.Disable(gl.TEXTURE_2D)
			}
			rendertest.BlankAndDrawRectNdc(-1, -1, 1, 1)
			result = debug.ScreenShotRgba(width, height)
		})
		queue.Purge()

		if len(result.Pix) != width*height*4 {
			panic(fmt.Errorf("wrong number of bytes, expected %d got %d", width*height*4, len(result.Pix)))
		}

		Convey("Should see red pixels", func() {
			So(queue, rendertest.ShouldLookLikeFile, "red64")
		})
	})
}

func DrawRectSpec() {
	Convey("with gl.TEXTURE_2D enabled", func() {
		RunTest(true)
	})

	Convey("with gl.TEXTURE_2D disabled", func() {
		RunTest(false)
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
