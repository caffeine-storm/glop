package gui_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustLoadFont(path string) *truetype.Font {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("couldn't read file %q: %w", path, err))
	}

	font, err := freetype.ParseFont(data)
	if err != nil {
		panic(fmt.Errorf("coudln't ParseFont: %w", err))
	}

	return font
}

func givenAFont() *truetype.Font {
	return mustLoadFont("testdata/fonts/skia.ttf")
}

func TestDictionarySerialization(t *testing.T) {
	t.Run("RasteredFont must round-trip through encoding/decoding", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		font := givenAFont()
		rastered := gui.RasterizeFont(font, 10)

		buf := bytes.Buffer{}
		d := gui.Dictionary{
			Data: rastered,
		}
		d.Store(&buf)

		// Load it back, compare 'd_prime.Data' to 'd.Data', make sure they match.
		d_prime := gui.Dictionary{}
		err := d_prime.Load(bytes.NewReader(buf.Bytes()))
		require.Nil(err)

		expectedData := d.Data
		reloadedData := d_prime.Data

		assert.Equal(expectedData.Dx, reloadedData.Dx)
		assert.Equal(expectedData.Dy, reloadedData.Dy)

		assert.Equal(expectedData.Baseline, reloadedData.Baseline)

		assert.Equal(expectedData.Miny, reloadedData.Miny)
		assert.Equal(expectedData.Maxy, reloadedData.Maxy)

		assert.Equal(expectedData.Kerning, reloadedData.Kerning)
		assert.Equal(expectedData.Info, reloadedData.Info)

		assert.Exactly(expectedData.Pix, reloadedData.Pix)
	})
}

func TestMakeAndInitializeDictionary(t *testing.T) {
	t.Run("minimalSubImage respects the origin", func(t *testing.T) {
		assert := assert.New(t)

		// Draw something floating above baseline and to the right of (0,0).
		img := image.NewNRGBA(image.Rect(-10, -10, 10, 10))
		allTheGray := color.Gray{Y: 255}
		img.Set(3, 4, allTheGray)
		img.Set(4, 5, allTheGray)
		img.Set(5, 4, allTheGray)
		img.Set(6, 6, allTheGray)

		sub := gui.MinimalSubImage(img)

		// Assert that the returned rectangles indicate the correct
		// padding/distance from the origin.
		assert.Equal(image.Point{2, 3}, sub.Bounds().Min)
		assert.Equal(image.Point{7, 7}, sub.Bounds().Max)
	})

	t.Run("fonts are correctly interpreted", func(t *testing.T) {
		font := givenAFont()

		rasterized := gui.RasterizeFont(font, 18)

		t.Run("grid of glyphs constructed", func(t *testing.T) {
			assert.NotEmpty(t, rasterized.Pix)
		})

		t.Run("rune info for 'M' should make sense", func(t *testing.T) {
			minfo, ok := rasterized.Info['M']
			require.True(t, ok)

			// The 'Advance' for an M ought to be about as much as the glyph is wide.
			// We'll check that the advance isn't more than double the width.
			assert.Less(t, minfo.Advance, float64(minfo.Bounds.Dx()*2))
		})
	})
}

func TestDictionaryRenderString(t *testing.T) {
	t.Run("has a reasonable API", func(t *testing.T) {
		testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
			d := gui.LoadAndInitializeDictionaryForTest(queue, glog.DebugLogger())

			// TODO(tmckee): clean: expose 'glop.font' name through the API instead of
			// copy-pasta.
			d.RenderString("render this", gui.Point{X: 12, Y: 2}, 14, gui.Left, rendertest.StubShaderBank("glop.font"))
		})
	})

	t.Run("works regardless of whether gl.TEXTURE_2D is enabled or not", func(t *testing.T) {
		testbuilder.New().WithSize(128, 32).WithQueue().Run(func(queue render.RenderQueueInterface) {

			d := gui.LoadAndInitializeDictionaryForTest(queue, glog.DebugLogger())

			queue.Queue(func(st render.RenderQueueState) {
				gl.Enable(gl.TEXTURE_2D)
				d.RenderString("kekw    ", gui.Point{}, d.MaxHeight(), gui.Left, st.Shaders())
				gl.Disable(gl.TEXTURE_2D)
				d.RenderString("    rofl", gui.Point{}, d.MaxHeight(), gui.Left, st.Shaders())
			})
			queue.Purge()

			rendertest.MustLookLikeFile(t, queue, "laughing")
		})
	})
}

func TestGetFontMetrics(t *testing.T) {
	t.Run("can get font metrics", func(t *testing.T) {
		d := gui.LoadDictionaryForTest()
		assert.NotEqual(t, 0, d.MaxHeight())
	})
}
