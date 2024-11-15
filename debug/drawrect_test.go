package debug_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	"github.com/spakin/netpbm"
)

// TODO(tmckee): clean: we can probably just use a color.RGBA
type pixel struct {
	r, g, b, a byte
}

type bounded struct {
	*image.Uniform
	bounds image.Rectangle
}

func (b *bounded) Bounds() image.Rectangle {
	return b.bounds
}

func boundedUniform(bounds image.Rectangle, colour color.Color) image.Image {
	return &bounded{
		Uniform: image.NewUniform(colour),
		bounds:  bounds,
	}
}

// Write the expectation file lazily; it's not in source control b/c it's
// (somewhat?) easily generated on demand.
func writeExpectationFile(fileKey string, width, height int, expectedPixel pixel) {
	expectedFilename := fmt.Sprintf("testdata/%s.pam", fileKey)
	out, err := os.Create(expectedFilename)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Create: %w", err))
	}
	defer out.Close()

	expectedColour := color.RGBA{
		R: expectedPixel.r,
		G: expectedPixel.g,
		B: expectedPixel.b,
		A: expectedPixel.a,
	}
	expectedImage := boundedUniform(image.Rect(0, 0, width, height), expectedColour)

	pamOpts := netpbm.EncodeOptions{
		Format:    netpbm.PAM,
		MaxValue:  255,
		TupleType: "RGB_ALPHA",
		Plain:     false,
	}
	err = netpbm.Encode(out, expectedImage, &pamOpts)
	if err != nil {
		panic(fmt.Errorf("couldn't netpbm.Encode: %w", err))
	}
}

func writeRejectionFile(fileKey string, width, height int, data []byte) {
	rejectionFile := fmt.Sprintf("testdata/%s.rej.pam", fileKey)
	out, err := os.Create(rejectionFile)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Create: %w", err))
	}
	defer out.Close()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Pix = data
	pamOpts := netpbm.EncodeOptions{
		Format:    netpbm.PAM,
		MaxValue:  255,
		TupleType: "RGB_ALPHA",
		Plain:     false,
	}
	err = netpbm.Encode(out, img, &pamOpts)
	if err != nil {
		panic(fmt.Errorf("couldn't netpbm.Encode: %w", err))
	}
}

func writeFailureArtifacts(width, height int, expected pixel, rgbaBytes []byte) {
	writeExpectationFile("test-draw-rect", width, height, expected)
	writeRejectionFile("test-draw-rect", width, height, rgbaBytes)
}

func TestDrawRect(t *testing.T) {
	width, height := 50, 50
	buffer := &bytes.Buffer{}

	rendertest.WithGlForTest(width, height, func(sys system.System, queue render.RenderQueueInterface) {
		queue.Queue(func(render.RenderQueueState) {
			debug.BlankAndDrawRectNdc(-1, -1, 1, 1)
			debug.ScreenShotRgba(width, height, buffer)
		})
		queue.Purge()
	})

	rgbaBytes := buffer.Bytes()
	if len(rgbaBytes) != width*height*4 {
		panic(fmt.Errorf("wrong number of bytes, expected %d got %d", width*height*4, len(rgbaBytes)))
	}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			// Each pixel is 4 bytes of {r, g, b, a} starting at x*4 + y*50*4
			// The whole screen should be {255, 0, 0, 255}
			idx := x*4 + y*50*4
			r, g, b, a := rgbaBytes[idx], rgbaBytes[idx+1], rgbaBytes[idx+2], rgbaBytes[idx+3]
			px := pixel{
				r: r,
				g: g,
				b: b,
				a: a,
			}
			expected := pixel{
				r: 255,
				g: 0,
				b: 0,
				a: 255,
			}
			if px != expected {
				writeFailureArtifacts(width, height, expected, rgbaBytes)
				t.Fatalf("pixel mismatch at (%d, %d): %+v", x, y, px)
			}
		}
	}
}

/* DANGER WILL ROBINSON! XXX: this has been crashing windows when running on
* WSL. Run at your PERIL!
* func TestDrawManyRects(t *testing.T) {
* for i := 0; i < 500; i++ {
*   TestDrawRect(t)
* }
}
*/
