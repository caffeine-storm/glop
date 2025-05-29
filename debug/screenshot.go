package debug

import (
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/imgmanip"
)

// TODO(tmckee): yuck! we ought to be able to do this on the GPU and/or just
// always upload alpha-premultiplied things.
func convertToAlphaPremultiplied(bs []byte) {
	for idx := 0; idx < len(bs); idx += 4 {
		alpha_norm := float64(bs[idx+3]) / 255
		bs[idx+0] = uint8(float64(bs[idx+0]) * alpha_norm)
		bs[idx+1] = uint8(float64(bs[idx+1]) * alpha_norm)
		bs[idx+2] = uint8(float64(bs[idx+2]) * alpha_norm)
	}
}

func ScreenShot(width, height int, out io.Writer) {
	// 4 bytes per pixel; one byte per RGBA component
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// RGBA data needs to be alpha-premultiplied.
	convertToAlphaPremultiplied(rgba.Pix)

	err := png.Encode(out, imgmanip.VertFlipped{Image: rgba})
	if err != nil {
		panic(fmt.Errorf("ScreenShot: png.Encode failed: %w", err))
	}
}

func ScreenShotRgba(width, height int) *image.RGBA {
	// 4 bytes per pixel; one byte per RGBA component
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// RGBA data needs to be alpha-premultiplied.
	convertToAlphaPremultiplied(rgba.Pix)

	// Flip the rows of pixels vertically so that the leading bytes correspond to
	// the 'top' row of pixels.
	imgmanip.FlipVertically(rgba)

	return rgba
}
