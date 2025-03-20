package systemtest

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

type Window interface {
	NewDriver() Driver
}

type testWindow struct {
	sys system.System
}

func (self *testWindow) NewDriver() Driver {
	return &testDriver{
		window: self,
	}
}

var _ Window = (*testWindow)(nil)

func NewTestWindow(sys system.System, queue render.RenderQueueInterface) Window {
	return &testWindow{
		sys: sys,
	}
}

func WithTestWindow(dx, dy int, fn func(window Window)) {
	rendertest.WithGlForTest(dx, dy, func(sys system.System, queue render.RenderQueueInterface) {
		window := NewTestWindow(sys, queue)
		queue.Queue(func(st render.RenderQueueState) {
			fn(window)
		})
		queue.Purge()
	})
}
