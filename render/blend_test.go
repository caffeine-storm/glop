package render_test

import (
	"testing"

	"github.com/caffeine-storm/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/smartystreets/goconvey/convey"
)

func TestBlending(t *testing.T) {
	convey.Convey("blending blends", t, func(c convey.C) {
		testbuilder.New().WithExpectation(c, "blend").Run(func() {
			gl.Enable(gl.BLEND)
			gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ZERO, gl.ONE)
			defer gl.Disable(gl.BLEND)

			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.LEQUAL)
			defer gl.Disable(gl.DEPTH_TEST)

			defer gl.Color4f(1, 1, 1, 1)

			render.WithFreshMatrices(func() {
				gl.Begin(gl.QUADS)
				gl.Color4ub(255, 0, 0, 255)
				gl.Vertex3d(-0.75, -0.75, 0)
				gl.Vertex3d(+0.75, -0.75, 0)
				gl.Vertex3d(+0.75, +0.75, 0)
				gl.Vertex3d(-0.75, +0.75, 0)

				gl.Color4ub(0, 0, 255, 128)

				// for i := 50; i > 0; i-- {
				gl.Vertex3d(-0.35, -0.35, 0)
				gl.Vertex3d(+0.35, -0.35, 0)
				gl.Vertex3d(+0.35, +0.35, 0)
				gl.Vertex3d(-0.35, +0.35, 0)
				// }

				gl.End()
			})

			render.LogAndClearGlErrors(glog.ErrorLogger())
		})
	})
}
