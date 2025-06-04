package imgmanip_test

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/runningwild/glop/imgmanip"
)

func TestIsTransparent(t *testing.T) {
	sixteenBySixteen := image.Rect(0, 0, 16, 16)
	testcases := []struct {
		name   string
		canvas draw.Image
	}{
		{
			"image.RGBA",
			image.NewRGBA(sixteenBySixteen),
		},
		{
			"imgmanip.GrayAlpha",
			imgmanip.NewGrayAlpha(sixteenBySixteen),
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			canvas := testcase.canvas
			rgba := imgmanip.ToNRGBA(canvas)
			if !imgmanip.IsTransparent(rgba) {
				t.Fatalf("a fresh image should be transparent!")
			}

			draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
			rgba = imgmanip.ToNRGBA(canvas)
			if imgmanip.IsTransparent(rgba) {
				t.Fatalf("after drawing something, the image should not be fully transparent")
			}
		})
	}
}
