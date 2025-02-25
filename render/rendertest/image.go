package rendertest

import (
	"fmt"
	"image"
	"io"
	"os"
	"path"

	"github.com/runningwild/glop/imgmanip"
)

func MustLoadImageFromTestdataReference(testdataref TestDataReference) image.Image {
	return MustLoadImage(testdataref.Path())
}

func MustLoadTestImage(testimageReference string) image.Image {
	imageFilePath := path.Join("testdata", testimageReference)
	return MustLoadImage(imageFilePath)
}

func MustLoadImage(imageFilePath string) image.Image {
	file, err := os.Open(imageFilePath)
	if err != nil {
		panic(fmt.Errorf("couldn't open file %q: %w", imageFilePath, err))
	}
	defer file.Close()

	return MustLoadImageFromReader(file)
}

func MustLoadImageFromReader(file io.Reader) image.Image {
	img, _, err := image.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't image.Decode: %w", err))
	}

	return img
}

func MustLoadRGBAImage(imageFilePath string) *image.RGBA {
	return imgmanip.ToRGBA(MustLoadTestImage(imageFilePath))
}
