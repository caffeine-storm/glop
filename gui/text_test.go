package gui

import (
	"os"
	"runtime"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
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
			data: dictData{
				Miny: 42.0,
				Maxy: 42.0,
			},
		}

		require.Equal(0.0, d.MaxHeight())
	})
	t.Run("height-clamped-non-negative", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			data: dictData{
				Miny: 42.0,
				Maxy: 0.0,
			},
		}

		require.Equal(0.0, d.MaxHeight())
	})
	t.Run("height-is-delta-min-max", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			data: dictData{
				Miny: 0.0,
				Maxy: 42.0,
			},
		}

		require.InDelta(42.0, d.MaxHeight(), 0.001)
	})
}

func TestRenderString(t *testing.T) {
	// TODO(tmckee): probably need to stop exporting Dictionary from gui and call
	// LoadDictionary to get an instance instead; it'll register shaders and
	// such.
	t.Run("rendering-should-not-panic", func(t *testing.T) {
		runtime.LockOSThread()
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
			d.RenderString("lol", 0, 0, 0, 12.0, Left)
			sys.SwapBuffers()
		})
		render.Purge()
	})
}
