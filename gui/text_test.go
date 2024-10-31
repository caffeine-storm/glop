package gui

import (
	"bytes"
	"fmt"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const screenPixelWidth = 512
const screenPixelHeight = 64

func readPixels(width, height int) ([]byte, error) {
	ret := make([]byte, width*height)
	gl.ReadPixels(0, 0, width, height, gl.RED, gl.UNSIGNED_BYTE, ret)
	return ret, nil
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

func renderStringForTest(toDraw string, sys system.System, render render.RenderQueueInterface, just Justification, logger *slog.Logger) {
	renderStringAtOffsetForTest(toDraw, 0, 0, sys, render, just, logger)
}

func renderStringAtOffsetForTest(toDraw string, xpixels, ypixels int, sys system.System, render render.RenderQueueInterface, just Justification, logger *slog.Logger) {
	d := loadDictionaryForTest(render, logger)

	xndc := float64(xpixels) / (float64(screenPixelWidth) / 2)
	yndc := float64(ypixels) / (float64(screenPixelHeight) / 2)

	render.Queue(func() {
		d.RenderString(toDraw, xndc, yndc, 0, d.MaxHeight(), just)
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

func expectPixelsMatch(render render.RenderQueueInterface, pgmFileExpected string) {
	var err error

	// Read all the pixels from the framebuffer through OpenGL
	var frameBufferBytes []byte
	render.Queue(func() {
		frameBufferBytes, err = readPixels(screenPixelWidth, screenPixelHeight)
		if err != nil {
			panic(fmt.Errorf("couldn't readPixels: %v", err))
		}
	})
	render.Purge()

	// Verify that the framebuffer's contents match our expected image.
	expectedBytes, err := os.ReadFile(pgmFileExpected)
	if err != nil {
		panic(err)
	}

	rejectFileName := makeRejectName(pgmFileExpected, ".pgm")

	magicHeader := fmt.Sprintf("P5 %d %d 255 ", screenPixelWidth, screenPixelHeight)
	pgmBytes := append([]byte(magicHeader), frameBufferBytes...)
	cmp := bytes.Compare(expectedBytes, pgmBytes)
	if cmp != 0 {
		// For debug purposes, copy the bad frame buffer for offline inspection.
		rejectFile, err := os.Create(rejectFileName)
		if err != nil {
			panic(fmt.Errorf("couldn't open rejection file: %s: %v", rejectFileName, err))
		}
		defer rejectFile.Close()

		io.Copy(rejectFile, bytes.NewReader(pgmBytes))
	}

	SoMsg("pixels should match", cmp, ShouldEqual, 0)
}

func TestDictionaryMaxHeight(t *testing.T) {
	t.Run("default-height-is-zero", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{}

		require.Equal(0.0, d.MaxHeight())
	})
	t.Run("zero-height-at-non-zero-offset", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 42.0,
				Maxy: 42.0,
			},
		}

		require.Equal(0.0, d.MaxHeight())
	})
	t.Run("height-clamped-non-negative", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 42.0,
				Maxy: 0.0,
			},
		}

		require.Equal(0.0, d.MaxHeight())
	})
	t.Run("height-is-delta-min-max", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 0.0,
				Maxy: 42.0,
			},
		}

		require.InDelta(42.0, d.MaxHeight(), 0.001)
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

	Convey("Can render 'lol'", func() {
		renderStringForTest("lol", sys, render, Left, slog.Default())

		expectPixelsMatch(render, "../testdata/text/lol.pgm")
	})

	Convey("Can render 'credits' centred", func() {
		renderStringForTest("Credits", sys, render, Center, glog.DebugLogger())

		expectPixelsMatch(render, "../testdata/text/credits.pgm")
	})
}

func TestRunTextSpecs(t *testing.T) {
	Convey("Text Specifications", t, func() {
		Convey("Dictionaries should render strings", DictionaryRenderStringSpec)
		Convey("CleanLogsPerFrameSpec", CleanLogsPerFrameSpec)
	})
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
		panic(fmt.Errorf("couldn't os.ReadFile on the read end of the pipe: %w", err))
	}

	if len(byteList) == 0 {
		return []string{}
	}

	return strings.Split(string(byteList), "\n")
}

func CleanLogsPerFrameSpec() {
	Convey("stdout isn't spammed by RenderString", func() {
		sys, render := initGlForTest()

		voidLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.Level(-42),
		}))
		stdoutLines := CollectOutput(func() {
			renderStringForTest("lol", sys, render, Left, voidLogger)
		})

		So(stdoutLines, ShouldEqual, []string{})
	})
}
