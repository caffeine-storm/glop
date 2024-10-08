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
		require := require.New(t)
		assert := assert.New(t)

		sys := system.Make(gos.GetSystemInterface())
		wdx := 1024
		wdy := 750

		sys.Startup()
		render.Init()
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

		d, err := LoadDictionary(dictReader)
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
		sys := system.Make(gos.GetSystemInterface())
		// wdx := 640
		// wdy := 512
		wdx := 512
		wdy := 64

		sys.Startup()
		render.Init()
		render.Queue(func() {
			sys.CreateWindow(0, 0, wdx, wdy)
			sys.EnableVSync(true)
			err := gl.Init()
			if err != 0 {
				panic(err)
			}
			gl.Enable(gl.BLEND)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		})

		require := require.New(t)
		dictReader, err := os.Open("../testdata/fonts/dict_10.gob")
		require.Nil(err)

		d, err := LoadDictionary(dictReader)
		require.Nil(err)

		render.Queue(func() {
			d.RenderString("lol", 0, 0, 0, d.MaxHeight(), Left)
			sys.SwapBuffers()
		})
		render.Purge()

		// Read a frame-buffer file to peek at Xvfb's data
		frameBufferBytes, err := os.ReadFile("../test/Xvfb_screen0")
		if err != nil {
			panic(err)
		}

		// Verify that we read the right "shape" of file.
		if len(frameBufferBytes) != 85152 {
			fmt.Printf("frameBufferSize: %v\n", fmt.Errorf("The framebuffer file was %d bytes but expected %d", len(frameBufferBytes), 85152))
		}

		// Verify that the framebuffer's contents match our expected image.
		expectedImage := "../testdata/text/lol.xwd"
		expectedBytes, err := os.ReadFile(expectedImage)
		if err != nil {
			panic(err)
		}

		rejectFileName := "../test/lol.rej.xwd"
		cmp := bytes.Compare(expectedBytes, frameBufferBytes)
		if cmp != 0 {
			// For debug purposes, copy the bad frame buffer for offline inspection.
			rejectFile, err := os.Create(rejectFileName)
			if err != nil {
				panic(fmt.Errorf("couldn't open rejection file: %s: %v", rejectFileName, err))
			}
			defer rejectFile.Close()

			io.Copy(rejectFile, bytes.NewReader(frameBufferBytes))

			t.Fatalf("framebuffer mismatch; see %s", rejectFileName)
		}
	})
}
