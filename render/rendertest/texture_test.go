package rendertest_test

import (
	"image"
	"testing"
	"unsafe"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
)

// TODO(tmckee:clean): copy/paste from debug/dump_buffer_test.go
func givenABufferWithData(data []float32) gl.Buffer {
	result := gl.GenBuffer()
	result.Bind(gl.ARRAY_BUFFER)

	floatSize := int(unsafe.Sizeof(float32(0)))
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(data), data, gl.STATIC_DRAW)

	return result
}

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

	Convey("doesn't care about state of gl.ELEMENT_ARRAY_BUFFER", t, func() {
		screen := image.Rect(0, 0, 64, 64)
		subscreen := image.Rect(16, 16, 48, 48)
		rendertest.WithGlForTest(screen.Dx(), screen.Dy(), func(sys system.System, queue render.RenderQueueInterface) {
			queue.Queue(func(st render.RenderQueueState) {
				gl.Buffer(0).Bind(gl.ELEMENT_ARRAY_BUFFER)
				tex := rendertest.GivenATexture("red/0.png")
				rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeFile, "subred")
		})
		rendertest.WithGlForTest(screen.Dx(), screen.Dy(), func(sys system.System, queue render.RenderQueueInterface) {
			queue.Queue(func(st render.RenderQueueState) {
				someStaleBuffer := givenABufferWithData([]float32{
					77, 55, 44, 33, 22, 11,
				})
				someStaleBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
				tex := rendertest.GivenATexture("red/0.png")
				rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeFile, "subred")
		})

	})
}
