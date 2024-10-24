package gui

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readPixels(width, height int) ([]byte, error) {
	ret := make([]byte, width*height)
	gl.ReadPixels(0, 0, width, height, gl.RED, gl.UNSIGNED_BYTE, ret)
	return ret, nil
}

func initGlForTest() (system.System, render.RenderQueueInterface, int, int) {
	runtime.LockOSThread()
	linuxSystemObject := gos.GetSystemInterface()
	sys := system.Make(linuxSystemObject)
	wdx := 512
	wdy := 64

	sys.Startup()
	render := render.MakeQueue(func() {
		sys.CreateWindow(0, 0, wdx, wdy)
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

	return sys, render, wdx, wdy
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

func renderStringForTest(toDraw string, sys system.System, render render.RenderQueueInterface, logger *slog.Logger) {
	d := loadDictionaryForTest(render, logger)

	render.Queue(func() {
		d.RenderString(toDraw, 0, 0, 0, d.MaxHeight(), Left)
		sys.SwapBuffers()
	})

	render.Purge()
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

func TestDictionaryRenderString(t *testing.T) {
	// TODO(tmckee): probably need to stop exporting Dictionary from gui and call
	// LoadDictionary to get an instance instead; it'll register shaders and
	// such.
	t.Run("CanRenderLol", func(t *testing.T) {
		sys, render, wdx, wdy := initGlForTest()

		renderStringForTest("lol", sys, render, slog.Default())

		var err error

		// Read all the pixels from the framebuffer through OpenGL
		var frameBufferBytes []byte
		render.Queue(func() {
			frameBufferBytes, err = readPixels(wdx, wdy)
			if err != nil {
				panic(fmt.Errorf("couldn't readPixels: %v", err))
			}
		})
		render.Purge()

		// Verify that the framebuffer's contents match our expected image.
		expectedImage := "../testdata/text/lol.pgm"
		expectedBytes, err := os.ReadFile(expectedImage)
		if err != nil {
			panic(err)
		}

		rejectFileName := "../test/lol.rej.pgm"
		pgmBytes := append([]byte("P5 512 64 255 "), frameBufferBytes...)
		cmp := bytes.Compare(expectedBytes, pgmBytes)
		if cmp != 0 {
			// For debug purposes, copy the bad frame buffer for offline inspection.
			rejectFile, err := os.Create(rejectFileName)
			if err != nil {
				panic(fmt.Errorf("couldn't open rejection file: %s: %v", rejectFileName, err))
			}
			defer rejectFile.Close()

			io.Copy(rejectFile, bytes.NewReader(pgmBytes))

			t.Fatalf("framebuffer mismatch; see %s", rejectFileName)
		}
	})
}

func TestRunTextSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(CleanLogsPerFrameSpec)
	gospec.MainGoTest(r, t)
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

func CleanLogsPerFrameSpec(c gospec.Context) {
	c.Specify("stdout isn't spammed by RenderString", func() {
		sys, render, _, _ := initGlForTest()

		voidLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.Level(-42),
		}))
		stdoutLines := CollectOutput(func() {
			renderStringForTest("lol", sys, render, voidLogger)
		})

		c.Expect(stdoutLines, gospec.ContainsExactly, []string{})
	})
}
