package gui_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render/rendertest"
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

func TestMakeDictionary(t *testing.T) {
	t.Run("CanMakeDict10.gob", func(t *testing.T) {
		discardQueue := rendertest.MakeDiscardingRenderQueue()

		font := mustLoadFont("../testdata/fonts/skia.ttf")
		d := gui.MakeDictionary(font, 10, discardQueue)

		buf := bytes.Buffer{}
		// TODO(tmckee): with the new 'Full_bounds' field, we will never compare
		// equal to the old .gob file. Consider a smarter comparison than
		// bytes.Compare.
		d.Store(&buf)
		actualBytes := buf.Bytes()

		expectedFilename := "../testdata/fonts/dict_10.gob"
		expectedBytes, err := os.ReadFile(expectedFilename)
		if err != nil {
			panic(fmt.Errorf("couldn't os.ReadFile: %w", err))
		}

		rejectFileName := makeRejectName(expectedFilename, ".gob")

		if bytes.Compare(expectedBytes, actualBytes) != 0 {
			// For debug purposes, copy the bad bytes for offline inspection.
			rejectFile, err := os.Create(rejectFileName)
			if err != nil {
				panic(fmt.Errorf("couldn't open rejection file: %s: %v", rejectFileName, err))
			}
			defer rejectFile.Close()

			io.Copy(rejectFile, bytes.NewReader(actualBytes))

			t.Fatalf(".gob file mismatch; see %s", rejectFileName)
		}
	})
}
