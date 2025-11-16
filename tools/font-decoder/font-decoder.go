package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"

	"github.com/caffeine-storm/gl"
	"github.com/caffeine-storm/glop/gin"
	"github.com/caffeine-storm/glop/glog"
	"github.com/caffeine-storm/glop/gos"
	"github.com/caffeine-storm/glop/gui"
	"github.com/caffeine-storm/glop/render"
	"github.com/caffeine-storm/glop/system"
)

func main() {
	fromFile := os.Args[1]
	toFile := os.Args[2]

	runtime.LockOSThread()
	sys := system.Make(gos.NewSystemInterface(), gin.MakeLogged(glog.InfoLogger()))
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

	d, err := gui.LoadAndInitializeDictionary(dictReader, render, glog.TraceLogger())
	if err != nil {
		panic(err)
	}

	f, err := os.Create(toFile)
	if err != nil {
		panic(err)
	}

	img := image.NRGBA{
		Pix:    d.Data.Pix,
		Stride: 4 * d.Data.Dx,
		Rect:   image.Rect(0, 0, d.Data.Dx, d.Data.Dy),
	}

	err = png.Encode(f, &img)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", d)
}
