package imgmanip_test

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/caffeine-storm/glop/imgmanip"
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
			"image.NRGBA",
			image.NewNRGBA(sixteenBySixteen),
		},
		{
			"imgmanip.GrayAlpha",
			imgmanip.NewGrayAlpha(sixteenBySixteen),
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			canvas := testcase.canvas
			nrgba := imgmanip.ToNRGBA(canvas)
			if !imgmanip.IsTransparent(nrgba) {
				t.Fatalf("a fresh image should be transparent!")
			}

			draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
			nrgba = imgmanip.ToNRGBA(canvas)
			if imgmanip.IsTransparent(nrgba) {
				t.Fatalf("after drawing something, the image should not be fully transparent")
			}
		})
	}
}
