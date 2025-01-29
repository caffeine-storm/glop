package imgmanip_test

import (
	"image"
	"testing"

	"github.com/runningwild/glop/imgmanip"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInvertedCanvasSpecs(t *testing.T) {
	Convey("inverted canvas wrapper", t, func() {
		canvasSize := image.Rect(0, 0, 42, 42)
		wrappedCanvas := image.NewRGBA(canvasSize)
		invCanvas := imgmanip.NewInvertedCanvas(wrappedCanvas)
		So(invCanvas, ShouldNotBeNil)
		So(invCanvas.Bounds(), ShouldEqual, wrappedCanvas.Bounds())
	})
}
