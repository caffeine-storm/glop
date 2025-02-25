package rendertest

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render"
)

type TestNumber uint8
type FileExtension string
type Threshold uint8
type BackgroundColour color.Color
type MakeRejectFiles bool

var defaultTestNumber = TestNumber(0)
var defaultFileExtension = FileExtension("png")
var defaultThreshold = Threshold(3)

// Default background is an opaque black
var defaultBackground = color.RGBA{
	R: 0,
	G: 0,
	B: 0,
	A: 255,
}
var transparent = color.RGBA{}

var defaultMakeRejectFiles = MakeRejectFiles(true)

func ExpectationFile(testDataKey TestDataReference, fileExt string, testnumber TestNumber) string {
	return testDataKey.Path(FileExtension(fileExt), testnumber)
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

func ImagesAreWithinThreshold(actual, expected image.Image, thresh Threshold, backgroundColour color.Color) bool {
	// Do a size check first so that we don't read out of bounds.
	if actual.Bounds() != expected.Bounds() {
		glog.ErrorLogger().Error("size mismatch", "actual", actual.Bounds(), "expected", expected.Bounds())
		return false
	}

	lhsrgba := imgmanip.DrawAsRgbaWithBackground(expected, backgroundColour)
	rhsrgba := imgmanip.ToRGBA(actual)

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

	if !ImagesAreWithinThreshold(actualImage, expectedImage, thresh, bg) {
		return false, actualImage
	}

	return true, nil
}

// Verify that the 'actualImage' is a match to our expected image to within a
// threshold. We need the fuzzy matching because OpenGL doesn't guarantee exact
// pixel-to-pixel matches across different hardware/driver combinations. It
// should be close, though! To support transparency in our testdata files, we
// also take a background to use as needed.
func expectPixelsMatchFile(actualImage image.Image, pngFileExpected string, thresh Threshold, bg color.Color) bool {
	pngFile, err := os.Open(pngFileExpected)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", pngFileExpected, err))
	}
	defer pngFile.Close()

	expectedImage, _, _ := readPng(pngFile)
	return ImagesAreWithinThreshold(actualImage, expectedImage, thresh, bg)
}

// For the given slice of trailing arguments to a 'convey.So' call, look for a
// value with the same type as 'defaultValue'. If found, assign it to the
// pointer wrapped in 'output', otherwise, assign 'defaultValue' to the pointer
// wrapped in 'output'. Return true iff the value written to 'output' was found
// in 'args'.
func getFromArgs(args []interface{}, defaultValue interface{}, output interface{}) bool {
	defaultReflectValue := reflect.ValueOf(defaultValue)
	targetType := defaultReflectValue.Type()
	outPtr := reflect.ValueOf(output).Elem()

	// We start at the second element because the first element always has to be
	// the testdata 'key'.
	for i := 1; i < len(args); i++ {
		val := reflect.ValueOf(args[i])
		if val.Type() == targetType {
			outPtr.Set(val)
			return true
		}
	}

	outPtr.Set(defaultReflectValue)
	return false
}

func getTestDataKeyFromArgs(args []interface{}) TestDataReference {
	// The only valid spot to look for a test data reference is at the head of
	// the slice.
	if len(args) < 1 {
		panic(fmt.Errorf("need a non-empty slice of options for getting the test data key"))
	}

	// It might be a TestDataKey already, otherwise it has to be a string.
	switch v := args[0].(type) {
	case string:
		return NewTestdataReference(v)
	case TestDataReference:
		return v
	}

	panic(fmt.Errorf("expected type string or TestDataReference, got %T", args[0]))
}

func getTestNumberFromArgs(args []interface{}) TestNumber {
	var result TestNumber
	getFromArgs(args, defaultTestNumber, &result)
	return result
}

func getThresholdFromArgs(args []interface{}) Threshold {
	var result Threshold
	getFromArgs(args, defaultThreshold, &result)
	return result
}

func getBackgroundFromArgs(args []interface{}) (BackgroundColour, bool) {
	var result BackgroundColour
	found := getFromArgs(args, defaultBackground, &result)
	return result, found
}

func getFileExtensionFromArgs(args []interface{}) FileExtension {
	var result FileExtension
	getFromArgs(args, defaultFileExtension, &result)
	return result
}

func getMakeRejectFilesFromArgs(args []interface{}) MakeRejectFiles {
	var result MakeRejectFiles
	getFromArgs(args, defaultMakeRejectFiles, &result)
	return result
}

func readerShouldLookLike(actual, expected io.Reader, args ...interface{}) string {
	// Did someone pass a Threshold?
	threshold := getThresholdFromArgs(args)

	ok, _ := expectReadersMatch(actual, expected, threshold)
	if ok {
		return ""
	}

	// TODO(tmckee): create and report a pair of temp files for debuggability.
	return fmt.Sprintf("io.Readers mismatched")
}

