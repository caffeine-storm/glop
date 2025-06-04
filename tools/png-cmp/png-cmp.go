package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func mustNrgba(img image.Image) *image.NRGBA {
	ret, ok := img.(*image.NRGBA)
	if ok {
		return ret
	}

	ret = image.NewNRGBA(img.Bounds())
	draw.Draw(ret, img.Bounds(), img, image.Point{}, draw.Src)

	return ret
}

type Delta struct {
	lhsColour color.Color
	rhsColour color.Color
	location  image.Point
}

func colourDistance(lhs, rhs color.Color) (dr, dg, db int) {
	lr, lg, lb, _ := lhs.RGBA()
	rr, rg, rb, _ := rhs.RGBA()

	dr = int(rr) - int(lr)
	dg = int(rg) - int(lg)
	db = int(rb) - int(lb)

	if dr < 0 {
		dr = -dr
	}
	if dg < 0 {
		dg = -dg
	}
	if db < 0 {
		db = -db
	}

	return
}

func ImageCompare(lhs, rhs *image.NRGBA) ([]Delta, int, Delta) {
	bounds := lhs.Bounds()
	if bounds != rhs.Bounds() {
		return nil, 0, Delta{}
	}

	baddies := []Delta{}
	maxDelta := Delta{}
	maxDist := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
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
				dr, dg, db := colourDistance(lhs.At(x, y), rhs.At(x, y))
				if max(dr, dg, db) > maxDist {
					maxDist = max(dr, dg, db)
					maxDelta = baddies[len(baddies)-1]
				}

			}
		}
	}

	return baddies, bounds.Dx() * bounds.Dy(), maxDelta
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

	lhsRgba := mustNrgba(lhsPng)
	rhsRgba := mustNrgba(rhsPng)

	deltas, _, maxdelta := ImageCompare(lhsRgba, rhsRgba)

	for _, delta := range deltas {
		fmt.Printf("%+v\n", delta)

		lr, lg, lb, _ := delta.lhsColour.RGBA()
		if lr != lg || lg != lb {
			fmt.Println("not-grey", delta.location, delta.lhsColour)
		}

		rr, rg, rb, _ := delta.rhsColour.RGBA()
		if rr != rg || rg != rb {
			fmt.Println("not-grey", delta.location, delta.rhsColour)
		}
	}

	fmt.Printf("maxdelta: %+v\n", maxdelta)
}
