package main

import (
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

		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()

		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		gl.ClearColor(1, 0, 0, 1)
	})
	render.Purge()

	// 0. Pre-load a font (TextLine takes a name-of-loaded-font to use when
	// rendering)
	// 1. Call TextLine constructor (MakeTextLine)
	// 2. reg := gui.Region{Point{x, y}, Dims{width, height}}
	// 3. ourTextLine.Draw(reg)
	// 4. swap buffers
	skiaTtfPath := "./skia.ttf"
	gui.MustLoadFontAs(skiaTtfPath, "glop.font")

	textLine := gui.MakeTextLine("glop.font", "lol", 200, 0, 1, 0, 1)
	if textLine == nil {
		panic("nil textLine returned")
	}

	region := gui.Region{
		Point: gui.Point{X: 0, Y: 0},
		Dims: gui.Dims{Dx: 50, Dy: 50},
	}

	for {
		sys.Think()
		render.Queue(func() {
			gl.Clear(gl.COLOR_BUFFER_BIT);

			gl.Begin(gl.QUADS)
			gl.Vertex2d(-0.5, -0.5)
			gl.Vertex2d( 0.5, -0.5)
			gl.Vertex2d( 0.5,  0.5)
			gl.Vertex2d(-0.5,  0.5)
			gl.End()

			textLine.Draw(region)

			sys.SwapBuffers()
		})
		render.Purge()

		time.Sleep(time.Millisecond * 100)
	}
}
