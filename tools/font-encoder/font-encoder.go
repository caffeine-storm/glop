package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func mustLoadFont(path string) *truetype.Font {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("couldn't read file %q: %w", path, err))
	}

	font, err := freetype.ParseFont(data)
	if err != nil {
		panic(fmt.Errorf("coudln't ParseFont: %w", err))
	}

	return font
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: %s <font.ttf> <point-size> <output.gob>\n", os.Args[0])
		os.Exit(1)
	}

	ttfFile := os.Args[1]
	pointSize := os.Args[2]
	outputFile := os.Args[3]

	runtime.LockOSThread()
	sys := system.Make(gos.NewSystemInterface(), gin.MakeLogged(glog.InfoLogger()))
	wdx := 1024
	wdy := 750

	sys.Startup()
	render := render.MakeQueue(func(render.RenderQueueState) {
		sys.CreateWindow(0, 0, wdx, wdy)
		sys.EnableVSync(true)
		err := gl.Init()
		if err != 0 {
			panic(err)
		}
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	render.StartProcessing()

	pointSizeInt, err := strconv.Atoi(pointSize)
	if err != nil {
		panic("couldn't parse %q as a point size")
	}

	trueTypeFont := mustLoadFont(ttfFile)

	d := gui.MakeDictionary(trueTypeFont, pointSizeInt, render, glog.VoidLogger())
	if d == nil {
		panic("gui.MakeDictionary returned nil!")
	}

	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d.Store(f)
}
