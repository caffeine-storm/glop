package gui

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	"github.com/spakin/netpbm"
	_ "github.com/spakin/netpbm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const screenPixelWidth = 512
const screenPixelHeight = 64

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

// Load and convert from .pgm's top-row-first to OpenGL's bottom-row-first.
func readAndFlipPgm(reader io.Reader) *netpbm.GrayM {
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

	return grayImage
}

func readPixels(width, height int) ([]byte, error) {
	ret := make([]byte, width*height)
	gl.ReadPixels(0, 0, width, height, gl.RED, gl.UNSIGNED_BYTE, ret)
	return ret, nil
}

func ShouldLookLike(actual interface{}, expected ...interface{}) string {
	render, ok := actual.(render.RenderQueueInterface)
	if !ok {
		panic(fmt.Errorf("ShouldLookLike needs a render queue but got %T", actual))
	}
	filename, ok := expected[0].(string)
	if !ok {
		panic(fmt.Errorf("ShouldLookLike needs a filename but got %T", expected))
	}

	ok, rejectFile := expectPixelsMatch(render, filename)
	if ok {
		return ""
	}

	return fmt.Sprintf("frame buffer mismatch; see %s", rejectFile)
}

func initGlForTest() (system.System, render.RenderQueueInterface) {
	linuxSystemObject := gos.GetSystemInterface()
	sys := system.Make(linuxSystemObject)

	sys.Startup()
	render := render.MakeQueue(func() {
		sys.CreateWindow(0, 0, screenPixelWidth, screenPixelHeight)
		sys.EnableVSync(true)
		err := gl.Init()
		if err != 0 {
			panic(fmt.Errorf("couldn't gl.Init: %d", err))
		}
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	render.StartProcessing()
	err := Init(render)
	if err != nil {
		panic(fmt.Errorf("couldn't gui.Init(): %w", err))
	}

	return sys, render
}

func loadDictionaryForTest(render render.RenderQueueInterface, logger *slog.Logger) *Dictionary {
	dictReader, err := os.Open("../testdata/fonts/dict_10.gob")
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open: %w", err))
	}

	d, err := LoadDictionary(dictReader, render, logger)
	if err != nil {
		panic(fmt.Errorf("couldn't LoadDictionary: %w", err))
	}

	return d
}

// Renders the given string with pixel units and an origin at the bottom-left.
func renderStringForTest(toDraw string, x, y, height int, sys system.System, render render.RenderQueueInterface, just Justification, logger *slog.Logger) {
	d := loadDictionaryForTest(render, logger)

	// d.RenderString assumes an origin at the top-left so we need to mirror our
	// y co-ordinate.
	y = screenPixelHeight - y

	render.Queue(func() {
		d.RenderString(toDraw, x, y, 0, height, just)
		sys.SwapBuffers()
	})

	render.Purge()
}

// Return the given file but with a '.rej' component to signify a 'rejection'.
func makeRejectName(exp, suffix string) string {
	dir, expectedFileName := path.Split(exp)
	rejectFileNameBase, ok := strings.CutSuffix(expectedFileName, suffix)
	if !ok {
		panic(fmt.Errorf("need a %s file, got %s", suffix, exp))
	}
	return path.Join(dir, rejectFileNameBase+".rej"+suffix)
}

func expectPixelsMatch(render render.RenderQueueInterface, pgmFileExpected string) (bool, string) {
	var err error

	// Read all the pixels from the framebuffer through OpenGL
	var frameBufferBytes []byte
	render.Queue(func() {
		frameBufferBytes, err = readPixels(screenPixelWidth, screenPixelHeight)
		if err != nil {
			panic(fmt.Errorf("couldn't readPixels: %w", err))
		}
	})
	render.Purge()

	// Verify that the framebuffer's contents match our expected image.
	pgmFile, err := os.Open(pgmFileExpected)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", pgmFileExpected, err))
	}
	defer pgmFile.Close()

	expectedImage := readAndFlipPgm(pgmFile)

	cmp := bytes.Compare(expectedImage.Pix, frameBufferBytes)
	if cmp != 0 {
		// For debug purposes, copy the bad frame buffer for offline inspection.
		actualImage := netpbm.NewGrayM(image.Rect(0, 0, screenPixelWidth, screenPixelHeight), 255)
		// Need to flip from bottom-row-first to top-row-first.
		actualImage.Pix = verticalFlipPix(frameBufferBytes, screenPixelWidth, screenPixelHeight)

		rejectFileName := makeRejectName(pgmFileExpected, ".pgm")
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

func TestDictionaryMaxHeight(t *testing.T) {
	t.Run("default-height-is-zero", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{}

		require.Equal(0, d.MaxHeight())
	})
	t.Run("zero-height-at-non-zero-offset", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 42,
				Maxy: 42,
			},
		}

		require.Equal(0, d.MaxHeight())
	})
	t.Run("height-clamped-non-negative", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 42,
				Maxy: 0,
			},
		}

		require.Equal(0, d.MaxHeight())
	})
	t.Run("height-is-delta-min-max", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 0,
				Maxy: 42,
			},
		}

		require.Equal(42, d.MaxHeight())
	})
}

