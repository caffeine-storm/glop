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
	checkers := rendertest.MustLoadTestImage("checker/0.png")
	Convey("inverted canvas wrapper", t, func() {
		canvasSize := checkers.Bounds()
		wrappedCanvas := image.NewRGBA(canvasSize)
		invCanvas := imgmanip.NewInvertedCanvas(wrappedCanvas)
		So(invCanvas, ShouldNotBeNil)
		So(invCanvas.Bounds(), ShouldEqual, wrappedCanvas.Bounds())

		Convey("draws things", func() {
			draw.Draw(invCanvas, invCanvas.Bounds(), checkers, image.Point{}, draw.Src)

			Convey("upside down in the wrapped image", func() {
				flipped := imgmanip.VertFlipped{Image: checkers}

				So(wrappedCanvas, rendertest.ShouldLookLike, flipped)
			})
		})
	})
}
