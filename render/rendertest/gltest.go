package rendertest

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func RunTestWithCachedContext(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	dotest := func(ctx *glContext) {
		e := ctx.prep(width, height, InvariantsCheckYes)
		if e != nil {
			// Even on error cases, we shouldn't leak GL state.
			ctx.clean(InvariantsCheckNo)
			panic(e)
		}

		ctx.run(fn)

		e = ctx.clean(InvariantsCheckYes)
		if e != nil {
			panic(e)
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
