package rendertest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MobRulesGames/mathgl"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func getBadMatrixStackSizes() map[string]int {
	sizes := [3]int{
		gl.GetInteger(gl.MODELVIEW_STACK_DEPTH),
		gl.GetInteger(gl.PROJECTION_STACK_DEPTH),
		gl.GetInteger(gl.TEXTURE_STACK_DEPTH),
	}
	modes := [3]string{
		"modelview",
		"projection",
		"texture",
	}

	ret := map[string]int{}

	for i, sz := range sizes {
		if sz != 1 {
			ret[modes[i]] = sz
		}
	}

	return ret
}

func getBadMatrixValues() map[string]mathgl.Mat4 {
	var buffer [3]mathgl.Mat4

	gl.GetFloatv(gl.MODELVIEW_MATRIX, buffer[0][:])
	gl.GetFloatv(gl.PROJECTION_MATRIX, buffer[1][:])
	gl.GetFloatv(gl.TEXTURE_MATRIX, buffer[2][:])

	ret := map[string]mathgl.Mat4{}

	if !buffer[0].IsIdentity() {
		ret["modelview"] = buffer[0]
	}
	if !buffer[1].IsIdentity() {
		ret["projection"] = buffer[1]
	}
	if !buffer[2].IsIdentity() {
		ret["texture"] = buffer[2]
	}
	return ret

}

func mustSatisfyMatrixInvariants() {
	// If the matrix stacks are size 1 with the identity on top, something is
	// wrong.
	mp := getBadMatrixStackSizes()
	if len(mp) > 0 {
		panic(fmt.Errorf("matrix stacks needed to all be size 1: stack sizes: %+v", mp))
	}
	mpp := getBadMatrixValues()
	if len(mpp) > 0 {
		reports := []string{}
		for key, val := range mpp {
			reports = append(reports, fmt.Sprintf("%s:\n%v", key, render.Showmat(val)))
		}
		panic(fmt.Errorf("matrix stacks needed to be topped with identity matrices:\n%s", strings.Join(reports, "\n")))
	}
}

func getImproperlyBoundState() []gl.GLenum {
	bindings := []gl.GLenum{
		gl.ARRAY_BUFFER_BINDING,
		gl.ELEMENT_ARRAY_BUFFER_BINDING,
		gl.PIXEL_PACK_BUFFER_BINDING,
		gl.PIXEL_UNPACK_BUFFER_BINDING,
		gl.TEXTURE_BINDING_2D,
	}

	badvals := []gl.GLenum{}
	for _, name := range bindings {
		val := gl.GetInteger(name)
		if val != 0 {
			badvals = append(badvals, name)
		}
	}

	return badvals
}

func mustSatisfyBindingsInvariants() {
	badvals := getImproperlyBoundState()
	if len(badvals) > 0 {
		panic(fmt.Errorf("need bindings unset but found bindings for: %v", badvals))
	}
}

func mustSatisfyInvariants() {
	mustSatisfyMatrixInvariants()
	mustSatisfyBindingsInvariants()
}

func enforceMatrixStacksMustBeIdentitySingletons() {
	sizes := [3]int{
		gl.GetInteger(gl.MODELVIEW_STACK_DEPTH),
		gl.GetInteger(gl.PROJECTION_STACK_DEPTH),
		gl.GetInteger(gl.TEXTURE_STACK_DEPTH),
	}
	modes := [3]render.MatrixMode{
		render.MatrixModeModelView,
		render.MatrixModeProjection,
		render.MatrixModeTexture,
	}

	for i, sizei := range sizes {
		if sizei != 1 {
			glog.WarningLogger().Warn("rendertest enforcing matrix invariant", "state leakage", fmt.Sprintf("matrix mode %v", modes[i]))
		}
		gl.MatrixMode(gl.GLenum(modes[i]))
		for j := sizei; j > 1; j-- {
			gl.PopMatrix()
		}
	}

	badMats := getBadMatrixValues()
	if len(badMats) > 0 {
		glog.WarningLogger().Warn("rendertest enforcing matrix invariant", "state leakage", "one or more matrices had non-identity value", "variants", badMats)
	}

	for i := range sizes {
		gl.MatrixMode(gl.GLenum(modes[i]))
		gl.LoadIdentity()
	}
}

func enforceClearBindingsSet() {
	badBindings := getImproperlyBoundState()
	if len(badBindings) > 0 {
		glog.WarningLogger().Warn("rendertest enforcing bindings invariant", "state leakage", badBindings)
	}

	bufferBindings := []gl.GLenum{
		gl.ARRAY_BUFFER,
		gl.ELEMENT_ARRAY_BUFFER,
		gl.PIXEL_PACK_BUFFER,
		gl.PIXEL_UNPACK_BUFFER,
	}
	for _, name := range bufferBindings {
		gl.Buffer(0).Bind(name)
	}

	textureBindings := []gl.GLenum{
		gl.TEXTURE_2D,
	}
	for _, name := range textureBindings {
		gl.Texture(0).Bind(name)
	}
}

