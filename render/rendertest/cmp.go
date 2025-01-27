package rendertest

import (
	"fmt"
	"image"
	"image/color"
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

type TestNumber uint8
type Threshold uint8
type BackgroundColour color.Color

var defaultThreshold = Threshold(3)

// Default background is an opaque black
var defaultBackground = color.RGBA{
	R: 0,
	G: 0,
	B: 0,
	A: 255,
}

func ExpectationFile(testDataKey, fileExt string, testnumber TestNumber) string {
	return path.Join("testdata", testDataKey, fmt.Sprintf("%d.%s", testnumber, fileExt))
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

func drawAsRgbaWithBackground(img image.Image, bg color.Color) *image.RGBA {
	ret := image.NewRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Over)
	return ret
}

func drawAsRgba(img image.Image) *image.RGBA {
	if ret, ok := img.(*image.RGBA); ok {
		return ret
	}

	ret := image.NewRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Src)

	return ret
}

func ImagesAreWithinThreshold(expected, actual image.Image, thresh Threshold, backgroundColour color.Color) bool {
	// Do a size check first so that we don't read out of bounds.
	if expected.Bounds() != actual.Bounds() {
		glog.ErrorLogger().Error("size mismatch", "expected", expected.Bounds(), "actual", actual.Bounds())
		return false
	}

	lhsrgba := drawAsRgbaWithBackground(expected, backgroundColour)
	rhsrgba := drawAsRgba(actual)

	return CompareWithThreshold(lhsrgba.Pix, rhsrgba.Pix, thresh) == 0
}

func CompareWithThreshold(lhs, rhs []byte, threshold Threshold) int {
	llen := len(lhs)
	rlen := len(rhs)

	minLen := llen
	if minLen > rlen {
		minLen = rlen
	}

	for i := 0; i < minLen; i++ {
		diff := int(lhs[i]) - int(rhs[i])
		absdiff := diff
		if absdiff < 0 {
			absdiff = -absdiff
		}
		if absdiff > int(threshold) {
			return diff / absdiff
		}
	}

	if llen == rlen {
		return 0
	}

	if llen < rlen {
		return -1
	}

	return 1
}

func expectReadersMatch(actual, expected io.Reader, threshold Threshold) (bool, []byte) {
	actualBytes, err := io.ReadAll(actual)
	if err != nil {
		panic(fmt.Errorf("couldn't io.ReadAll from 'actual': %w", err))
	}
	expectedBytes, err := io.ReadAll(expected)
	if err != nil {
		panic(fmt.Errorf("couldn't io.ReadAll from 'expected': %w", err))
	}

	if CompareWithThreshold(actualBytes, expectedBytes, threshold) != 0 {
		return false, actualBytes
	}

	return true, nil
}

func expectPixelReadersMatch(actual, expected io.Reader, thresh Threshold, bg color.Color) (bool, image.Image) {
	expectedImage, _, _ := readPng(expected)
	var actualScreenWidth, actualScreenHeight uint32

	// Read all the pixels from the input source
	actualImage := image.NewRGBA(image.Rect(0, 0, int(actualScreenWidth), int(actualScreenHeight)))
	var err error
	actualImage.Pix, err = io.ReadAll(actual)
	if err != nil {
		panic(fmt.Errorf("couldn't read from 'actual': %w", err))
	}

	if !ImagesAreWithinThreshold(expectedImage, actualImage, thresh, bg) {
		return false, actualImage
	}

	return true, nil
}

