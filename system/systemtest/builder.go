package systemtest

import (
	"github.com/caffeine-storm/glop/render"
	"github.com/caffeine-storm/glop/render/rendertest/testbuilder"
	"github.com/caffeine-storm/glop/system"
)

type testBuilder struct {
	delegate *testbuilder.GlTestBuilder
}

func (b *testBuilder) Run(fn func(window Window)) {
	b.delegate.RunWithAllTheThings(func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		window := NewTestWindow(sys, hdl, queue)
		fn(window)
	})
}

func TestBuilder(delegate *testbuilder.GlTestBuilder) *testBuilder {
	return &testBuilder{
		delegate: delegate,
	}
}
