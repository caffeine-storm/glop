package testbuilder

import (
	"fmt"

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

func (b *GlTestBuilder) RunForQueueState(fn func(render.RenderQueueState)) {
	b.WithQueue().Run(func(queue render.RenderQueueInterface) {
		queue.Queue(fn)
		queue.Purge()
	})
}

func (b *GlTestBuilder) WithExpectation(C C, ref rendertest.TestDataReference, args ...any) *expectationGlTestBuilder {
	bgColour := rendertest.DefaultBackground
	for _, arg := range args {
		switch v := arg.(type) {
		case rendertest.BackgroundColour:
			bgColour = v
		default:
			panic(fmt.Errorf("unexpected trailing arg type: %T", arg))
		}
	}

	return &expectationGlTestBuilder{
		ctx:           b,
		expectation:   ref,
		bgColour:      bgColour,
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
	bgColour      rendertest.BackgroundColour
	conveyContext C
}

func (b *expectationGlTestBuilder) Run(fn func()) {
	b.RunForQueueState(func(render.RenderQueueState) {
		fn()
	})
}

func (b *expectationGlTestBuilder) RunForQueueState(fn func(render.RenderQueueState)) {
	b.ctx.WithQueue().Run(func(queue render.RenderQueueInterface) {
		queue.Queue(func(st render.RenderQueueState) {
			fn(st)
		})
		queue.Purge()

		b.conveyContext.So(queue, rendertest.ShouldLookLikeFile, b.expectation, b.bgColour)
	})
}

func (b *expectationGlTestBuilder) RunWithAllTheThings(fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	b.ctx.RunWithAllTheThings(func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		fn(sys, hdl, queue)
		b.conveyContext.So(queue, rendertest.ShouldLookLikeFile, b.expectation, b.bgColour)
	})
}

type expectationQueueGlTestBuilder struct {
	ctx *expectationGlTestBuilder
}

func (b *expectationGlTestBuilder) WithQueue() *expectationQueueGlTestBuilder {
	return &expectationQueueGlTestBuilder{
		ctx: b,
	}
}

func (b *expectationQueueGlTestBuilder) Run(fn func(render.RenderQueueInterface)) {
	b.ctx.ctx.WithQueue().Run(func(queue render.RenderQueueInterface) {
		fn(queue)
		queue.Purge()

		b.ctx.conveyContext.So(queue, rendertest.ShouldLookLikeFile, b.ctx.expectation)
	})
}

func New() *GlTestBuilder {
	return &GlTestBuilder{}
}

func Run(ffn any) {
	it := New()
	dorun(it.Run, it.WithQueue().Run, it.RunForQueueState, it.RunWithAllTheThings, ffn)
}

func WithSize(dx, dy int, ffn any) {
	it := New().WithSize(dx, dy)
	dorun(it.Run, it.WithQueue().Run, it.RunForQueueState, it.RunWithAllTheThings, ffn)
}

func WithExpectation(c C, ref rendertest.TestDataReference, arg0 any, args ...any) {
	args = append([]any{arg0}, args...)
	ffn := args[len(args)-1]
	args = args[:len(args)-1]

	it := New().WithExpectation(c, ref, args...)

	dorun(it.Run, it.WithQueue().Run, it.RunForQueueState, it.RunWithAllTheThings, ffn)
}

func WithSizeAndExpectation(dx, dy int, c C, ref rendertest.TestDataReference, arg0 any, args ...any) {
	args = append([]any{arg0}, args...)
	ffn := args[len(args)-1]
	args = args[:len(args)-1]

	it := New().WithSize(dx, dy).WithExpectation(c, ref, args...)

	dorun(it.Run, it.WithQueue().Run, it.RunForQueueState, it.RunWithAllTheThings, ffn)
}

func dorun(f1 func(func()), f2 func(func(render.RenderQueueInterface)), f3 func(func(render.RenderQueueState)), f4 func(func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)), ffn any) {
	switch fn := ffn.(type) {
	case func():
		f1(fn)
	case func(render.RenderQueueInterface):
		f2(fn)
	case func(render.RenderQueueState):
		f3(fn)
	case func(system.System, system.NativeWindowHandle, render.RenderQueueInterface):
		f4(fn)
	default:
		panic(fmt.Errorf("unknown test implmentation type: %T", ffn))
	}
}
