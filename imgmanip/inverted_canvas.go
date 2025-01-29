package imgmanip

import "image/draw"

type InvertedCanvas struct {
	draw.Image
}

func NewInvertedCanvas(img draw.Image) *InvertedCanvas {
	return &InvertedCanvas{
		Image: img,
	}
}
