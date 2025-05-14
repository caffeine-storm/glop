package gui_test

import (
	"testing"

	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRegionClipping(t *testing.T) {
	Convey("region clipping", t, func(c C) {
		testbuilder.New().WithExpectation(c, "red-with-border").Run(func() {
			// Set a clipping region to block any drawing outside of a square in the
			// middle.
			r := gui.Region{
				Point: gui.Point{
					X: 4,
					Y: 4,
				},
				Dims: gui.Dims{
					Dx: 56,
					Dy: 56,
				},
			}

			r.PushClipPlanes()
			defer r.PopClipPlanes()

			// Draw a red square across the 'whole' viewport.
			rendertest.DrawRectNdc(-1, -1, 1, 1)
		})
	})
}
