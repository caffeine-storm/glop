package imgmanip

import (
	"image/color"
	"image/draw"
)

type InvertedCanvas struct {
	draw.Image
}

var _ draw.Image = ((*InvertedCanvas)(nil))

func NewInvertedCanvas(img draw.Image) *InvertedCanvas {
	return &InvertedCanvas{
		Image: img,
	}
}

func (canv *InvertedCanvas) At(x, y int) color.Color {
	y = canv.Bounds().Dy() - y - 1
	return canv.Image.At(x, y)
}

func (canv *InvertedCanvas) Set(x, y int, c color.Color) {
	y = canv.Bounds().Dy() - y - 1
	canv.Image.Set(x, y, c)
}
