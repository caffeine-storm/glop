package debug_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
)

const maxuint16 = 0xffff

func foreachPixel(img image.Image, check func(x, y int, col color.Color)) {
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			check(x, y, img.At(x, y))
		}
	}
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

func isBlue(c color.Color) bool {
	r, g, b, a := c.RGBA()
	if r != 0 {
		return false
	}
	if g != 0 {
		return false
	}
	if b != maxuint16 {
		return false
	}

	return a == maxuint16
}

func blitOntoBlue(img image.Image) *image.NRGBA {
	blue := image.NewUniform(color.RGBA{
		R: 0,
		G: 0,
		B: 255,
		A: 255,
	})
	result := image.NewNRGBA(img.Bounds())

	draw.Draw(result, img.Bounds(), blue, image.Point{}, draw.Src)
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Over)

	return result
}

func TestTextureDebugging(t *testing.T) {
	t.Run("can dump texture to image object", func(t *testing.T) {
		var dumpedImage *image.NRGBA
		var err error

		testbuilder.Run(func() {
			tex, cleanup := rendertest.GivenATexture("red/0.png")
			defer cleanup()

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

		testbuilder.Run(func() {
			tex, cleanup := rendertest.GivenATexture("red/0.png")
			defer cleanup()

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

		testbuilder.Run(func() {
			tex, cleanup := rendertest.GivenATexture("checker/0.png")
			defer cleanup()

			err := debug.DumpTextureAsPng(tex, pngBuffer)
			if err != nil {
				t.Fatalf("dumping failed: %v", err)
			}
		})

		dumpedImage, err := png.Decode(pngBuffer)
		if err != nil {
			panic(fmt.Errorf("couldn't decode pngBuffer: %w", err))
		}

		// When things are quite broken, we'll just ellide noisy log messages.
		logCount := 0
		logProblem := func(args ...interface{}) {
			logCount++
			if logCount > 10 {
				return
			}
			if logCount == 10 {
				t.Log("(supressing duplicate messages")
			}

			t.Log(args...)
		}

		// The dump should have produced a 64x64 pixel image of a cycle of squares
		// (each 4x4 pixels) that are black, transparent then red.
		foreachPixel(dumpedImage, func(x, y int, col color.Color) {
			idx := (y/4)*16 + (x/4)*1
			switch idx % 3 {
			case 0:
				if !isBlack(col) {
					logProblem("non-black pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			case 1:
				if !isTransparent(col) {
					logProblem("non-transparent pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			case 2:
				if !isRed(col) {
					logProblem("non-red pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			}
		})
	})

	t.Run("textures should match the underlying image", func(t *testing.T) {
		// - Load an image
		expectedImage := rendertest.MustLoadTestImage("checker")

		expectedImage = blitOntoBlue(expectedImage)

		// - Screen-shotting reads every pixel without prior knowledge; any
		// 'transparent' pieces will be set to the clear-colour. Set a blue
		// clear-colour (blue does not exist in the checker image), then draw the
		// checker image over top a blue background. Comparing the screenshot to
		// the drawn image is what we need to do.

		bounds := expectedImage.Bounds()
		width, height := bounds.Dx(), bounds.Dy()
		testbuilder.New().WithSize(width, height).WithQueue().Run(func(queue render.RenderQueueInterface) {
			var result bool
			var resultImage image.Image
			queue.Queue(func(st render.RenderQueueState) {
				// - Convert it to a texture
				tex, cleanup := rendertest.GivenATexture("checker/0.png")
				defer cleanup()

				render.WithBlankScreen(0, 0, 1, 1, func() {
					// - Blit the texture accross the entire viewport
					rendertest.DrawTexturedQuad(bounds, tex, st.Shaders())

					// - Verify a screenshot matches the image.
					resultImage = debug.ScreenShotNrgba(width, height)

					result = rendertest.ImagesAreWithinThreshold(expectedImage, resultImage, rendertest.Threshold(3), color.RGBA{R: 0, G: 0, B: 255, A: 255})
				})
			})
			queue.Purge()

			if result != true {
				// stash a copy of the broken image we just made
				imgmanip.MustDumpImage(resultImage, rendertest.MakeRejectName("testdata/checker/0.png", ".png"))
				t.Fatalf("image mismatch")
			}
		})
	})
}
