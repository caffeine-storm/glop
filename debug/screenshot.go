package debug

import (
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/imgmanip"
)

var logger = glog.WarningLogger()

func checkForAlphaPremultiplied(bs []byte) {
	for idx := 0; idx < len(bs); idx += 4 {
		if max(bs[idx+0], bs[idx+1], bs[idx+2]) > bs[idx+3] {
			logger.Warn("found non-normalized colour", "idx", idx)
		}
	}
}

func ScreenShot(width, height int, out io.Writer) {
	// 4 bytes per pixel; one byte per RGBA component
	rgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// Pixel data needs to be non-alpha-premultiplied.
	checkForAlphaPremultiplied(rgba.Pix)

	err := png.Encode(out, imgmanip.VertFlipped{Image: rgba})
	if err != nil {
		panic(fmt.Errorf("ScreenShot: png.Encode failed: %w", err))
	}
}

func ScreenShotNrgba(width, height int) *image.NRGBA {
	// 4 bytes per pixel; one byte per RGBA component
	rgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// Pixel data needs to be non-alpha-premultiplied.
	checkForAlphaPremultiplied(rgba.Pix)

	// Flip the rows of pixels vertically so that the leading bytes correspond to
	// the 'top' row of pixels.
	imgmanip.FlipVertically(rgba)

	return rgba
}
