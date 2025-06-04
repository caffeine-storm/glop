package imgmanip

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

func MustDumpImage(img image.Image, filePath string) {
	// TODO(tmckee:43): we should warn if someone passes in not-an-NRGBA because
	// the encoding could be lossy in that case.
	f, err := os.Create(filePath)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Create(%q): %w", filePath, err))
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		panic(fmt.Errorf("couldn't png.Encode(%q): %w", filePath, err))
	}
}
