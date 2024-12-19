package rendertest

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
)

type TestNumber uint8
type Threshold uint8

var defaultThreshold = Threshold(3)

func ExpectationFile(testDataKey, fileExt string, testnumber TestNumber) string {
	_, cmpGoFilePath, _, _ := runtime.Caller(0)
	projectDir := path.Clean(path.Join(path.Dir(cmpGoFilePath), "..", ".."))
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("couldn't os.Getwd(): %w", err))
	}

	relTestDataDir, err := filepath.Rel(workDir, path.Join(projectDir, "testdata"))
	if projectDir == "/" {
		panic("should not have found ./rendertest/ to be /")
	}
	return path.Join(relTestDataDir, testDataKey, fmt.Sprintf("%d.%s", testnumber, fileExt))
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

func imagesAreWithinThreshold(expected, actual image.Image, thresh Threshold) bool {
	// Do a size check first so that we don't read out of bounds.
	if expected.Bounds() != actual.Bounds() {
		glog.ErrorLogger().Error("size mismatch", "expected", expected.Bounds(), "actual", actual.Bounds())
		return false
	}

	lhsrgba := drawAsRgba(expected)
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

func expectPixelReadersMatch(actual, expected io.Reader, thresh Threshold) (bool, image.Image) {
	expectedImage, _, _ := readPng(expected)
	var actualScreenWidth, actualScreenHeight uint32

	// Read all the pixels from the input source
	actualImage := image.NewRGBA(image.Rect(0, 0, int(actualScreenWidth), int(actualScreenHeight)))
	var err error
	actualImage.Pix, err = io.ReadAll(actual)
	if err != nil {
		panic(fmt.Errorf("couldn't read from 'actual': %w", err))
	}

	if !imagesAreWithinThreshold(expectedImage, actualImage, thresh) {
		return false, actualImage
	}

	return true, nil
}

// Verify that the 'actualImage' is a match to our expected image to within a
// threshold. We need the fuzzy matching because OpenGL doesn't guarantee exact
// pixel-to-pixel matches across different hardware/driver combinations. It
// should be close, though!
func expectPixelsMatch(actualImage image.Image, pngFileExpected string, thresh Threshold) (bool, string) {
	pngFile, err := os.Open(pngFileExpected)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", pngFileExpected, err))
	}
	defer pngFile.Close()

	expectedImage, _, _ := readPng(pngFile)
	if !imagesAreWithinThreshold(expectedImage, actualImage, thresh) {
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

func ShouldLookLikeFile(actual interface{}, expected ...interface{}) string {
	var actualImage *image.RGBA
	switch v := actual.(type) {
	case render.RenderQueueInterface:
		// If we're given a RenderQueueInterface, take a debug-screenshot of the
		// associated back-buffer.
		queue := v
		// Read all the pixels from the framebuffer through OpenGL
		queue.Queue(func(render.RenderQueueState) {
			_, _, actualScreenWidth, actualScreenHeight := debug.GetViewport()
			actualImage = debug.ScreenShotRgba(int(actualScreenWidth), int(actualScreenHeight))
		})
		queue.Purge()
	case *image.RGBA:
		actualImage = v
	default:
		panic(fmt.Errorf("ShouldLookLikeFile needs a *image.RGBA or render.RenderQueueInterface but got %T", actual))
	}

	testDataKey, ok := expected[0].(string)
	if !ok {
		panic(fmt.Errorf("ShouldLookLikeFile needs a string but got %T", expected[0]))
	}

	// For table-tests, usage is
	//   'So(render, ShouldLookLike, "test-case-family", TestNumber(N))'
	// Use a default 'testnumber = 0' for non-table tests.
	testnumber := getTestNumberFromArgs(expected)

	filename := ExpectationFile(testDataKey, "png", testnumber)

	thresh := getThresholdFromArgs(expected)
	ok, rejectFile := expectPixelsMatch(actualImage, filename, thresh)
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
