package rendertest

import (
	"fmt"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func InitGlForTest(width, height int) (system.System, render.RenderQueueInterface) {
	linuxSystemObject := gos.GetSystemInterface()
	sys := system.Make(linuxSystemObject)

	sys.Startup()
	render := render.MakeQueue(func() {
		sys.CreateWindow(0, 0, width, height)
		sys.EnableVSync(true)
		err := gl.Init()
		if err != 0 {
			panic(fmt.Errorf("couldn't gl.Init: %d", err))
		}
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	})
	render.StartProcessing()

	return sys, render
}

func WithGl(fn func()) {
	_, render := InitGlForTest(50, 50)
	logger := glog.ErrorLogger()

	errors := []gl.GLenum{}
	render.Queue(func() {
		// Clear out GL's error queue so that a leaky test doesn't break us by
		// accident.
		for err := gl.GetError(); err != 0; err = gl.GetError() {
			logger.Error("glErrors before given func", "error_code", err)
			err = gl.GetError()
		}

		fn()

		for err := gl.GetError(); err != 0; err = gl.GetError() {
			errors = append(errors, err)
			err = gl.GetError()
		}
	})
	render.Purge()

	// If there were GL errors _caused_ by the given func, fail!
	if len(errors) > 0 {
		// TODO(tmckee): add a helper to strerror error codes from GL.
		panic(fmt.Errorf("WithGl ran a func that produced %d GL errors: %v", len(errors), errors))
	}
}
