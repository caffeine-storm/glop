package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func mustRgba(img image.Image) *image.RGBA {
	ret, ok := img.(*image.RGBA)
	if ok {
		return ret
	}

	ret = image.NewRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Src)

	return ret
}

type Delta struct {
	lhsColour color.Color
	rhsColour color.Color
	location  image.Point
}

func imageCompare(lhs, rhs image.Image) ([]Delta, int) {
	bounds := lhs.Bounds()
	if bounds != rhs.Bounds() {
		return nil, 0
	}

	baddies := []Delta{}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			lhsColour := lhs.At(x, y)
			rhsColour := rhs.At(x, y)
			if lhs.At(x, y) != rhs.At(x, y) {
				baddies = append(baddies, Delta{
					lhsColour: lhsColour,
					rhsColour: rhsColour,
					location: image.Point{
						X: x,
						Y: y,
					},
				})
			}
		}
	}

	return baddies, bounds.Dx() * bounds.Dy()
}

func mustPng(fname string) image.Image {
	file, err := os.Open(fname)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open %q: %w", fname, err))
	}
	defer file.Close()

	ret, err := png.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't png.Decode %q: %w", fname, err))
	}

	return ret
}

func main() {
	if len(os.Args) != 3 {
		panic("usage: png-cmp a.png b.png")
	}

	lhs, rhs := os.Args[1], os.Args[2]

	lhsPng := mustPng(lhs)
	rhsPng := mustPng(rhs)

	lhsRgba := mustRgba(lhsPng)
	rhsRgba := mustRgba(rhsPng)

	fmt.Println(bytes.Compare(lhsRgba.Pix, rhsRgba.Pix))
	deltas, _ := imageCompare(lhsRgba, rhsRgba)

	for _, delta := range deltas {
		fmt.Printf("%+v\n", delta)

		r, g, b, _ := delta.lhsColour.RGBA()
		if r != g || g != b {
			fmt.Println("not-grey", delta.location, delta.lhsColour)
		}

		r, g, b, _ = delta.lhsColour.RGBA()
		if r != g || g != b {
			fmt.Println("not-grey", delta.location, delta.rhsColour)
		}
	}
}
