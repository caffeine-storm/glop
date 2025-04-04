package systemtest

import (
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

type Window interface {
	NewDriver() Driver
	GetQueue() render.RenderQueueInterface
}

type testWindow struct {
	sys   system.System
	hdl   system.NativeWindowHandle
	queue render.RenderQueueInterface
}

func (self *testWindow) NewDriver() Driver {
	result := &testDriver{
		window: self,
	}

	self.sys.AddInputListener(result)

	return result
}

func (self *testWindow) GetQueue() render.RenderQueueInterface {
	return self.queue
}

var _ Window = (*testWindow)(nil)

func (self *testWindow) getWindowHeight() int {
	_, _, _, dy := self.sys.GetWindowDims()
	return dy
}

func NewTestWindow(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) Window {
	return &testWindow{
		sys:   sys,
		hdl:   hdl,
		queue: queue,
	}
}

func WithTestWindow(dx, dy int, fn func(window Window)) {
	rendertest.WithGlAndHandleForTest(dx, dy, func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		window := NewTestWindow(sys, hdl, queue)
		queue.Queue(func(st render.RenderQueueState) {
			fn(window)
		})
		queue.Purge()
	})
}
