package gui

import (
	"image"
	"image/color"
	"sort"
)

type subImage struct {
	im     image.Image
	bounds image.Rectangle
}
type transparent struct{}

func (t transparent) RGBA() (r, g, b, a uint32) {
	return 0, 0, 0, 0
}
func (si *subImage) ColorModel() color.Model {
	return si.im.ColorModel()
}
func (si *subImage) Bounds() image.Rectangle {
	return si.bounds
}
func (si *subImage) At(x, y int) color.Color {
	b := si.bounds
	if (image.Point{x, y}).In(b) {
		return si.im.At(x, y)
	}
	return transparent{}
}

// Returns a sub-image of the input image. The bounding rectangle is the
// smallest possible rectangle that includes all pixels that have alpha > 0,
// with one pixel of border on all sides.
// TODO(tmckee): ought to be able to save a lot of effort by skipping internal
// pixels; i.e. each row has a min/max X pixel set; we don't need to check
// between them.
func MinimalSubImage(src image.Image) *subImage {
	bounds := src.Bounds()
	var new_bounds image.Rectangle
	new_bounds.Max = bounds.Min
	new_bounds.Min = bounds.Max
	for x := bounds.Min.X; x <= bounds.Max.X; x++ {
		for y := bounds.Min.Y; y <= bounds.Max.Y; y++ {
			c := src.At(x, y)
			_, _, _, a := c.RGBA()
			if a > 0 {
				if x < new_bounds.Min.X {
					new_bounds.Min.X = x
				}
				if y < new_bounds.Min.Y {
					new_bounds.Min.Y = y
				}
				if x > new_bounds.Max.X {
					new_bounds.Max.X = x
				}
				if y > new_bounds.Max.Y {
					new_bounds.Max.Y = y
				}
			}
		}
	}

	// We want one row/col of boundary between characters so that we don't get
	// annoying artifacts
	new_bounds.Min.X--
	new_bounds.Min.Y--
	new_bounds.Max.X++
	new_bounds.Max.Y++

	if new_bounds.Min.X > new_bounds.Max.X || new_bounds.Min.Y > new_bounds.Max.Y {
		new_bounds = image.Rect(0, 0, 0, 0)
	}

	return &subImage{
		im:     src,
		bounds: new_bounds,
	}
}

// This stupid thing is just so that our idiot-packedImage can answer queries
// faster.  If we're going to query every pixel then it makes sense to check
// the largest rectangles first, since they will be the correct response more
// often than the smaller rectangles.
type packedImageSortByArea struct {
	*packedImage
}

func (p *packedImageSortByArea) Len() int {
	return len(p.ims)
}
func (p *packedImageSortByArea) Less(i, j int) bool {
	ai := p.ims[i].Bounds().Dx() * p.ims[i].Bounds().Dy()
	aj := p.ims[j].Bounds().Dx() * p.ims[j].Bounds().Dy()
	return ai > aj
}
func (p *packedImageSortByArea) Swap(i, j int) {
	p.ims[i], p.ims[j] = p.ims[j], p.ims[i]
	p.off[i], p.off[j] = p.off[j], p.off[i]
}

type packedImage struct {
	ims    []image.Image
	off    []image.Point
	bounds image.Rectangle
}

func (p *packedImage) Len() int {
	return len(p.ims)
}
func (p *packedImage) Less(i, j int) bool {
	return p.ims[i].Bounds().Dy() < p.ims[j].Bounds().Dy()
}
func (p *packedImage) Swap(i, j int) {
	p.ims[i], p.ims[j] = p.ims[j], p.ims[i]
	p.off[i], p.off[j] = p.off[j], p.off[i]
}
func (p *packedImage) GetPackedLocation(im image.Image) image.Rectangle {
	for i := range p.ims {
		if im == p.ims[i] {
			return p.ims[i].Bounds().Add(p.off[i])
		}
	}
	return image.Rectangle{}
}
func (p *packedImage) ColorModel() color.Model {
	return p.ims[0].ColorModel()
}
func (p *packedImage) Bounds() image.Rectangle {
	return p.bounds
}
func (p *packedImage) At(x, y int) color.Color {
	point := image.Point{x, y}
	for i := range p.ims {
		if point.In(p.ims[i].Bounds().Add(p.off[i])) {
			return p.ims[i].At(x-p.off[i].X, y-p.off[i].Y)
		}
	}
	return transparent{}
}

func packImages(ims []image.Image) *packedImage {
	var p packedImage
	if len(ims) == 0 {
		panic("Cannot pack zero images")
	}
	p.ims = ims
	p.off = make([]image.Point, len(p.ims))
	// Sorts p.ims by height
	sort.Sort(&p)

	run := 0
	height := 0
	max_width := 512
	max_height := 0
	for i := 1; i < len(p.off); i++ {
		run += p.ims[i-1].Bounds().Dx()
		if run+p.ims[i].Bounds().Dx() > max_width {
			run = 0
			height += max_height
			max_height = 0
		}
		if p.ims[i].Bounds().Dy() > max_height {
			max_height = p.ims[i].Bounds().Dy()
		}
		p.off[i].X = run
		p.off[i].Y = height
	}
	for i := range p.ims {
		p.off[i] = p.off[i].Sub(p.ims[i].Bounds().Min)
	}

	// Done packing - now figure out the resulting bounds
	p.bounds.Min.X = 1e9 // if we exceed this something else will break first
	p.bounds.Min.Y = 1e9
	p.bounds.Max.X = -1e9
	p.bounds.Max.Y = -1e9
	for i := range p.ims {
		b := p.ims[i].Bounds()
		min := b.Add(p.off[i]).Min
		max := b.Add(p.off[i]).Max
		if min.X < p.bounds.Min.X {
			p.bounds.Min.X = min.X
		}
		if min.Y < p.bounds.Min.Y {
			p.bounds.Min.Y = min.Y
		}
		if max.X > p.bounds.Max.X {
			p.bounds.Max.X = max.X
		}
		if max.Y > p.bounds.Max.Y {
			p.bounds.Max.Y = max.Y
		}
	}

	sort.Sort(&packedImageSortByArea{&p})

	return &p
}
