package imgmanip

import (
	"fmt"
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
	ret := canv.Image.At(x, y)
	fmt.Printf("getting %d %d %v\n", x, y, ret)
	return ret
}

func (canv *InvertedCanvas) Set(x, y int, c color.Color) {
	y = canv.Bounds().Dy() - y - 1
	fmt.Printf("setting %d %d %v\n", x, y, c)
	canv.Image.Set(x, y, c)
}
