package imgmanip_test

import (
	"image"
	"testing"

	"github.com/caffeine-storm/glop/imgmanip"
	"github.com/caffeine-storm/glop/render/rendertest"
)

func givenAnImage() image.Image {
	return rendertest.MustLoadTestImage("checker")
}

func TestScale(t *testing.T) {
	img := givenAnImage()

	t.Run("no-op scale", func(t *testing.T) {
		notScaled := imgmanip.Scale(img, 1, 1)

		if notScaled.Bounds() != img.Bounds() {
			t.Fatalf("if we're not scaling, the bounds shouldn't change")
		}

		if img == notScaled {
			t.Fatalf("we shouldn't return a reference to an input; prefer immutable values being passed around")
		}
	})

	t.Run("scale up", func(t *testing.T) {
		bigger := imgmanip.Scale(img, 2, 1)

		if bigger.Bounds().Dx() != 2*img.Bounds().Dx() {
			t.Fatalf("did not scale by 2 in the X dimension (old: %d, new: %d)", img.Bounds().Dx(), bigger.Bounds().Dx())
		}
	})

	t.Run("scale down", func(t *testing.T) {
		smaller := imgmanip.Scale(img, 0.5, 1)

		if 2*smaller.Bounds().Dx() != img.Bounds().Dx() {
			t.Fatalf("did not scale by 1/2 in the X dimension (old: %d, new: %d)", img.Bounds().Dx(), smaller.Bounds().Dx())
		}
	})
}
