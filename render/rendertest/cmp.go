package rendertest

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"strings"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/spakin/netpbm"
	_ "github.com/spakin/netpbm"
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
	ret := make([]byte, width*height)
	gl.ReadPixels(0, 0, width, height, gl.RED, gl.UNSIGNED_BYTE, ret)
	return ret, nil
}

// Load and convert from .pgm's top-row-first to OpenGL's bottom-row-first.
func readAndFlipPgm(reader io.Reader) (*netpbm.GrayM, int, int) {
	img, magic, err := image.Decode(reader)
	if err != nil {
		panic(fmt.Errorf("image.Decode failed: %w", err))
	}

	if magic != "pgm" {
		panic(fmt.Errorf("expected .pgm file but got %q", magic))
	}

	grayImage, ok := img.(*netpbm.GrayM)
	if !ok {
		panic(fmt.Errorf("the expected image should have been a netpbm.GrayM image"))
	}

	width := grayImage.Bounds().Dx()
	height := grayImage.Bounds().Dy()
	grayImage.Pix = verticalFlipPix(grayImage.Pix, width, height)

	return grayImage, width, height
}

func verticalFlipPix(pixels []byte, width, height int) []byte {
	result := make([]byte, len(pixels))

	// Convert from bottom-first-row to top-first-row.
	for row := 0; row < height; row++ {
		resultRowIdx := row * width
		resultRowEnd := resultRowIdx + width
		inputRowIdx := (height - row - 1) * width
		inputRowEnd := inputRowIdx + width

		copy(result[resultRowIdx:resultRowEnd], pixels[inputRowIdx:inputRowEnd])
	}

	return result
}

// Verify that the framebuffer's contents match our expected image.
func expectPixelsMatch(queue render.RenderQueueInterface, pgmFileExpected string) (bool, string) {
	pgmFile, err := os.Open(pgmFileExpected)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", pgmFileExpected, err))
	}
	defer pgmFile.Close()

	expectedImage, screenWidth, screenHeight := readAndFlipPgm(pgmFile)
	// Read all the pixels from the framebuffer through OpenGL
	var frameBufferBytes []byte
	queue.Queue(func(render.RenderQueueState) {
		frameBufferBytes, err = readPixels(screenWidth, screenHeight)
		if err != nil {
			panic(fmt.Errorf("couldn't readPixels: %w", err))
		}
	})
	queue.Purge()

	cmp := bytes.Compare(expectedImage.Pix, frameBufferBytes)
	if cmp != 0 {
		// For debug purposes, copy the bad frame buffer for offline inspection.
		actualImage := netpbm.NewGrayM(image.Rect(0, 0, screenWidth, screenHeight), 255)
		// Need to flip from bottom-row-first to top-row-first.
		actualImage.Pix = verticalFlipPix(frameBufferBytes, screenWidth, screenHeight)

		rejectFileName := MakeRejectName(pgmFileExpected, ".pgm")
		rejectFile, err := os.Create(rejectFileName)
		if err != nil {
			panic(fmt.Errorf("couldn't open rejectFileName %q: %w", rejectFileName, err))
		}
		defer rejectFile.Close()

		pgmOpts := netpbm.EncodeOptions{
			Format:   netpbm.PGM,
			MaxValue: 255,
			Plain:    false,
		}
		err = netpbm.Encode(rejectFile, actualImage, &pgmOpts)
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

	filename := ExpectationFile(testDataKey, "pgm", testnumber)

	ok, rejectFile := expectPixelsMatch(render, filename)
	if ok {
		return ""
	}

	return fmt.Sprintf("frame buffer mismatch; see %s", rejectFile)
}
