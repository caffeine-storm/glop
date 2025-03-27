package systemtest

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
)

type Driver interface {
	Click(wx, wy int)
	ProcessFrame()
	PositionWindow(x, y int)
	AddMouseListener(func(gin.MouseEvent))
}

type testDriver struct {
	window *testWindow
}

func (d *testDriver) Click(wx, wy int) {
	// Run 'xdotool click $wx $wy'
	glog.DebugLogger().Debug("testDriver.Click>move", "wx", wx, "wy", wy, "self", d)
	runXDoTool("mousemove", "--window", d.window.hdl, "--sync", wx, wy)
	glog.DebugLogger().Debug("testDriver.Click>click", "wx", wx, "wy", wy, "self", d)
	runXDoTool("click", "--window", d.window.hdl, "1")
	glog.DebugLogger().Debug("testDriver.Click>done", "wx", wx, "wy", wy, "self", d)
}

func (d *testDriver) ProcessFrame() {
	d.window.sys.Think()
}

func (d *testDriver) PositionWindow(x, y int) {
	runXDoTool("windowmove", d.window.hdl, x, y)
}

func (d *testDriver) AddMouseListener(listener func(gin.MouseEvent)) {
	d.window.sys.AddMouseListener(listener)
}

var _ Driver = (*testDriver)(nil)

func WithTestWindowDriver(dx, dy int, fn func(driver Driver)) {
	WithTestWindow(dx, dy, func(window Window) {
		fn(window.NewDriver())
	})
}
