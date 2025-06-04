package imgmanip

import (
	"image"
	"image/color"
	"image/draw"
)

func ToNRGBA(img image.Image) *image.NRGBA {
	casted, ok := img.(*image.NRGBA)
	if ok {
		return casted
	}

	result := image.NewNRGBA(img.Bounds())
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Src)

	return result
}

func DrawAsNrgbaWithBackground(img image.Image, bg color.Color) *image.NRGBA {
	ret := image.NewNRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Over)
	return ret
}
