package rendertest

import (
	"fmt"

	"github.com/MobRulesGames/mathgl"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func matrixStacksMustBeSize1() {
	sizes := [3]int{
		gl.GetInteger(gl.MODELVIEW_STACK_DEPTH),
		gl.GetInteger(gl.PROJECTION_STACK_DEPTH),
		gl.GetInteger(gl.TEXTURE_STACK_DEPTH),
	}

	if sizes[0] != 1 || sizes[1] != 1 || sizes[2] != 1 {
		panic(fmt.Errorf("matrix stacks needed to all be size 1: stack sizes: %+v", sizes))
	}
}

func matrixStacksMustBeIdentity() {
	var buffer [3]mathgl.Mat4

	gl.GetFloatv(gl.MODELVIEW_MATRIX, buffer[0][:])
	gl.GetFloatv(gl.PROJECTION_MATRIX, buffer[1][:])
	gl.GetFloatv(gl.TEXTURE_MATRIX, buffer[2][:])

	if buffer[0].IsIdentity() && buffer[1].IsIdentity() && buffer[2].IsIdentity() {
		return
	}

	panic(fmt.Errorf(
		`matrix stacks needed to be topped with identity matrices:
modelview:
%v
projection:
%v
texture:
%v`, render.Showmat(buffer[0]), render.Showmat(buffer[1]), render.Showmat(buffer[2])))
}

func checkMatrixInvariants() {
	// If the matrix stacks are size 1 with the identity on top, something is
	// wrong.
	matrixStacksMustBeSize1()
	matrixStacksMustBeIdentity()
}

func newGlWindowForTest(width, height int) (system.System, system.NativeWindowHandle, render.RenderQueueInterface) {
	linuxSystemObject := gos.NewSystemInterface()
	sys := system.Make(linuxSystemObject, gin.MakeLogged(glog.DebugLogger()))

	// Use a channel to wait for a NativeWindowHandle to show up; we want to let
	// initialization happen off-thread but the glContext needs to know the
	// native window id immediately.
	hdl := make(chan system.NativeWindowHandle)

	sys.Startup()
	render := render.MakeQueue(func(render.RenderQueueState) {
		hdl <- sys.CreateWindow(0, 0, width, height)

		sys.EnableVSync(true)
		err := gl.Init()
		if err != 0 {
			panic(fmt.Errorf("couldn't gl.Init: %d", err))
		}
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

		sys.SwapBuffers()
	})
	render.StartProcessing()

	return sys, <-hdl, render
}

type glContext struct {
	sys          system.System
	windowHandle system.NativeWindowHandle
	render       render.RenderQueueInterface
}

func (ctx *glContext) Prep(width, height int) {
	if ctx.windowHandle == nil {
		panic(fmt.Errorf("logic error: a glContext should hang onto a single NativeWindowHandle for its lifetime"))
	}

	ctx.render.Queue(func(render.RenderQueueState) {
		checkMatrixInvariants()

		ctx.sys.SetWindowSize(width, height)

		gl.MatrixMode(gl.MODELVIEW)
		gl.PushMatrix()
		gl.LoadIdentity()

		// Use an orthographic projection because all the gui code assumes it's
		// rendering with such a projection.
		gl.Ortho(0, float64(width), 0, float64(height), 1000, -1000)
		gl.Viewport(0, 0, width, height)

		gl.MatrixMode(gl.PROJECTION)
		gl.PushMatrix()
		gl.LoadIdentity()

		gl.MatrixMode(gl.TEXTURE)
		gl.PushMatrix()
		gl.LoadIdentity()

		gl.ClearColor(0, 0, 0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// SwapBuffers should flush the GL command queue and synchronize with the
		// X-server. Without doing so, things break!
		ctx.sys.SwapBuffers()
	})
	ctx.render.Purge()
}

func (ctx *glContext) Clean() {
	ctx.render.Queue(func(render.RenderQueueState) {
		// Undo matrix mode identity loads
		gl.MatrixMode(gl.TEXTURE)
		gl.PopMatrix()

		gl.MatrixMode(gl.PROJECTION)
		gl.PopMatrix()

		gl.MatrixMode(gl.MODELVIEW)
		gl.PopMatrix()

		checkMatrixInvariants()
	})
	ctx.render.Purge()
}

func (ctx *glContext) Run(fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	fn(ctx.sys, ctx.windowHandle, ctx.render)
}

func newGlContextForTest(width, height int) *glContext {
	sys, windowHandle, render := newGlWindowForTest(width, height)
	return &glContext{
		sys:          sys,
		windowHandle: windowHandle,
		render:       render,
	}
}

var glTestContextSource = make(chan *glContext, 24)

func WithGlAndHandleForTest(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	select {
	case cachedContext := <-glTestContextSource:
		cachedContext.Prep(width, height)
		cachedContext.Run(fn)
		cachedContext.Clean()
		glTestContextSource <- cachedContext
	default:
		newContext := newGlContextForTest(width, height)
		newContext.Prep(width, height)
		newContext.Run(fn)
		newContext.Clean()
		glTestContextSource <- newContext
	}
}

func WithIsolatedGlAndHandleForTest(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	newContext := newGlContextForTest(width, height)
	newContext.Prep(width, height)
	newContext.Run(fn)
	newContext.Clean()
}

func WithGlForTest(width, height int, fn func(system.System, render.RenderQueueInterface)) {
	WithGlAndHandleForTest(width, height, func(sys system.System, _ system.NativeWindowHandle, queue render.RenderQueueInterface) {
		fn(sys, queue)
	})
}

func WithGl(fn func()) {
	WithGlForTest(50, 50, func(sys system.System, renderQueue render.RenderQueueInterface) {
		logger := glog.ErrorLogger()

		errors := []gl.GLenum{}
		renderQueue.Queue(func(render.RenderQueueState) {
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
		renderQueue.Purge()

		// If there were GL errors _caused_ by the given func, fail!
		if len(errors) > 0 {
			// TODO(tmckee): add a helper to strerror error codes from GL.
			panic(fmt.Errorf("WithGl ran a func that produced %d GL errors: %v", len(errors), errors))
		}
	})
}

type GlTestBuilder struct{}

func (b *GlTestBuilder) Run(fn func()) {
	WithGl(fn)
}

func GlTest() *GlTestBuilder {
	return &GlTestBuilder{}
}
