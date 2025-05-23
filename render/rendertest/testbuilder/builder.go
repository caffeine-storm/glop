package testbuilder

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
)

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

func (b *GlTestBuilder) WithExpectation(C C, ref rendertest.TestDataReference) *expectationGlTestBuilder {
	return &expectationGlTestBuilder{
		ctx:           b,
		expectation:   ref,
		conveyContext: C,
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

func (b *GlTestBuilder) RunWithAllTheThings(fn func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface)) {
	dx, dy := b.Dx, b.Dy
	if dx == 0 || dy == 0 {
		// Pick a default of 64x64
		dx = 64
		dy = 64
	}

	rendertest.RunTestWithCachedContext(dx, dy, func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		fn(sys, hdl, queue)
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

	rendertest.RunTestWithCachedContext(int(dx), int(dy), func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		fn(queue)
	})
}

func (b *queueGlTestBuilder) WithSize(dx, dy int) *queueGlTestBuilder {
	b.ctx.Dx = dx
	b.ctx.Dy = dy
	return b
}

type expectationGlTestBuilder struct {
	ctx           *GlTestBuilder
	expectation   rendertest.TestDataReference
	conveyContext C
}

func (b *expectationGlTestBuilder) Run(fn func()) {
	b.ctx.WithQueue().Run(func(queue render.RenderQueueInterface) {
		queue.Queue(func(st render.RenderQueueState) {
			fn()
		})
		queue.Purge()

		b.conveyContext.So(queue, rendertest.ShouldLookLikeFile, b.expectation)
	})
}

// TODO(tmckee:#38): take a pointer-to-testing.T so that we can properly attribute
// on-render-thread failures with the test that was running.
func New() *GlTestBuilder {
	return &GlTestBuilder{}
}
