package imgmanip

import (
	"image"
	"image/color"
	"image/draw"
)

type interpolatingImage struct {
	image.Image
	xratio, yratio float32
}

func (img *interpolatingImage) Bounds() image.Rectangle {
	oldBounds := img.Image.Bounds()
	return image.Rect(
		0, 0,
		int(float32(oldBounds.Dx())*img.xratio),
		int(float32(oldBounds.Dy())*img.yratio),
	)
}

func (img *interpolatingImage) At(x, y int) color.Color {
	// TODO(tmckee:#14): we probably want to sample nearby points and blend them
	// when scaling up.
	newx, newy := int(float32(x)/img.xratio), int(float32(y)/img.yratio)
	minpoint := img.Image.Bounds().Min
	return img.Image.At(newx+minpoint.X, newy+minpoint.Y)
}

func Scale(img image.Image, xratio, yratio float32) image.Image {
	interpolated := &interpolatingImage{
		Image:  img,
		xratio: xratio,
		yratio: yratio,
	}

	newBounds := interpolated.Bounds()
	result := image.NewNRGBA(newBounds)

	draw.Draw(result, newBounds, interpolated, image.Point{}, draw.Src)

	return result
}
