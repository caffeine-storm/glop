package rendertest

import (
	"errors"
	"fmt"
	"sync"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func makeTestTemplate(checkInvariants bool, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) func(width, height int, ctx *glContext) {
	return func(width, height int, ctx *glContext) {
		e := ctx.prep(width, height, checkInvariants)
		if e != nil {
			// Even on error cases, we shouldn't leak GL state.
			ee := ctx.clean(InvariantsCheckNo)
			if ee != nil {
				panic(fmt.Errorf("after prep-failure: %w, couldn't clean: %w", e, ee))
			}
			panic(fmt.Errorf("couldn't prep: %w", e))
		}

		e = ctx.run(fn)

		ee := ctx.clean(checkInvariants)
		if ee != nil {
			panic(fmt.Errorf("couldn't clean: %w, testresult: %w", ee, e))
		}

		if e != nil {
			panic(fmt.Errorf("error on render-thread: %w", e))
		}
	}
}

// Like a render.renderQueue but, if there were on-render-thread errors,
// subsequent Purge() and Queue() calls will panic.
type failfast struct {
	render.RenderQueueInterface
	// TODO(clean): there's probabaly a way to not store two copies of these
	// errors; the glContext has them and maybe would let us use them?
	failuresMut sync.Mutex
	failures    []error
}

var _ render.RenderQueueInterface = (*failfast)(nil)

func (ff *failfast) checkErrors() {
	ff.failuresMut.Lock()
	fails := ff.failures
	ff.failures = nil
	ff.failuresMut.Unlock()

	if len(fails) > 0 {
		panic(fmt.Errorf("failfast queue checkErrors: %w", errors.Join(fails...)))
	}
}

func (ff *failfast) addError(e error) {
	ff.failuresMut.Lock()
	defer ff.failuresMut.Unlock()
	ff.failures = append(ff.failures, e)
}

func (ff *failfast) Queue(job render.RenderJob) {
	ff.checkErrors()
	ff.RenderQueueInterface.Queue(job)
}

func (ff *failfast) Purge() {
	ff.RenderQueueInterface.Purge()
	ff.checkErrors()
}

func Failfastqueue(q render.RenderQueueInterface) render.RenderQueueInterface {
	ret := &failfast{
		RenderQueueInterface: q,
	}
	ret.AddErrorCallback(func(q render.RenderQueueInterface, e error) {
		ret.addError(e)
	})
	return ret
}

func newGlContextForTest(width, height int) *glContext {
	sys, windowHandle, renderQueue := newGlWindowForTest(width, height)
	ctx := &glContext{
		sys:          sys,
		windowHandle: windowHandle,
		render:       Failfastqueue(renderQueue),
		// Note: recordedFailures should be considered local to the render thread;
		// reading/writing to it must synchronize accordingly.
		recordedFailures: nil,
	}
	renderQueue.AddErrorCallback(func(q render.RenderQueueInterface, e error) {
		ctx.recordedFailures = append(ctx.recordedFailures, e)
	})
	return ctx
}

var glTestContextSource = make(chan *glContext, 24)

func runOverCachedContext(width, height int, dotest func(int, int, *glContext)) {
	var theContext *glContext
	select {
	case cachedContext := <-glTestContextSource:
		theContext = cachedContext
	default:
		theContext = newGlContextForTest(width, height)
	}
	defer func() {
		glTestContextSource <- theContext
	}()

	dotest(width, height, theContext)
}

func RunDeprecatedTestWithCachedContext(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	dotest := makeTestTemplate(InvariantsCheckNo, fn)
	runOverCachedContext(width, height, dotest)
}

func RunTestWithCachedContext(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	dotest := makeTestTemplate(InvariantsCheckYes, fn)
	runOverCachedContext(width, height, dotest)
}
