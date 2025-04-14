package gui_test

import (
	"testing"

	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRegionClipping(t *testing.T) {
	Convey("region clipping", t, func() {
		rendertest.WithGlForTest(64, 64, func(sys system.System, queue render.RenderQueueInterface) {
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

			queue.Queue(func(render.RenderQueueState) {
				r.PushClipPlanes()
				defer r.PopClipPlanes()

				// Draw a red square across the 'whole' viewport.
				rendertest.DrawRectNdc(-1, -1, 1, 1)

			})
			queue.Purge()

			// Check that no pixels outside the region got drawn to.
			So(queue, rendertest.ShouldLookLikeFile, "red-with-border")
		})
	})
}
