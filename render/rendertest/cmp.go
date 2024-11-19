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
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
)

func ExpectationFile(testDataKey, fileExt string, testnumber int) string {
	return fmt.Sprintf("../testdata/%s/%d.%s", testDataKey, testnumber, fileExt)
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

func drawAsRgba(img image.Image) *image.RGBA {
	if ret, ok := img.(*image.RGBA); ok {
		return ret
	}

	ret := image.NewRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Src)

	return ret
}

func identicalImages(expected, actual image.Image) bool {
	// Do a size check first so that we don't read out of bounds.
	if expected.Bounds() != actual.Bounds() {
		glog.ErrorLogger().Error("size mismatch", "expected", expected.Bounds(), "actual", actual.Bounds())
		return false
	}

	lhsrgba := drawAsRgba(expected)
	rhsrgba := drawAsRgba(actual)

	return bytes.Compare(lhsrgba.Pix, rhsrgba.Pix) == 0
}

func expectReadersMatch(actual, expected io.Reader) (bool, []byte) {
	actualBytes, err := io.ReadAll(actual)
	if err != nil {
		panic(fmt.Errorf("couldn't io.ReadAll from 'actual': %w", err))
	}
	expectedBytes, err := io.ReadAll(expected)
	if err != nil {
		panic(fmt.Errorf("couldn't io.ReadAll from 'expected': %w", err))
	}

	if bytes.Compare(actualBytes, expectedBytes) != 0 {
		return false, actualBytes
	}

	return true, nil
}

// Verify that the framebuffer's contents match our expected image.
func expectPixelsMatch(queue render.RenderQueueInterface, pngFileExpected string) (bool, string) {
	pngFile, err := os.Open(pngFileExpected)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", pngFileExpected, err))
	}
	defer pngFile.Close()

	expectedImage, _, _ := readPng(pngFile)
	var actualScreenWidth, actualScreenHeight uint32

	// Read all the pixels from the framebuffer through OpenGL
	frameBufferBytes := &bytes.Buffer{}
	queue.Queue(func(render.RenderQueueState) {
		_, _, actualScreenWidth, actualScreenHeight = debug.GetViewport()
		debug.ScreenShotRgba(int(actualScreenWidth), int(actualScreenHeight), frameBufferBytes)
	})
	queue.Purge()

	actualImage := image.NewRGBA(image.Rect(0, 0, int(actualScreenWidth), int(actualScreenHeight)))
	actualImage.Pix = frameBufferBytes.Bytes()

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
	actualReader, ok := actual.(io.Reader)
	if !ok {
		panic(fmt.Errorf("ShouldLookLike needs a io.Reader but got %T", actual))
	}
	expectedReader, ok := expected[0].(io.Reader)
	if !ok {
		panic(fmt.Errorf("ShouldLookLikeFile needs a string but got %T", expected[0]))
	}

	ok, _ = expectReadersMatch(actualReader, expectedReader)
	if ok {
		return ""
	}

	return fmt.Sprintf("io.Readers mismatched")
}

func ShouldLookLikeFile(actual interface{}, expected ...interface{}) string {
	render, ok := actual.(render.RenderQueueInterface)
	if !ok {
		panic(fmt.Errorf("ShouldLookLikeFile needs a render queue but got %T", actual))
	}
	testDataKey, ok := expected[0].(string)
	if !ok {
		panic(fmt.Errorf("ShouldLookLikeFile needs a string but got %T", expected[0]))
	}

	// For table-tests, usage is
	//   'So(render, ShouldLookLike, "test-case-family", testNumberN)'
	// Use a default 'testnumber = 0' for non-table tests.
	testnumber := 0
	if len(expected) > 1 {
		testnumber, ok = expected[1].(int)
		if !ok {
			panic(fmt.Errorf("ShouldLookLikeFile needs a string but got %T", expected[0]))
		}
	}

	filename := ExpectationFile(testDataKey, "png", testnumber)

	ok, rejectFile := expectPixelsMatch(render, filename)
	if ok {
		return ""
	}

	return fmt.Sprintf("frame buffer mismatch; see %s", rejectFile)
}

func ShouldLookLikeText(actual interface{}, expected ...interface{}) string {
	testDataKey, ok := expected[0].(string)
	if !ok {
		panic(fmt.Errorf("ShouldLookLikeText needs a string but got %T", expected[0]))
	}

	expected[0] = "text/" + testDataKey
	return ShouldLookLikeFile(actual, expected...)
}