func TestDictionaryGetInfo(t *testing.T) {
	t.Run("AsciiInfoSucceeds", func(t *testing.T) {
		assert := assert.New(t)

		render := rendertest.MakeDiscardingRenderQueue()
		d := loadDictionaryForTest(render, slog.Default())

		emptyRuneInfo := runeInfo{}
		// In ascii, all the characters we care about are between 0x20 (space) and
		// 0x7E (tilde).
		for idx := ' '; idx <= '~'; idx++ {
			info := d.getInfo(rune(idx))
			assert.NotEqual(emptyRuneInfo, info)
		}
	})

	// TODO(tmckee): verify slices of texture by runeInfo correspond to correct
	// letters
	// TODO(tmckee): verify texture image in GL matches expectations
}

func DictionaryRenderStringSpec() {
	sys, render := initGlForTest()

	leftPixel := screenPixelWidth / 2
	bottomPixel := screenPixelHeight / 2
	height := 22
	just := Left
	logger := slog.Default()

	doRenderString := func(toDraw string) {
		renderStringForTest(toDraw, leftPixel, bottomPixel, height, sys, render, just, logger)
	}

	Convey("Can render 'lol'", func() {
		doRenderString("lol")

		So(render, ShouldLookLike, "../testdata/text/lol.pgm")
	})

	Convey("Can render 'credits' centred", func() {
		just = Center
		doRenderString("Credits")

		So(render, ShouldLookLike, "../testdata/text/credits.pgm")
	})

	Convey("Can render somewhere other than the origin", func() {
		Convey("can render at the bottom left", func() {
			leftPixel = 0
			bottomPixel = 0
			logger = glog.DebugLogger()
			doRenderString("offset")

			So(render, ShouldLookLike, "../testdata/text/offset.pgm")
		})
	})

	Convey("Can render to a given height", func() {
		height = 5
		logger = glog.DebugLogger()
		doRenderString("tall-or-small")

		So(render, ShouldLookLike, "../testdata/text/tall-or-small.pgm")
	})

	Convey("stdout isn't spammed by RenderString", func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.Level(-42),
		}))

		stdoutLines := CollectOutput(func() {
			doRenderString("lol")
		})

		So(stdoutLines, ShouldEqual, []string{})
	})
}

func TestRunTextSpecs(t *testing.T) {
	Convey("Dictionaries should render strings", t, DictionaryRenderStringSpec)
}

// Runs the given operation and returns a slice of strings that the operation
// wrote to log.Default().*, slog.Default().*, stdout and stderr combined.
func CollectOutput(operation func()) []string {
	read, write, err := os.Pipe()
	if err != nil {
		panic(fmt.Errorf("couldn't os.Pipe: %w", err))
	}

	go func() {
		stdlogger := log.Default()
		oldLogOut := stdlogger.Writer()
		stdlogger.SetOutput(write)
		defer stdlogger.SetOutput(oldLogOut)

		stdSlogger := slog.Default()
		pipeSlogger := slog.New(slog.NewTextHandler(write, nil))
		slog.SetDefault(pipeSlogger)
		defer slog.SetDefault(stdSlogger)

		oldStdout := os.Stdout
		os.Stdout = write
		defer func() { os.Stdout = oldStdout }()

		oldStderr := os.Stderr
		os.Stderr = write
		defer func() { os.Stderr = oldStderr }()

		// Prefer to defer closing the write end of the pipe. If operation panics,
		// the pipe still needs to be closed or else the reading goroutine would
		// block forever.
		defer write.Close()

		operation()
	}()

	byteList, err := io.ReadAll(read)
	if err != nil {
		panic(fmt.Errorf("couldn't io.ReadAll on the read end of the pipe: %w", err))
	}

	if len(byteList) == 0 {
		return []string{}
	}

	return strings.Split(string(byteList), "\n")
}
