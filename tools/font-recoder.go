package main

import (
	"image"
	"image/png"
	"os"

	"github.com/runningwild/glop/gui"
)

func main() {
	fromFile := os.Args[0]
	toFile := os.Args[1]

	dictReader, err := os.Open(fromFile)
	if err != nil {
		panic(err)
	}

	d, err := gui.LoadDictionary(dictReader)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(toFile)
	if err != nil {
		panic(err)
	}

	img := image.RGBA{
		Pix:    d.Data.Pix,
		Stride: 4 * d.Data.Dx,
		Rect: image.Rectangle{
			Min: image.Point{
				X: 0,
				Y: 0,
			},
			Max: image.Point{
				X: d.Data.Dx,
				Y: d.Data.Dy,
			},
		},
	}

	err = png.Encode(f, &img)
	if err != nil {
		panic(err)
	}
}

