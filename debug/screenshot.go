package debug

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/go-gl-legacy/gl"
)

type vflipped struct {
	image.Image
}

func (v vflipped) At(x, y int) color.Color {
	flippedy := v.Bounds().Dy() - y - 1
	return v.Image.At(x, flippedy)
}

func ScreenShot(width, height int, out io.Writer) {
	// 4 bytes per pixel; one byte per RGBA component
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	err := png.Encode(out, vflipped{rgba})
	if err != nil {
		panic(fmt.Errorf("ScreenShot: png.Encode failed: %w", err))
	}
}

func ScreenShotRgba(width, height int) *image.RGBA {
	// 4 bytes per pixel; one byte per RGBA component
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// Flip the rows of pixels vertically so that the leading bytes correspond to
	// the 'top' row of pixels.
	tmp := make([]byte, width*4)
	for rowIdx := 0; rowIdx < height/2; rowIdx++ {
		a, b := rowIdx, height-rowIdx-1
		if a == b {
			break
		}
		arow := rgba.Pix[a*width*4 : (a+1)*width*4]
		brow := rgba.Pix[b*width*4 : (b+1)*width*4]
		copy(tmp, arow)
		copy(arow, brow)
		copy(brow, tmp)
	}

	return rgba
}
