package imgmanip_test

import (
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/caffeine-storm/glop/imgmanip"
)

func withTempFile(tempname string, fn func(*os.File)) {
	f, err := os.CreateTemp("", tempname)
	if err != nil {
		panic(fmt.Errorf("couldn't os.CreateTemp(%q): %w", tempname, err))
	}
	defer os.Remove(f.Name())

	fn(f)
}

func TestDumpImage(t *testing.T) {
	t.Run("can dump an image to a path", func(t *testing.T) {
		var genericImage image.Image
		genericImage = givenAnImage()

		withTempFile("dump.png", func(f *os.File) {
			imgmanip.MustDumpImage(genericImage, f.Name())
		})
	})
}
