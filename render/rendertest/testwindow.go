package rendertest

import (
	"fmt"

	"github.com/MobRulesGames/mathgl"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func matrixStacksMustBeSize1() {
	var buffer [3]int32
	gl.GetIntegerv(gl.MODELVIEW_STACK_DEPTH, buffer[0:1])
	gl.GetIntegerv(gl.PROJECTION_STACK_DEPTH, buffer[1:2])
	gl.GetIntegerv(gl.TEXTURE_STACK_DEPTH, buffer[2:3])

	if buffer[0] != 1 || buffer[1] != 1 || buffer[2] != 1 {
		panic(fmt.Errorf("matrix stacks needed to all be size 1: stack sizes: %+v", buffer))
	}
}

type showmat mathgl.Mat4

func (m showmat) String() string {
	return fmt.Sprintf(
		`%f, %f, %f, %f
%f, %f, %f, %f
%f, %f, %f, %f
%f, %f, %f, %f`,
		m[0], m[1], m[2], m[3],
		m[4], m[5], m[6], m[7],
		m[8], m[9], m[10], m[11],
		m[12], m[13], m[14], m[15],
	)
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
%v`, showmat(buffer[0]), showmat(buffer[1]), showmat(buffer[2])))
}

func newGlWindowForTest(width, height int) (system.System, render.RenderQueueInterface) {
	linuxSystemObject := gos.GetSystemInterface()
	sys := system.Make(linuxSystemObject)

	sys.Startup()
	render := render.MakeQueue(func(render.RenderQueueState) {
		sys.CreateWindow(0, 0, width, height)
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

	return sys, render
}

type glContext struct {
	sys    system.System
	render render.RenderQueueInterface
}

func (ctx *glContext) Prep(width, height int) {
	ctx.render.Queue(func(render.RenderQueueState) {
		// Each test should be allowed to run as if the GL context was freshly
		// initialized.
		matrixStacksMustBeSize1()
		matrixStacksMustBeIdentity()

		ctx.sys.SetWindowSize(width, height)

		gl.MatrixMode(gl.MODELVIEW)
		gl.PushMatrix()
		gl.LoadIdentity()

		gl.MatrixMode(gl.PROJECTION)
		gl.PushMatrix()
		gl.LoadIdentity()

		// Use an orthonormal projection because all the gui code assumes it's
		// rendering with such a projection.
		gl.Ortho(0, float64(width), 0, float64(height), 10, -10)
		gl.Viewport(0, 0, width, height)

		gl.MatrixMode(gl.TEXTURE)
		gl.PushMatrix()
		gl.LoadIdentity()

		gl.ClearColor(0, 0, 0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

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

		// If the matrix stacks are not length 1, something is wrong.
		matrixStacksMustBeSize1()
		matrixStacksMustBeIdentity()
	})
	ctx.render.Purge()
}

func (ctx *glContext) Run(fn func(system.System, render.RenderQueueInterface)) {
	fn(ctx.sys, ctx.render)
}

func newGlContextForTest(width, height int) *glContext {
	sys, render := newGlWindowForTest(width, height)
	return &glContext{
		sys:    sys,
		render: render,
	}
}

var glTestContextSource = make(chan *glContext, 24)

func WithGlForTest(width, height int, fn func(system.System, render.RenderQueueInterface)) {
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
