package rendertest

import (
	"fmt"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func RunTestWithCachedContext(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	dotest := func(ctx *glContext) {
		e := ctx.prep(width, height, InvariantsCheckYes)
		if e != nil {
			// Even on error cases, we shouldn't leak GL state.
			ctx.clean(InvariantsCheckNo)
			panic(fmt.Errorf("previous state leakage encountered during prep: %w", e))
		}

		e = ctx.run(fn)

		ee := ctx.clean(InvariantsCheckYes)
		if ee != nil {
			panic(fmt.Errorf("state leakage during cleanup: %w", e))
		}

		if e != nil {
			panic(fmt.Errorf("error on render-thread: %w", e))
		}
	}

	var theContext *glContext
	select {
	case cachedContext := <-glTestContextSource:
		theContext = cachedContext
	default:
		theContext = newGlContextForTest(width, height)
	}

	dotest(theContext)

	glTestContextSource <- theContext
}
