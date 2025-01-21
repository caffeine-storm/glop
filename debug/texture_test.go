package debug_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render/rendertest"
)

const maxuint16 = 0xffff

func foreachPixel(img image.Image, check func(x, y int, col color.Color)) {
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			check(x, y, img.At(x, y))
		}
	}
}

func givenATexture(imageFilePath string) gl.Texture {
	imageFilePath = path.Join("testdata", imageFilePath)
	file, err := os.Open(imageFilePath)
	if err != nil {
		panic(fmt.Errorf("couldn't open file %q: %w", imageFilePath, err))
	}

	img, _, err := image.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't image.Decode: %w", err))
	}

	rgbaImage := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImage, img.Bounds(), img, image.Point{}, draw.Src)

	return uploadTextureFromImage(rgbaImage)
}

func uploadTextureFromImage(img *image.RGBA) gl.Texture {
	bounds := img.Bounds()
	gl.Enable(gl.TEXTURE_2D)
	texture := gl.GenTexture()
	texture.Bind(gl.TEXTURE_2D)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	gl.ActiveTexture(gl.TEXTURE0 + 0)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		bounds.Dx(),
		bounds.Dy(),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		img.Pix,
	)

	gl.Disable(gl.TEXTURE_2D)

	return texture
}

func isBlack(c color.Color) bool {
	r, g, b, a := c.RGBA()
	if r != 0 {
		return false
	}
	if g != 0 {
		return false
	}
	if b != 0 {
		return false
	}

	return a == maxuint16
}

func isTransparent(c color.Color) bool {
	_, _, _, a := c.RGBA()
	return a == 0
}
func isRed(c color.Color) bool {
	r, g, b, a := c.RGBA()
	if r != maxuint16 {
		return false
	}
	if g != 0 {
		return false
	}
	if b != 0 {
		return false
	}

	return a == maxuint16
}

func TestTextureDebugging(t *testing.T) {
	t.Run("can dump texture to image object", func(t *testing.T) {
		var dumpedImage *image.RGBA
		var err error

		rendertest.WithGl(func() {
			tex := givenATexture("red/0.png")

			dumpedImage, err = debug.DumpTexture(tex)
			if err != nil {
				t.Fatalf("dumping failed: %v", err)
			}
		})

		// The dump should have produced a 50x50 pixel image of all red.
		foreachPixel(dumpedImage, func(x, y int, col color.Color) {
			if !isRed(col) {
				t.Log("non-red pixel", "x", x, "y", y, "colour", col)
				t.Fail()
			}
		})
	})

	t.Run("can dump to writer", func(t *testing.T) {
		pngBuffer := &bytes.Buffer{}

		rendertest.WithGl(func() {
			tex := givenATexture("red/0.png")

			err := debug.DumpTextureAsPng(tex, pngBuffer)
			if err != nil {
				t.Fatalf("dumping failed: %v", err)
			}
		})

		dumpedImage, err := png.Decode(pngBuffer)
		if err != nil {
			panic(fmt.Errorf("couldn't decode pngBuffer: %w", err))
		}

		// The dump should have produced a 50x50 pixel image of all red.
		foreachPixel(dumpedImage, func(x, y int, col color.Color) {
			if !isRed(col) {
				t.Log("non-red pixel", "x", x, "y", y, "colour", col)
				t.Fail()
			}
		})
	})

	t.Run("can dump non-uniform texture", func(t *testing.T) {
		pngBuffer := &bytes.Buffer{}

		rendertest.WithGl(func() {
			tex := givenATexture("checker/0.png")

			err := debug.DumpTextureAsPng(tex, pngBuffer)
			if err != nil {
				t.Fatalf("dumping failed: %v", err)
			}
		})

		dumpedImage, err := png.Decode(pngBuffer)
		if err != nil {
			panic(fmt.Errorf("couldn't decode pngBuffer: %w", err))
		}

		// The dump should have produced a 64x64 pixel image of a cycle of squares
		// (each 4x4 pixels) that are black, transparent then red.
		foreachPixel(dumpedImage, func(x, y int, col color.Color) {
			idx := (y/4)*16 + (x/4)*1
			switch idx % 3 {
			case 0:
				if !isBlack(col) {
					t.Log("non-black pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			case 1:
				if !isTransparent(col) {
					t.Log("non-transparent pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			case 2:
				if !isRed(col) {
					t.Log("non-red pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			}
		})
	})
}
