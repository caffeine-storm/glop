package gui_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
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
		d := gui.MakeDictionary(font, 10, discardQueue)

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
