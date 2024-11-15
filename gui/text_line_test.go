package gui_test

import (
	"bytes"
	"fmt"
	"image"
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
	"github.com/spakin/netpbm"
)

// Load a .pam image from the given stream.
func readPam(reader io.Reader) *netpbm.RGBAM {
	img, magic, err := image.Decode(reader)
	if err != nil {
		panic(fmt.Errorf("image.Decode failed: %w", err))
	}

	if magic != "pam" {
		panic(fmt.Errorf("expected .pam file but got %q", magic))
	}

	result, ok := img.(*netpbm.RGBAM)
	if !ok {
		panic(fmt.Errorf("the expected image should have been a image.RGBA image, got %T", img))
	}

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

			pamFileExpected := "../testdata/text/some-text/0.pam"
			pamReader, err := os.Open(pamFileExpected)
			if err != nil {
				panic(fmt.Errorf("couldn't read expectaction file %q, err: %w", pamFileExpected, err))
			}

			expected := readPam(pamReader)

			if bytes.Compare(actualBytes, expected.Pix) != 0 {
				// For debug purposes, copy the bad frame buffer for offline inspection.
				actualImage := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
				actualImage.Pix = actualBytes

				rejectFileName := makeRejectName(pamFileExpected, ".pam")
				rejectFile, err := os.Create(rejectFileName)
				if err != nil {
					panic(fmt.Errorf("couldn't open rejectFileName %q: %w", rejectFileName, err))
				}
				defer rejectFile.Close()

				pamOpts := netpbm.EncodeOptions{
					Format:    netpbm.PAM,
					MaxValue:  255,
					TupleType: "RGB_ALPHA",
					Plain:     false,
				}
				err = netpbm.Encode(rejectFile, actualImage, &pamOpts)
				if err != nil {
					panic(fmt.Errorf("couldn't write rejection file: %s: %w", rejectFileName, err))
				}

				t.Fatalf("YOU SHALL NOT PASS!")
			}
		})
	})
}
