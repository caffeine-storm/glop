package gui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readPixels(width, height int) ([]byte, error) {
	ret := make([]byte, width*height)
	gl.ReadPixels(0, 0, width, height, gl.RED, gl.UNSIGNED_BYTE, ret)
	return ret, nil
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

type discardQueue struct {}

func (*discardQueue) Queue(f func()) {}
func (*discardQueue) Purge() {}
func (*discardQueue) StartProcessing() {}

func makeDiscardingRenderQueue() render.RenderQueueInterface {
	return &discardQueue{}
}

func TestDictionaryGetInfo(t *testing.T) {
	t.Run("AsciiInfoSucceeds", func(t *testing.T) {
		require := require.New(t)
		assert := assert.New(t)

		linuxSystemObject := gos.GetSystemInterface()
		sys := system.Make(linuxSystemObject)
		wdx := 1024
		wdy := 750

		sys.Startup()
		render := makeDiscardingRenderQueue()
		render.StartProcessing()
		render.Queue(func() {
			sys.CreateWindow(10, 10, wdx, wdy)
			sys.EnableVSync(true)
			err := gl.Init()
			if err != 0 {
				panic("couldn't init GL")
			}
		})
		render.Purge()

		dictReader, err := os.Open("../testdata/fonts/dict_10.gob")
		require.Nil(err)

		d, err := LoadDictionary(dictReader, render)
		require.Nil(err)

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
		runtime.LockOSThread()
		linuxSystemObject := gos.GetSystemInterface()
		sys := system.Make(linuxSystemObject)
		// wdx := 640
		// wdy := 512
		wdx := 512
		wdy := 64

		sys.Startup()
		render := render.MakeQueue(func() {
			// TODO(tmckee): DRY out creating a window
			sys.CreateWindow(0, 0, wdx, wdy)
			sys.EnableVSync(true)
			err := gl.Init()
			if err != 0 {
				panic(err)
			}
			gl.Enable(gl.BLEND)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		})
		render.StartProcessing()

		require := require.New(t)
		dictReader, err := os.Open("../testdata/fonts/dict_10.gob")
		require.Nil(err)

		d, err := LoadDictionary(dictReader, render)
		require.Nil(err)

		// TODO(tmckee): make a 'test-render' that SetGlContext's for you.
		render.Queue(func() {
			// TODO(tmckee): with initializtion bound to render queue construction,
			// this call should no longer be necessary.
			linuxSystemObject.SetGlContext()
			d.RenderString("lol", 0, 0, 0, d.MaxHeight(), Left)
			sys.SwapBuffers()
		})
		render.Purge()

		// Read all the pixels from the framebuffer through OpenGL
		var frameBufferBytes []byte
		render.Queue(func() {
			// TODO(tmckee): with initializtion bound to render queue construction,
			// this call should no longer be necessary.
			linuxSystemObject.SetGlContext()

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
