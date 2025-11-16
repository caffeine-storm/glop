package rendertest_test

import (
	"testing"

	"github.com/caffeine-storm/gl"
	"github.com/caffeine-storm/glop/render"
	"github.com/caffeine-storm/glop/render/rendertest"
	"github.com/caffeine-storm/glop/render/rendertest/testbuilder"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDrawRect(t *testing.T) {
	Convey("rendertest.DrawRect should work", t, DrawRectSpec)
}

func RunTest(c C, withTex2d bool) {
	testbuilder.New().WithExpectation(c, "red64").Run(func() {
		if withTex2d {
			gl.Enable(gl.TEXTURE_2D)
		} else {
			gl.Disable(gl.TEXTURE_2D)
		}
		rendertest.BlankAndDrawRectNdc(-1, -1, 1, 1)
	})
}

func DrawRectSpec() {
	Convey("with gl.TEXTURE_2D enabled", func(c C) {
		RunTest(c, true)
	})

	Convey("with gl.TEXTURE_2D disabled", func(c C) {
		RunTest(c, false)
	})

	Convey("DrawRect uses 'incoming' coordinates", func(c C) {
		testbuilder.New().WithExpectation(c, "subred").Run(func() {
			modelView := &render.Matrix{}
			modelView.Identity()
			modelView.Scaling(0.5, 0.5, 1)
			render.WithMatrixInMode(modelView, render.MatrixModeModelView, func() {
				rendertest.DrawRect(-1, -1, 1, 1)
			})
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
