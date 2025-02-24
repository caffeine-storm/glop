package imgmanip_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/runningwild/glop/imgmanip"
)

func givenAnImage() image.Image {
	return image.NewUniform(color.RGBA{
		R: 127,
		G: 127,
		B: 127,
		A: 127,
	})
}

func TestScale(t *testing.T) {
	img := givenAnImage()

	t.Run("no-op scale", func(t *testing.T) {
		notScaled := imgmanip.Scale(img, 1, 1)

		if notScaled.Bounds() != img.Bounds() {
			t.Fatalf("if we're not scaling, the bounds shouldn't change")
		}
	})
}
