package main

import (
	"os"
	"runtime"
	"time"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func main() {
	runtime.LockOSThread()
	sys := system.Make(gos.GetSystemInterface())
	wdx := 1024
	wdy := 750

	sys.Startup()
	render.Init()
	render.Queue(func() {
		sys.CreateWindow(10, 10, wdx, wdy)
		sys.EnableVSync(true)
		err := gl.Init()
		if err != 0 {
			panic(err)
		}
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	render.Purge()

	dictReader, err := os.Open("../../testdata/fonts/dict_10.gob")
	if err != nil {
		panic(err)
	}

	d, err := gui.LoadDictionary(dictReader)
	if err != nil {
		panic(err)
	}

	render.Queue(func() {
		sys.SwapBuffers()
		d.RenderString("lol", 0, 0, 0, 12.0, gui.Left)
	})
	render.Purge()

	sys.Think()

	render.Queue(func() {
		sys.SwapBuffers()
		d.RenderString("lol", 0, 0, 0, 12.0, gui.Left)
	})
	render.Purge()

	sys.Think()

	time.Sleep(1 * time.Second)
}
