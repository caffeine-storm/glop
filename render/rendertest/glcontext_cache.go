package rendertest

import (
	"errors"
	"fmt"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func RunTestWithCachedContext(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	ctx, returnToCache := getContextFromCache(width, height)
	defer returnToCache()

	e := ctx.prep(width, height)
	if e != nil {
		// Even on error cases, we shouldn't leak GL state.
		panic(errors.Join(fmt.Errorf("couldn't prep: %w", e), ctx.clean()))
	}

	e = ctx.run(fn)

	// Whether the test failed or not, we need to clean()
	ee := ctx.clean()
	e = errors.Join(e, ee)

	if e != nil {
		halting := &conveyIsHalting{}
		if errors.As(e, &halting) {
			// It might be that Convey is trying to halt the tests; we need to
			// preserve their semantics in that case.

			// It may be that ctx.clean() also reported an error. Unless we log it
			// here, it will not be reported.
			glog.ErrorLogger().Error("cleaning also failed", "cleanerror", ee)

			panic(halting.s)
		}

		// Report any test or cleaning errors
		panic(e)
	}
}

var glTestContextSource = make(chan *glContext, 24)

func getContextFromCache(width, height int) (*glContext, func()) {
	var theContext *glContext
	select {
	case cachedContext := <-glTestContextSource:
		theContext = cachedContext
	default:
		theContext = newGlContextForTest(width, height)
	}

	return theContext, func() {
		if theContext.render.IsDefunct() {
			return
		}
		glTestContextSource <- theContext
	}
}
