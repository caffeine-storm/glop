package systemtest

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/runningwild/glop/system"
)

type testBuilder struct {
	delegate *testbuilder.GlTestBuilder
}

func (b *testBuilder) Run(fn func(window Window)) {
	b.delegate.RunWithAllTheThings(func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		window := NewTestWindow(sys, hdl, queue)
		queue.Queue(func(st render.RenderQueueState) {
			fn(window)
		})
		queue.Purge()
	})
}

func TestBuilder(delegate *testbuilder.GlTestBuilder) *testBuilder {
	return &testBuilder{
		delegate: delegate,
	}
}
