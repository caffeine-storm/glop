package rendertest

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gos"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

type conveyIsHalting struct {
	s string
}

func (c *conveyIsHalting) Error() string {
	return c.s
}

type glContext struct {
	sys              system.System
	windowHandle     system.NativeWindowHandle
	render           render.RenderQueueInterface
	recordedFailures []error
}

func newGlContextForTest(width, height int) *glContext {
	sys, windowHandle, renderQueue := newGlWindowForTest(width, height)
	ctx := &glContext{
		sys:          sys,
		windowHandle: windowHandle,
		render:       renderQueue,
		// Note: recordedFailures should be considered local to the render thread;
		// reading/writing to it must synchronize accordingly.
		recordedFailures: nil,
	}
	renderQueue.AddErrorCallback(func(q render.RenderQueueInterface, e error) {
		ctx.recordedFailures = append(ctx.recordedFailures, e)
	})
	return ctx
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

func (ctx *glContext) prep(width, height int, invariantscheck bool) (prepError error) {
	if ctx.windowHandle == nil {
		panic(fmt.Errorf("logic error: a glContext should hang onto a single NativeWindowHandle for its lifetime"))
	}

	// On the way out of this function, check for any errors to include from the
	// context.
	defer func() {
		prepError = errors.Join(prepError, ctx.takeLastError())
	}()

	// On the way into this function, check for any pre-existing errors in the
	// context.
	ctx.render.Purge()
	if e := ctx.takeLastError(); e != nil {
		prepError = e
		return fmt.Errorf("prep preconditions failed: %w", e)
	}

	ctx.render.Queue(func(render.RenderQueueState) {
		if invariantscheck {
			prepError = checkInvariants()
		}
		enforceInvariants()
		if prepError != nil {
			panic(fmt.Errorf("prep: invariants violated: %w", prepError))
		}

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

		// Push 'clear' and 'depth' things on the attribute stack for safe keeping
		// then set up our expected state.
		gl.PushAttrib(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(0, 0, 0, 1)
		gl.ClearDepth(1)
		gl.Disable(gl.DEPTH_TEST)
		gl.Disable(gl.BLEND)

		// SwapBuffers should flush the GL command queue and synchronize with the
		// X-server. Without doing so, things break!
		ctx.sys.SwapBuffers()

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	})
	ctx.render.Purge()

	return
}

func (ctx *glContext) clean(invariantscheck bool) (cleanError error) {
	defer func() {
		cleanError = errors.Join(cleanError, ctx.takeLastError())
	}()
	ctx.render.Queue(func(render.RenderQueueState) {
		// Undo server-side attribute push.
		gl.PopAttrib()

		// Undo matrix mode identity loads.
		gl.MatrixMode(gl.TEXTURE)
		gl.PopMatrix()

		gl.MatrixMode(gl.PROJECTION)
		gl.PopMatrix()

		gl.MatrixMode(gl.MODELVIEW)
		gl.PopMatrix()

		defer enforceInvariants()

		if invariantscheck {
			cleanError = checkInvariants()
			if cleanError != nil {
				panic(fmt.Errorf("clean: invariants violated: %w", cleanError))
			}
		}
	})
	ctx.render.Purge()

	return
}

func (ctx *glContext) run(fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) error {
	var err error
	func() {
		// We run this defer to capture any error that a test panics on.
		defer func() {
			if e := recover(); e != nil {
				switch v := e.(type) {
				case string:
					// It might be that Convey is trying to halt the tests; we need to
					// preserve the value in that case.
					if v == "___FAILURE_HALT___" {
						err = &conveyIsHalting{
							s: v,
						}
						return
					}
				case error:
					// Panicking on an error value is a way to signal test failure.
					// Capture a stacktrace to make it easier to debug test failures.
					//
					// Note that we can't just return the error value because it doesn't
					// hold a stacktrace. Stacktraces are normally printed when a
					// panicking goroutine hits the last of its defer'd handlers,
					// critically, while the stack still exists.
					buffer := debug.Stack()

					// We don't really need a stacktrace that's more than 1MB.
					buffer = buffer[:min(len(buffer), 1024*1024)]

					err = fmt.Errorf("test failure: %w\n---test-stacktrace:\n%s---end-of-test-stacktrace", v, string(buffer))
					return
				}
				// Otherwise, someone panicked with a non-error which is, in a way,
				// even worse T_T. This will not be considered a 'test failure' but a
				// 'test error'. Subtly different but important to distinguish problems
				// in application code from problems in test code.
				panic(fmt.Errorf("recover() returned a non-error type: %T value: %v", e, e))
			}
		}()
		// TODO(tmckee:#40): we need to find a way to cleanup even if 'fn' calls
		// runtime.Goexit(). We could spawn a sacrificial goroutine to do this call
		// but we'd still need a way to recover if someone calls t.Fatalf on the
		// render thread.
		fn(ctx.sys, ctx.windowHandle, Failfastqueue(ctx))
	}()

	return errors.Join(err, ctx.takeLastError())
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
		// be able to plumb a testing.T instance in here and call its t.Fail.
		glog.ErrorLogger().Error("test-render-queue.OnError", "err", e)
	})
	renderQueue.StartProcessing()

	return sys, <-hdl, renderQueue
}
