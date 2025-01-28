package imgmanip

import (
	"image"
	"image/color"
	"image/draw"
)

func ToRGBA(img image.Image) *image.RGBA {
	casted, ok := img.(*image.RGBA)
	if ok {
		return casted
	}

	result := image.NewRGBA(img.Bounds())
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Src)

	return result
}

func DrawAsRgbaWithBackground(img image.Image, bg color.Color) *image.RGBA {
	ret := image.NewRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Over)
	return ret
}
