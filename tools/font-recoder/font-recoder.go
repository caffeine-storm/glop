package main

import (
	"image"
	"image/png"
	"log/slog"
	"os"
	"runtime"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func main() {
	fromFile := os.Args[1]
	toFile := os.Args[2]

	runtime.LockOSThread()
	sys := system.Make(gos.GetSystemInterface())
	wdx := 1024
	wdy := 750

	sys.Startup()
	render := render.MakeQueue(func(render.RenderQueueState) {
		sys.CreateWindow(10, 10, wdx, wdy)
		sys.EnableVSync(true)
		err := gl.Init()
		if err != 0 {
			panic(err)
		}
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	render.StartProcessing()

	dictReader, err := os.Open(fromFile)
	if err != nil {
		panic(err)
	}

	d, err := gui.LoadDictionary(dictReader, render, slog.Default())
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
