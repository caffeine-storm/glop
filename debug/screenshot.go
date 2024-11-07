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

func ScreenShotRgba(width, height int, out io.Writer) {
	// 4 bytes per pixel; one byte per RGBA component
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// Flip the rows of pixels vertically so that the leading bytes correspond to
	// the 'top' row of pixels.
	for rowIdx := height - 1; rowIdx >= 0; rowIdx-- {
		row := rgba.Pix[rowIdx*width*4 : (rowIdx+1)*width*4]
		n, err := out.Write(row)
		if err != nil {
			panic(fmt.Errorf("out.Write failed: %w", err))
		}
		if n != len(row) {
			panic(fmt.Errorf("only wrote %d elements, not %d", n, len(row)))
		}
	}
}
