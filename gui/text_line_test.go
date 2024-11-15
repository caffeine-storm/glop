package gui_test

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

// Load a .png image from the given stream.
func readPng(reader io.Reader) *image.RGBA {
	img, err := png.Decode(reader)
	if err != nil {
		panic(fmt.Errorf("png.Decode failed: %w", err))
	}

	// Need to redraw the decoded image to get the right byte format out for
	// comparisons.
	result := image.NewRGBA(img.Bounds())
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Src)
	return result
}

func verticalFlipRgbaPix(rgbaPixels []byte, width, height int) []byte {
	if len(rgbaPixels) != width*height*4 {
		panic(fmt.Errorf("invalid dimensions: need %d (%dx%dx4) but got %d", width*height*4, width, height, len(rgbaPixels)))
	}
	result := make([]byte, len(rgbaPixels))

	// Convert from bottom-first-row to top-first-row.
	byteWidth := width * 4
	for rowIndex := 0; rowIndex < height; rowIndex++ {
		inputRowPtr := (height - rowIndex - 1) * byteWidth
		inputRowEnd := inputRowPtr + byteWidth
		resultRowPtr := rowIndex * byteWidth
		resultRowEnd := resultRowPtr + byteWidth

		copy(result[resultRowPtr:resultRowEnd], rgbaPixels[inputRowPtr:inputRowEnd])
	}

	return result
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

func TestTextLine(t *testing.T) {
	screenWidth, screenHeight := 200, 50

	t.Run("Can make a 'lol' line", func(t *testing.T) {
		renderQueue := rendertest.MakeDiscardingRenderQueue()
		dict := gui.LoadDictionaryForTest(renderQueue, &gui.ConstDimser{}, glog.VoidLogger())
		gui.AddDictForTest("glop.font", dict, &render.ShaderBank{})
		textLine := gui.MakeTextLine("glop.font", "lol", 42, 1, 1, 1, 1)
		if textLine == nil {
			t.Fatalf("got a nil TextLine back :(")
		}
	})

	t.Run("TextLine draws its text", func(t *testing.T) {
		rendertest.WithGlForTest(screenWidth, screenHeight, func(sys system.System, queue render.RenderQueueInterface) {
			// TODO(tmckee): XXX: having to remember to gui.Init is ... sad-making
			gui.Init(queue)
			dimser := &gui.ConstDimser{Value: gui.Dims{screenWidth, screenHeight}}
			dict := gui.LoadDictionaryForTest(queue, dimser, glog.DebugLogger())

			var shaderBank *render.ShaderBank
			queue.Queue(func(rqs render.RenderQueueState) {
				shaderBank = rqs.Shaders()
			})
			queue.Purge()

			gui.AddDictForTest("glop.font", dict, shaderBank)

			textLine := gui.MakeTextLine("glop.font", "some text", 32, 1, 1, 1, 1)
			var actualBytes []byte

			queue.Queue(func(render.RenderQueueState) {
				textLine.Draw(gui.Region{
					Point: gui.Point{},
					Dims:  gui.Dims{screenWidth, screenHeight},
				})

				buffer := &bytes.Buffer{}
				debug.ScreenShotRgba(screenWidth, screenHeight, buffer)
				actualBytes = buffer.Bytes()
			})
			queue.Purge()

			pngFileExpected := "../testdata/text/some-text/0.png"
			pngReader, err := os.Open(pngFileExpected)
			if err != nil {
				panic(fmt.Errorf("couldn't read expectaction file %q, err: %w", pngFileExpected, err))
			}

			expected := readPng(pngReader)

			if bytes.Compare(actualBytes, expected.Pix) != 0 {
				// For debug purposes, copy the bad frame buffer for offline inspection.
				actualImage := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
				actualImage.Pix = actualBytes

				rejectFileName := makeRejectName(pngFileExpected, ".png")
				rejectFile, err := os.Create(rejectFileName)
				if err != nil {
					panic(fmt.Errorf("couldn't open rejectFileName %q: %w", rejectFileName, err))
				}
				defer rejectFile.Close()

				err = png.Encode(rejectFile, actualImage)
				if err != nil {
					panic(fmt.Errorf("couldn't write rejection file: %s: %w", rejectFileName, err))
				}

				t.Fatalf("framebuffer mismatch; see %q", rejectFileName)
			}
		})
	})
}
