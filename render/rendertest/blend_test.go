package rendertest_test

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type normalizedColour struct {
	R, G, B, A float64
}

func normColour(R, G, B, A uint8) *normalizedColour {
	return &normalizedColour{
		R: float64(R) / 255,
		G: float64(G) / 255,
		B: float64(B) / 255,
		A: float64(A) / 255,
	}
}

func norm(b uint8) float64 {
	return float64(b) / 255
}

func (nc *normalizedColour) Within(epsilon float64, other *normalizedColour) bool {
	if math.Pow(nc.R-other.R, 2) > epsilon {
		return false
	}

	if math.Pow(nc.G-other.G, 2) > epsilon {
		return false
	}

	if math.Pow(nc.B-other.B, 2) > epsilon {
		return false
	}

	if math.Pow(nc.A-other.A, 2) > epsilon {
		return false
	}

	return true
}

func TestDrawingSetsExpectedAlpha(t *testing.T) {
	testcases := []uint8{
		0x00,
		0x40,
		0x80,
		0xc0,
		0xff,
	}
	assert := assert.New(t)
	onePixel := image.Rect(0, 0, 1, 1)

	// Note: draw.Draw expects colours that are alpha-premultiplied!
	for _, backgroundAlpha := range testcases {
		bg := image.NewUniform(color.RGBA{R: backgroundAlpha, A: backgroundAlpha})
		for _, foregroundAlpha := range testcases {
			fg := image.NewUniform(color.RGBA{B: foregroundAlpha, A: foregroundAlpha})

			canvas := image.NewRGBA(onePixel)

			// Draw the background into the canvas ignoring what's in the canvas
			// already. Effecitvely, just set the canvas to be the background.
			draw.Draw(canvas, onePixel, bg, image.Point{}, draw.Src)
			result := canvas.Pix
			expected := normColour(backgroundAlpha, 0, 0, backgroundAlpha)
			actual := normColour(result[0], result[1], result[2], result[3])
			assert.True(actual.Within(0.0001, expected), "expected: %v, actual: %v, fg: %v, bg: %v", expected, actual, foregroundAlpha, backgroundAlpha)

			// Draw the foreground over the backrgound so that things 'blend'.
			draw.Draw(canvas, onePixel, fg, image.Point{}, draw.Over)
			redExpected := norm(backgroundAlpha) * (1 - norm(foregroundAlpha))
			blueExpected := norm(foregroundAlpha)
			alphaExpected := norm(foregroundAlpha) + norm(backgroundAlpha)*(1-norm(foregroundAlpha))
			expected = &normalizedColour{R: redExpected, G: 0, B: blueExpected, A: alphaExpected}
			actual = normColour(result[0], result[1], result[2], result[3])
			assert.True(actual.Within(0.0001, expected), "expected: %v, actual: %v, fg: %v, bg: %v", expected, actual, foregroundAlpha, backgroundAlpha)
		}
	}
}
