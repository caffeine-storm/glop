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

func ScreenShot(width, height int, out io.Writer) {
	// 4 bytes per pixel; one byte per RGBA component
	nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, nrgba.Pix)

	err := png.Encode(out, imgmanip.VertFlipped{Image: nrgba})
	if err != nil {
		panic(fmt.Errorf("ScreenShot: png.Encode failed: %w", err))
	}
}

func ScreenShotNrgba(width, height int) *image.NRGBA {
	// 4 bytes per pixel; one byte per RGBA component
	nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, nrgba.Pix)

	// Flip the rows of pixels vertically so that the leading bytes correspond to
	// the 'top' row of pixels.
	imgmanip.FlipVertically(nrgba)

	return nrgba
}
