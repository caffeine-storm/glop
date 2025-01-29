package imgmanip_test

import (
	"image"
	"image/draw"
	"testing"

	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render/rendertest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInvertedCanvasSpecs(t *testing.T) {
	Convey("inverted canvas wrapper", t, func() {
		dx, dy := 64, 64
		canvasSize := image.Rect(0, 0, dx, dy)
		wrappedCanvas := image.NewRGBA(canvasSize)
		invCanvas := imgmanip.NewInvertedCanvas(wrappedCanvas)
		So(invCanvas, ShouldNotBeNil)
		So(invCanvas.Bounds(), ShouldEqual, wrappedCanvas.Bounds())

		Convey("draws things", func() {
			checkers := rendertest.MustLoadImage("checker/0.png")
			draw.Draw(invCanvas, invCanvas.Bounds(), checkers, image.Point{}, draw.Src)

			Convey("upside down in the wrapped image", func() {
				flipped := imgmanip.VertFlipped{Image: wrappedCanvas}

				So(wrappedCanvas, rendertest.ShouldLookLike, flipped)
			})
		})
	})
}
