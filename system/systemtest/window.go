package systemtest

import (
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

type Window interface {
	NewDriver() Driver
	GetQueue() render.RenderQueueInterface
	GetDims() gui.Dims
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

func (self *testWindow) GetDims() gui.Dims {
	_, _, dx, dy := self.sys.GetWindowDims()
	return gui.Dims{
		Dx: dx,
		Dy: dy,
	}
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
	rendertest.DeprecatedWithGlAndHandleForTest(dx, dy, func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
		window := NewTestWindow(sys, hdl, queue)
		queue.Queue(func(st render.RenderQueueState) {
			fn(window)
		})
		queue.Purge()
	})
}
