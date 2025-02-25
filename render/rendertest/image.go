package rendertest

import (
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"strings"

	"github.com/runningwild/glop/imgmanip"
)

func MustLoadImageFromTestdataReference(testdataref TestDataReference) image.Image {
	p := testdataref.Path()
	sliced := strings.SplitN(p, "/", 2)
	return MustLoadImage(path.Join(sliced[1:]...))
}

// TODO(tmckee): clean this up; MustLoadImage ought to not mess with the path.
func MustLoadImage(imageFilePath string) image.Image {
	imageFilePath = path.Join("testdata", imageFilePath)
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
	return imgmanip.ToRGBA(MustLoadImage(imageFilePath))
}
