package rendertest

import (
	"fmt"
	"image"
	"io"
	"os"

	"github.com/runningwild/glop/imgmanip"
)

func MustLoadTestImage(testdataref TestDataReference) image.Image {
	return MustLoadImage(testdataref.Path())
}

func MustLoadImage(imageFilePath string) image.Image {
	file, err := os.Open(imageFilePath)
	if err != nil {
		panic(fmt.Errorf("couldn't open file %q: %w", imageFilePath, err))
	}
	defer file.Close()

	return MustLoadImageFromReader(file)
}

func MustLoadImageRGBA(imageFilePath string) *image.RGBA {
	return imgmanip.ToRGBA(MustLoadImage(imageFilePath))
}

func MustLoadImageFromReader(file io.Reader) image.Image {
	img, _, err := image.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't image.Decode: %w", err))
	}

	return img
}
