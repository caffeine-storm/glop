package imgmanip

import (
	"image"
	"image/color"
)

type Boundser interface {
	Bounds() image.Rectangle
}

// Rewrites the given input image flipping it vertically.
func FlipVertically[T Boundser](img Boundser, data []byte) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	tmp := make([]byte, width*4)
	for rowIdx := 0; rowIdx < height/2; rowIdx++ {
		a, b := rowIdx, height-rowIdx-1
		if a >= b {
			break
		}
		arow := data[a*width*4 : (a+1)*width*4]
		brow := data[b*width*4 : (b+1)*width*4]
		copy(tmp, arow)
		copy(arow, brow)
		copy(brow, tmp)
	}
}

// An image.Image that looks vertically flipped w.r.t. the embedded
// image.Image.
type VertFlipped struct {
	image.Image
}

func (v VertFlipped) At(x, y int) color.Color {
	flippedy := v.Bounds().Dy() - y - 1
	return v.Image.At(x, flippedy)
}
