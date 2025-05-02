package rendertest_test

import (
	"image"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDrawTexturedQuad(t *testing.T) {
	Convey("doesn't care about state of gl.TEXTURE_2D", t, func() {
		screen := image.Rect(0, 0, 64, 64)
		subscreen := image.Rect(16, 16, 48, 48)
		rendertest.WithGlForTest(screen.Dx(), screen.Dy(), func(sys system.System, queue render.RenderQueueInterface) {
			queue.Queue(func(st render.RenderQueueState) {
				gl.Disable(gl.TEXTURE_2D)
				tex := rendertest.GivenATexture("red/0.png")
				rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeFile, "subred")
		})
		rendertest.WithGlForTest(screen.Dx(), screen.Dy(), func(sys system.System, queue render.RenderQueueInterface) {
			queue.Queue(func(st render.RenderQueueState) {
				gl.Enable(gl.TEXTURE_2D)
				tex := rendertest.GivenATexture("red/0.png")
				rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeFile, "subred")
		})
	})
}
