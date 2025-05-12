package rendertest

import (
	"fmt"

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
