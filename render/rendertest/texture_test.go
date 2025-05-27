package rendertest_test

import (
	"image"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDrawTexturedQuad(t *testing.T) {
	Convey("doesn't care about state of gl.TEXTURE_2D", t, func(c C) {
		subscreen := image.Rect(16, 16, 48, 48)
		testbuilder.WithExpectation(c, "subred", func(st render.RenderQueueState) {
			gl.Disable(gl.TEXTURE_2D)
			tex := rendertest.GivenATexture("red/0.png")
			rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
		})
		testbuilder.WithExpectation(c, "subred", func(st render.RenderQueueState) {
			gl.Enable(gl.TEXTURE_2D)
			tex := rendertest.GivenATexture("red/0.png")
			rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
		})
	})

	Convey("doesn't care about state of gl.ELEMENT_ARRAY_BUFFER", t, func(c C) {
		subscreen := image.Rect(16, 16, 48, 48)
		testbuilder.WithExpectation(c, "subred", func(st render.RenderQueueState) {
			gl.Buffer(0).Bind(gl.ELEMENT_ARRAY_BUFFER)
			tex := rendertest.GivenATexture("red/0.png")
			rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
		})
		testbuilder.WithExpectation(c, "subred", func(st render.RenderQueueState) {
			someStaleBuffer := rendertest.GivenABufferWithData([]float32{
				77, 55, 44, 33, 22, 11,
			})
			someStaleBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
			tex := rendertest.GivenATexture("red/0.png")
			rendertest.DrawTexturedQuad(subscreen, tex, st.Shaders())
		})
	})

	Convey("GivenATexture must panic if there's no image file", t, func(c C) {
		mustPanic := func() {
			tex := rendertest.GivenATexture("thisfiledoesnotexist.nope")
			tex.Delete()
		}
		c.So(func() {
			testbuilder.New().Run(func() {
				mustPanic()
			})
		}, ShouldPanic)
		Convey("With the deprecated helpers too", func(c C) {
			c.So(func() {
				rendertest.DeprecatedWithGl(func() {
					mustPanic()
				})
			}, ShouldPanic)
		})
	})
}
