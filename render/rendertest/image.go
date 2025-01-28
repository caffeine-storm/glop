package rendertest

import (
	"fmt"
	"image"
	"os"
	"path"

	"github.com/runningwild/glop/imgmanip"
)

func MustLoadImage(imageFilePath string) image.Image {
	imageFilePath = path.Join("testdata", imageFilePath)
	file, err := os.Open(imageFilePath)
	if err != nil {
		panic(fmt.Errorf("couldn't open file %q: %w", imageFilePath, err))
	}

	img, _, err := image.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't image.Decode: %w", err))
	}

	return img
}

func MustLoadRGBAImage(imageFilePath string) *image.RGBA {
	return imgmanip.ToRGBA(MustLoadImage(imageFilePath))
}