// Verify that the 'actualImage' is a match to our expected image to within a
// threshold. We need the fuzzy matching because OpenGL doesn't guarantee exact
// pixel-to-pixel matches across different hardware/driver combinations. It
// should be close, though!
func expectPixelsMatch(actualImage image.Image, pngFileExpected string, thresh Threshold, bg color.Color) (bool, string) {
	pngFile, err := os.Open(pngFileExpected)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", pngFileExpected, err))
	}
	defer pngFile.Close()

	expectedImage, _, _ := readPng(pngFile)
	if !ImagesAreWithinThreshold(expectedImage, actualImage, thresh, bg) {
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

func getTestNumberFromArgs(args []interface{}) TestNumber {
	// TODO(tmckee): why start at second element? we should test this!
	for i := 1; i < len(args); i++ {
		if val, found := args[i].(TestNumber); found {
			return val
		}
	}

	return TestNumber(0)
}

func getThresholdFromArgs(args []interface{}) Threshold {
	for i := 1; i < len(args); i++ {
		if val, found := args[i].(Threshold); found {
			return val
		}
	}

	return defaultThreshold
}

func getBackgroundFromArgs(args []interface{}) (BackgroundColour, bool) {
	for i := 1; i < len(args); i++ {
		if val, found := args[i].(BackgroundColour); found {
			return val, true
		}
	}
	return defaultBackground, false
}

func ShouldLookLike(actual interface{}, expected ...interface{}) string {
	actualReader, ok := actual.(io.Reader)
	if !ok {
		panic(fmt.Errorf("ShouldLookLike needs a io.Reader but got %T", actual))
	}
	expectedReader, ok := expected[0].(io.Reader)
	if !ok {
		panic(fmt.Errorf("ShouldLookLike needs a string but got %T", expected[0]))
	}

	// Did someone pass a Threshold?
	threshold := getThresholdFromArgs(expected)

	ok, _ = expectReadersMatch(actualReader, expectedReader, threshold)
	if ok {
		return ""
	}

	// TODO(tmckee): create and report a pair of temp files for debuggability.
	return fmt.Sprintf("io.Readers mismatched")
}

func imageShouldLookLike(actualImage *image.RGBA, expected ...interface{}) string {
	testDataKey, ok := expected[0].(string)
	if !ok {
		panic(fmt.Errorf("ShouldLookLikeFile needs a string but got %T", expected[0]))
	}

	// For table-tests, usage is
	//   'So(something, ShouldLookLikeFile, "test-case-family", TestNumber(N))'
	// Use a default 'testnumber = 0' for non-table tests.
	testnumber := getTestNumberFromArgs(expected)

	filename := ExpectationFile(testDataKey, "png", testnumber)

	bg, _ := getBackgroundFromArgs(expected)

	thresh := getThresholdFromArgs(expected)
	ok, rejectFile := expectPixelsMatch(actualImage, filename, thresh, bg)
	if ok {
		return ""
	}

	return fmt.Sprintf("frame buffer mismatch; see %s", rejectFile)
}

func backBufferShouldLookLike(queue render.RenderQueueInterface, expected ...interface{}) string {
	var actualImage *image.RGBA

	var currentBackground color.Color
	// Read all the pixels from the framebuffer through OpenGL
	queue.Queue(func(render.RenderQueueState) {
		_, _, actualScreenWidth, actualScreenHeight := debug.GetViewport()
		actualImage = debug.ScreenShotRgba(int(actualScreenWidth), int(actualScreenHeight))
		currentBackground = GetCurrentBackgroundColor()
	})
	queue.Purge()

	_, ok := getBackgroundFromArgs(expected)
	if !ok {
		expected = append(expected, BackgroundColour(currentBackground))
	}

	// When screen shotting, we only read opaque pixels; if the expectation file
	// has transparent portions, we need to not compare those pixels.
	return imageShouldLookLike(actualImage, expected...)
}

func ShouldLookLikeFile(actual interface{}, expected ...interface{}) string {
	switch v := actual.(type) {
	case render.RenderQueueInterface:
		return backBufferShouldLookLike(v, expected...)
	case *image.RGBA:
		_, foundBg := getBackgroundFromArgs(expected)
		if !foundBg {
			expected = append(expected, BackgroundColour(color.RGBA{}))
		}
		return imageShouldLookLike(v, expected...)
	default:
		panic(fmt.Errorf("ShouldLookLikeFile needs a *image.RGBA or render.RenderQueueInterface but got %T", actual))
	}
}

func ShouldLookLikeText(actual interface{}, expected ...interface{}) string {
	testDataKey, ok := expected[0].(string)
	if !ok {
		panic(fmt.Errorf("ShouldLookLikeText needs a string but got %T", expected[0]))
	}

	expected[0] = "text/" + testDataKey
	return ShouldLookLikeFile(actual, expected...)
}
