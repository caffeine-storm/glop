package rendertest_test

import (
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
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

}

/* DANGER WILL ROBINSON! XXX: this has been crashing windows when running on
* WSL if we don't cache XWindow instances/glxContexts. Run at your PERIL!
* func TestDrawManyRects(t *testing.T) {
	* for i := 0; i < 500; i++ {
		*   TestDrawRect(t)
		* }
	}
*/
