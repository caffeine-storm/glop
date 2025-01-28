package imgmanip

import "image"
import "image/draw"

func ToRGBA(img image.Image) *image.RGBA {
	casted, ok := img.(*image.RGBA)
	if ok {
		return casted
	}

	result := image.NewRGBA(img.Bounds())
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Src)

	return result
}
