package rendertest

import (
	"errors"
	"fmt"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func doTest(checkInvariants bool, width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	ctx, cleanup := getContextFromCache(width, height)
	defer cleanup()

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
		halting := &conveyIsHalting{}
		if errors.As(e, &halting) {
			// It might be that Convey is trying to halt the tests; we need to
			// preserve their semantics in that case.
			panic(halting.s)
		}

		panic(fmt.Errorf("error on render-thread: %w", e))
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

func RunDeprecatedTestWithCachedContext(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	doTest(InvariantsCheckNo, width, height, fn)
}

func RunTestWithCachedContext(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	doTest(InvariantsCheckYes, width, height, fn)
}
