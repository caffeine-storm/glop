package rendertest

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

func runTestWithCachedContext(width, height int, checkinvariants bool, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	select {
	case cachedContext := <-glTestContextSource:
		cachedContext.prep(width, height, checkinvariants)
		cachedContext.run(fn)
		cachedContext.clean(checkinvariants)
		glTestContextSource <- cachedContext
	default:
		newContext := newGlContextForTest(width, height)
		newContext.prep(width, height, checkinvariants)
		newContext.run(fn)
		newContext.clean(checkinvariants)
		glTestContextSource <- newContext
	}
}

type GlTestBuilder struct {
	Dx, Dy int
}

type queueGlTestBuilder struct {
	ctx *GlTestBuilder
}

func (b *GlTestBuilder) WithQueue() *queueGlTestBuilder {
	return &queueGlTestBuilder{
		ctx: b,
	}
}

func (b *GlTestBuilder) Run(fn func()) {
	delegate := &queueGlTestBuilder{
		ctx: b,
	}
	delegate.Run(func(queue render.RenderQueueInterface) {
		queue.Queue(func(render.RenderQueueState) {
			fn()
		})
		queue.Purge()
	})
}

func (b *GlTestBuilder) WithSize(dx, dy int) *GlTestBuilder {
	b.Dx = dx
	b.Dy = dy
	return b
}

func (b *queueGlTestBuilder) Run(fn func(render.RenderQueueInterface)) {
	dx, dy := b.ctx.Dx, b.ctx.Dy
	if dx == 0 || dy == 0 {
		// Pick a default of 64x64
		dx = 64
		dy = 64
	}

	runTestWithCachedContext(int(dx), int(dy), InvariantsCheckYes, func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		fn(queue)
	})
}

func (b *queueGlTestBuilder) WithSize(dx, dy int) *queueGlTestBuilder {
	b.ctx.Dx = dx
	b.ctx.Dy = dy
	return b
}

func GlTest() *GlTestBuilder {
	return &GlTestBuilder{}
}
