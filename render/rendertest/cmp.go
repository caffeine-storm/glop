package rendertest

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
)

func ExpectationFile(testDataKey, fileExt string, testnumber int) string {
	return fmt.Sprintf("../testdata/text/%s/%d.%s", testDataKey, testnumber, fileExt)
}

// Return the given file but with a '.rej' component to signify a 'rejection'.
func MakeRejectName(exp, suffix string) string {
	dir, expectedFileName := path.Split(exp)
	rejectFileNameBase, ok := strings.CutSuffix(expectedFileName, suffix)
	if !ok {
		panic(fmt.Errorf("need a %s file, got %s", suffix, exp))
	}
	return path.Join(dir, rejectFileNameBase+".rej"+suffix)
}

func readPixels(width, height int) ([]byte, error) {
	ret := make([]byte, width*height*4)
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, ret)
	return ret, nil
}

// Load a .png
func readPng(reader io.Reader) (image.Image, int, int) {
	img, err := png.Decode(reader)
	if err != nil {
		panic(fmt.Errorf("image.Decode failed: %w", err))
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	return img, width, height
}

func verticalFlipRgbaPixels(rgbaPixels []byte, width, height int) []byte {
	result := make([]byte, len(rgbaPixels))

	// Convert from bottom-first-row to top-first-row.
	byteWidth := width * 4
	for row := 0; row < height; row++ {
		resultRowIdx := row * byteWidth
		resultRowEnd := resultRowIdx + byteWidth
		inputRowIdx := (height - row - 1) * byteWidth
		inputRowEnd := inputRowIdx + byteWidth

		copy(result[resultRowIdx:resultRowEnd], rgbaPixels[inputRowIdx:inputRowEnd])
	}

	return result
}

func drawAsRgba(img image.Image) *image.RGBA {
	if ret, ok := img.(*image.RGBA); ok {
		return ret
	}

	ret := image.NewRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Src)

	return ret
}

func identicalImages(lhs, rhs image.Image) bool {
	lhsrgba := drawAsRgba(lhs)
	rhsrgba := drawAsRgba(rhs)

	return bytes.Compare(lhsrgba.Pix, rhsrgba.Pix) == 0
}

// Verify that the framebuffer's contents match our expected image.
func expectPixelsMatch(queue render.RenderQueueInterface, pngFileExpected string) (bool, string) {
	pngFile, err := os.Open(pngFileExpected)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", pngFileExpected, err))
	}
	defer pngFile.Close()

	expectedImage, screenWidth, screenHeight := readPng(pngFile)

	// Read all the pixels from the framebuffer through OpenGL
	var frameBufferBytes []byte
	queue.Queue(func(render.RenderQueueState) {
		frameBufferBytes, err = readPixels(screenWidth, screenHeight)
		if err != nil {
			panic(fmt.Errorf("couldn't readPixels: %w", err))
		}
	})
	queue.Purge()

	actualImage := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
	actualImage.Pix = verticalFlipRgbaPixels(frameBufferBytes, screenWidth, screenHeight)
	if !identicalImages(expectedImage, actualImage) {
		rejectFileName := MakeRejectName(pngFileExpected, ".png")
		rejectFile, err := os.Create(rejectFileName)
		if err != nil {
			panic(fmt.Errorf("couldn't open rejectFileName %q: %w", rejectFileName, err))
		}
		defer rejectFile.Close()

		err = png.Encode(rejectFile, actualImage)
		if err != nil {
			panic(fmt.Errorf("couldn't write rejection file: %s: %w", rejectFileName, err))
		}

		return false, rejectFileName
	}

	return true, ""
}

func ShouldLookLike(actual interface{}, expected ...interface{}) string {
	render, ok := actual.(render.RenderQueueInterface)
	if !ok {
		panic(fmt.Errorf("ShouldLookLike needs a render queue but got %T", actual))
	}
	testDataKey, ok := expected[0].(string)
	if !ok {
		panic(fmt.Errorf("ShouldLookLike needs a string but got %T", expected[0]))
	}

	// For table-tests, usage is
	//   'So(render, ShouldLookLike, "test-case-family", testNumberN)'
	// Use a default 'testnumber = 0' for non-table tests.
	testnumber := 0
	if len(expected) > 1 {
		testnumber, ok = expected[1].(int)
		if !ok {
			panic(fmt.Errorf("ShouldLookLike needs a string but got %T", expected[0]))
		}
	}

	filename := ExpectationFile(testDataKey, "png", testnumber)

	ok, rejectFile := expectPixelsMatch(render, filename)
	if ok {
		return ""
	}

	return fmt.Sprintf("frame buffer mismatch; see %s", rejectFile)
}
