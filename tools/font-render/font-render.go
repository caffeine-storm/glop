package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <dict.gob> <string-to-render>\n", os.Args[0])
		os.Exit(1)
	}
	fromFile := os.Args[1]
	stringToRender := os.Args[2]

	runtime.LockOSThread()
	sys := system.Make(gos.NewSystemInterface(), gin.MakeLogged(glog.InfoLogger()))
	wdx := 1024
	wdy := 750

	sys.Startup()
	renderQueue := render.MakeQueue(func(render.RenderQueueState) {
		sys.CreateWindow(0, 0, wdx, wdy)
		sys.EnableVSync(true)
		err := gl.Init()
		if err != 0 {
			panic(err)
		}
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	renderQueue.StartProcessing()

	dictReader, err := os.Open(fromFile)
	if err != nil {
		panic(err)
	}

	logger := glog.New(&glog.Opts{
		Level: glog.LevelTrace,
	})
	d, err := gui.LoadAndInitializeDictionary(dictReader, renderQueue, logger)
	if err != nil {
		panic(err)
	}

	logger.Debug("pre-render-queue", "stringToRender", stringToRender, "max height", d.MaxHeight())
	renderQueue.Queue(func(st render.RenderQueueState) {
		gl.Ortho(0, float64(wdx), 0, float64(wdy), 100, -100)
		d.RenderString(stringToRender, gui.Point{}, d.MaxHeight(), gui.Left, st.Shaders())
		sys.SwapBuffers()
	})
	renderQueue.Purge()

	var in string
	fmt.Fprintf(os.Stderr, "hit enter to exit\n")
	fmt.Fscanf(os.Stdin, "%s", &in)
}
