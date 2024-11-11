package rendertest

import (
	"fmt"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func newGlWindowForTest(width, height int) (system.System, render.RenderQueueInterface) {
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

type glContext struct {
	sys system.System
	render render.RenderQueueInterface
}

func (ctx *glContext) Prep(width, height int) {
	ctx.render.Queue(func() {
		ctx.sys.SetWindowSize(width, height)

		gl.MatrixMode(gl.MODELVIEW)
		gl.PushMatrix()
		gl.LoadIdentity()

		gl.MatrixMode(gl.PROJECTION)
		gl.PushMatrix()
		gl.LoadIdentity()

		gl.MatrixMode(gl.TEXTURE)
		gl.PushMatrix()
		gl.LoadIdentity()

		gl.ClearColor(0, 0, 0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		ctx.sys.SwapBuffers()
		gl.Clear(gl.COLOR_BUFFER_BIT)
		ctx.sys.SwapBuffers()
	})
	ctx.render.Purge()
}

func (ctx *glContext) Clean() {
	// Undo matrix mode identity loads
	gl.MatrixMode(gl.TEXTURE)
	gl.PopMatrix()

	gl.MatrixMode(gl.PROJECTION)
	gl.PopMatrix()
	
	gl.MatrixMode(gl.MODELVIEW)
	gl.PopMatrix()
}

func (ctx *glContext) Run(fn func(system.System, render.RenderQueueInterface)) {
	fn(ctx.sys, ctx.render)
}

func newGlContextForTest(width, height int) *glContext {
	fmt.Printf(" --- new gl context: %dx%d --- \n", width, height)
	sys, render := newGlWindowForTest(width, height)
	return &glContext {
		sys: sys,
		render: render,
	}
}

var glTestContextSource = make(chan *glContext, 24)

func WithGlForTest(width, height int, fn func(system.System, render.RenderQueueInterface)) {
	select {
	case cachedContext := <- glTestContextSource:
		cachedContext.Prep(width, height)
		cachedContext.Run(fn)
		cachedContext.Clean()
		// glTestContextSource <- cachedContext
	default:
		newContext := newGlContextForTest(width, height)
		newContext.Run(fn)
		newContext.Clean()
		// glTestContextSource <- newContext
	}
}

func WithGl(fn func()) {
	WithGlForTest(50, 50, func(sys system.System, render render.RenderQueueInterface) {
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
	})
}