func enforceInvariants() {
	enforceMatrixStacksMustBeIdentitySingletons()
	enforceClearBindingsSet()
}

func newGlWindowForTest(width, height int) (system.System, system.NativeWindowHandle, render.RenderQueueInterface) {
	linuxSystemObject := gos.NewSystemInterface()
	sys := system.Make(linuxSystemObject, gin.MakeLogged(glog.DebugLogger()))

	// Use a channel to wait for a NativeWindowHandle to show up; we want to let
	// initialization happen off-thread but the glContext needs to know the
	// native window id immediately.
	hdl := make(chan system.NativeWindowHandle)

	sys.Startup()
	renderQueue := render.MakeQueue(func(render.RenderQueueState) {
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
	renderQueue.AddErrorCallback(func(q render.RenderQueueInterface, e error) {
		// TODO(tmckee:#38): we need better attribution here; it's hard right now to
		// know _which_ test was running the render job that panicked. We ought to
		// be able to plumb a testing.T instance in here and call its t.Fail/t.Fatalf
		glog.ErrorLogger().Error("test-render-queue.OnError", "err", e)
	})
	renderQueue.StartProcessing()

	return sys, <-hdl, renderQueue
}

type glContext struct {
	sys              system.System
	windowHandle     system.NativeWindowHandle
	render           render.RenderQueueInterface
	recordedFailures []error
}

const InvariantsCheckNo = false
const InvariantsCheckYes = true

// Helper for getting the last on-render-queue error. Clears the state used to
// track on-render-queue errors.
func (ctx *glContext) takeLastError() error {
	var allErrors []error
	ctx.render.Queue(func(render.RenderQueueState) {
		allErrors = ctx.recordedFailures
		ctx.recordedFailures = nil
	})
	ctx.render.Purge()

	return errors.Join(allErrors...)
}

func (ctx *glContext) prep(width, height int, invariantscheck bool) (err error) {
	if ctx.windowHandle == nil {
		panic(fmt.Errorf("logic error: a glContext should hang onto a single NativeWindowHandle for its lifetime"))
	}

	defer func() {
		err = errors.Join(err, ctx.takeLastError())
	}()

	ctx.render.Purge()
	if e := ctx.takeLastError(); e != nil {
		return fmt.Errorf("prep preconditions failed: %w", e)
	}

	ctx.render.Queue(func(render.RenderQueueState) {
		if invariantscheck {
			mustSatisfyInvariants()
		}
		enforceInvariants()

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

	return
}

func (ctx *glContext) clean(invariantscheck bool) (err error) {
	defer func() {
		err = ctx.takeLastError()
	}()
	ctx.render.Queue(func(render.RenderQueueState) {
		// Undo matrix mode identity loads
		gl.MatrixMode(gl.TEXTURE)
		gl.PopMatrix()

		gl.MatrixMode(gl.PROJECTION)
		gl.PopMatrix()

		gl.MatrixMode(gl.MODELVIEW)
		gl.PopMatrix()

		defer enforceInvariants()

		if invariantscheck {
			mustSatisfyInvariants()
		}
	})
	ctx.render.Purge()

	return
}

func (ctx *glContext) run(fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) error {
	var err error
	func() {
		defer func() {
			if e := recover(); e != nil {
				var ok bool
				err, ok = e.(error)
				if !ok {
					panic(fmt.Errorf("recover() returned a non-error: %v", e))
				}
			}
		}()
		fn(ctx.sys, ctx.windowHandle, Failfastqueue(ctx))
	}()

	return errors.Join(err, ctx.takeLastError())
}

// TODO(#37): prefer GlTest()
func DeprecatedWithGlAndHandleForTest(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	RunDeprecatedTestWithCachedContext(width, height, fn)
}

// TODO(#37): prefer GlTest()
func DeprecatedWithIsolatedGlAndHandleForTest(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	newContext := newGlContextForTest(width, height)
	newContext.prep(width, height, InvariantsCheckNo)
	newContext.run(fn)
	newContext.clean(InvariantsCheckNo)
}

// TODO(#37): prefer GlTest()
func DeprecatedWithGlForTest(width, height int, fn func(system.System, render.RenderQueueInterface)) {
	DeprecatedWithGlAndHandleForTest(width, height, func(sys system.System, _ system.NativeWindowHandle, queue render.RenderQueueInterface) {
		fn(sys, queue)
	})
}

// TODO(#37): prefer GlTest()
func DeprecatedWithGl(fn func()) {
	DeprecatedWithGlForTest(50, 50, func(sys system.System, renderQueue render.RenderQueueInterface) {
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
