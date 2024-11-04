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
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render/rendertest"
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

func TestDictionarySerialization(t *testing.T) {
	t.Run("Dictionary.Data must round-trip through encoding/decoding", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		discardQueue := rendertest.MakeDiscardingRenderQueue()

		font := mustLoadFont("../testdata/fonts/skia.ttf")
		d := gui.MakeDictionary(font, 10, discardQueue, &gui.ConstDimser{}, glog.VoidLogger())

		buf := bytes.Buffer{}
		d.Store(&buf)

		// Load it back, compare to 'd_prime.Data' to 'd.Data', make sure they
		// match.
		d_prime := gui.Dictionary{}
		err := d_prime.Load(bytes.NewReader(buf.Bytes()))
		require.Nil(err)

		expectedData := d.Data
		reloadedData := d_prime.Data

		assert.Equal(expectedData.Dx, reloadedData.Dx)
		assert.Equal(expectedData.Dy, reloadedData.Dy)

		assert.Equal(expectedData.Baseline, reloadedData.Baseline)
		assert.Equal(expectedData.Scale, reloadedData.Scale)

		assert.Equal(expectedData.Miny, reloadedData.Miny)
		assert.Equal(expectedData.Maxy, reloadedData.Maxy)

		assert.Equal(expectedData.Kerning, reloadedData.Kerning)
		assert.Equal(expectedData.Info, reloadedData.Info)
		assert.Equal(expectedData.Ascii_info, reloadedData.Ascii_info)

		assert.Exactly(expectedData.Pix, reloadedData.Pix)
	})
}

func TestMakeDictionary(t *testing.T) {
	t.Run("minimalSubImage respects the origin", func(t *testing.T) {
		assert := assert.New(t)

		// Draw something floating above baseline and to the right of (0,0).
		img := image.NewRGBA(image.Rect(-10, -10, 10, 10))
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

	t.Run("MakeDictionary takes a logger", func(t *testing.T) {
		logger := glog.VoidLogger()
		font := mustLoadFont("../testdata/fonts/skia.ttf")

		_ = gui.MakeDictionary(font, 42, rendertest.MakeDiscardingRenderQueue(), &gui.ConstDimser{}, logger)
	})
}

func TestDictionaryRenderString(t *testing.T) {
	t.Run("has a reasonable API", func(t *testing.T) {
		d := gui.LoadDictionaryForTest(rendertest.MakeDiscardingRenderQueue(), &gui.ConstDimser{}, glog.DebugLogger())

		d.RenderString("render this", gui.Point{X: 12, Y: 2}, 14, gui.Left)
	})
}
