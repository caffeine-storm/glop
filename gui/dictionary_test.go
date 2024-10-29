package gui_test

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Return the given file but with a '.rej' component to signify a 'rejection'.
// TODO(tmckee): this is copy-pasta; DRY out making 'rejection' files.
func makeRejectName(exp, suffix string) string {
	dir, expectedFileName := path.Split(exp)
	rejectFileNameBase, ok := strings.CutSuffix(expectedFileName, suffix)
	if !ok {
		panic(fmt.Errorf("need a %s file, got %s", suffix, exp))
	}
	return path.Join(dir, rejectFileNameBase+".rej"+suffix)
}

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

		voidLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.Level(-42),
		}))

		// Load it back, compare to 'd_prime.Data' to 'd.Data', make sure they
		// match.
		d_prime, err := gui.LoadDictionary(bytes.NewReader(buf.Bytes()), discardQueue, voidLogger)
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
