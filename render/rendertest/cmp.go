package rendertest

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render"
)

var transparent = color.RGBA{}

func readPixels(width, height int) ([]byte, error) {
	ret := make([]byte, width*height*4)
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, ret)
	return ret, nil
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
	expectedImage := MustLoadImageFromReader(expected)
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

	expectedImage := MustLoadImageFromReader(pngFile)
	return ImagesAreWithinThreshold(actualImage, expectedImage, thresh, bg)
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

func makeFallbackImage() *image.RGBA {
	r, g, b, a := defaultBackground.RGBA()
	fallbackPixel := []uint8{
		uint8(r), uint8(g), uint8(b), uint8(a),
	}

	return &image.RGBA{
		Pix:    fallbackPixel,
		Stride: 4,
		Rect:   image.Rect(0, 0, 1, 1),
	}
}

func backBufferShouldLookLike(queue render.RenderQueueInterface, expected ...interface{}) (testResult string) {
	defer func() {
		if e := recover(); e != nil {
			testResult = fmt.Sprintf("panic during image comparison: %v", e)
		}
	}()

	// Sometimes, the given queue is a no-op queue so we will have a default
	// 'actualImage' to avoid passing around a nil image.Image value.
	var actualImage *image.RGBA = makeFallbackImage()

	// Read all the pixels from the framebuffer through OpenGL
	var backgroundForImageCmp color.Color = defaultBackground
	queue.Queue(func(render.RenderQueueState) {
		_, _, actualScreenWidth, actualScreenHeight := debug.GetViewport()
		actualImage = debug.ScreenShotRgba(int(actualScreenWidth), int(actualScreenHeight))
		backgroundForImageCmp = GetCurrentBackgroundColor()
	})
	queue.Purge()

	// When screen shotting, we only read opaque pixels; there's always going to
	// be _some_ value for each element of the frame buffer.
	// If the expectation file has transparent portions, we need to compose it
	// over a suitable background before comparing pixel values.
	_, ok := getBackgroundFromArgs(expected)
	if !ok {
		// If nobody specified the background to compose over, we'll use either
		// OpenGL's 'ClearColor' or, if 'queue' is a no-op, whatever the default
		// background is (typically black).
		expected = append(expected, BackgroundColour(backgroundForImageCmp))
	}

	return imageShouldLookLikeFile(actualImage, expected...)
}

// Usage is
//
//	'So(something, ShouldLookLikeFile, "test-case-family", options...)'
//
// 'options' is an optional bag of control parameters
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

// Usage is
//
//	'MustLookLikeFile(t, something, options...)'
//
// 'options' is an optional bag of control parameters
func MustLookLikeFile(t *testing.T, actual any, expected ...any) {
	testresult := ShouldLookLikeFile(actual, expected...)
	if testresult != "" {
		t.Fatalf("image mismatch: %s", testresult)
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