func ShouldLookLike(actual interface{}, expected ...interface{}) string {
	switch v := actual.(type) {
	case io.Reader:
		expectedReader, ok := expected[0].(io.Reader)
		if !ok {
			panic(fmt.Errorf("ShouldLookLike needs matching actual/expected types; actual had type %T, but expected had type %T", actual, expected[0]))
		}
		return readerShouldLookLike(v, expectedReader, expected...)
	case image.Image:
		expectedImage, ok := expected[0].(image.Image)
		if !ok {
			panic(fmt.Errorf("ShouldLookLike needs matching actual/expected types; actual had type %T, but expected had type %T", actual, expected[0]))
		}

		// Use a transparent background for image-to-image comparison by default.
		expected = append(expected, BackgroundColour(transparent))
		return imageShouldLookLike(expectedImage, v, expected...)
	default:
		panic(fmt.Errorf("ShouldLookLike needs either io.Readers or image.Images but got %T", actual))
	}
}

func imageShouldLookLike(actualImage, expectedImage image.Image, expected ...interface{}) string {
	bg, _ := getBackgroundFromArgs(expected)

	thresh := getThresholdFromArgs(expected)
	if ImagesAreWithinThreshold(actualImage, expectedImage, thresh, bg) {
		return ""
	}

	return "image mismatch; rejection file creation elided"
}

func imageShouldLookLikeFile(actualImage image.Image, expected ...interface{}) string {
	testDataKey, ok := expected[0].(TestDataReference)
	if !ok {
		panic(fmt.Errorf("imageShouldLookLikeFile needs a TestDataReference but got %T", expected[0]))
	}

	// For table-tests, usage is
	//   'So(something, ShouldLookLikeFile, "test-case-family", TestNumber(N))'
	// Use a default 'testnumber = 0' for non-table tests.
	testnumber := getTestNumberFromArgs(expected)

	expectedFileName := ExpectationFile(testDataKey, "png", testnumber)

	bg, _ := getBackgroundFromArgs(expected)

	thresh := getThresholdFromArgs(expected)
	if expectPixelsMatchFile(actualImage, expectedFileName, thresh, bg) {
		return ""
	}

	doMakeRejectFiles := getMakeRejectFilesFromArgs(expected)
	if doMakeRejectFiles == true {
		rejectFileName := MakeRejectName(expectedFileName, ".png")
		rejectFile, err := os.Create(rejectFileName)
		if err != nil {
			panic(fmt.Errorf("couldn't open rejectFileName %q: %w", rejectFileName, err))
		}
		defer rejectFile.Close()

		err = png.Encode(rejectFile, actualImage)
		if err != nil {
			panic(fmt.Errorf("couldn't write rejection file: %s: %w", rejectFileName, err))
		}
		return fmt.Sprintf("image mismatch; see %s", rejectFileName)
	} else {
		return "image mismatch; rejection file creation elided"
	}
}

func backBufferShouldLookLike(queue render.RenderQueueInterface, expected ...interface{}) string {
	var currentBackground color.Color = defaultBackground
	r, g, b, a := currentBackground.RGBA()
	fallbackPixel := []uint8{
		uint8(r), uint8(g), uint8(b), uint8(a),
	}

	var actualImage *image.RGBA = &image.RGBA{
		Pix:    fallbackPixel,
		Stride: 4,
		Rect:   image.Rect(0, 0, 1, 1),
	}

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
	return imageShouldLookLikeFile(actualImage, expected...)
}

func ShouldLookLikeFile(actual interface{}, expected ...interface{}) string {
	expected[0] = getTestDataKeyFromArgs(expected)

	switch v := actual.(type) {
	case render.RenderQueueInterface:
		return backBufferShouldLookLike(v, expected...)
	case image.Image:
		_, foundBg := getBackgroundFromArgs(expected)
		if !foundBg {
			// When comparing a given image, we should make sure its transparency
			// matches the expected transparency so we need to use a transparent
			// background when comparing.
			expected = append(expected, BackgroundColour(transparent))
		}
		return imageShouldLookLikeFile(v, expected...)
	default:
		panic(fmt.Errorf("ShouldLookLikeFile needs a *image.RGBA or render.RenderQueueInterface but got %T", actual))
	}
}

func ShouldNotLookLikeFile(actual interface{}, expected ...interface{}) string {
	expected = append(expected, MakeRejectFiles(false))
	doesLook := ShouldLookLikeFile(actual, expected...)
	if doesLook == "" {
		return "arguments matched but should have been different"
	}

	return ""
}

func ShouldLookLikeText(actual interface{}, expected ...interface{}) string {
	testDataKey := getTestDataKeyFromArgs(expected)

	expected[0] = NewTestdataReference("text/" + string(testDataKey))
	return ShouldLookLikeFile(actual, expected...)
}
