package systemtest

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/system"
)

type Driver interface {
	Click(wx, wy int)
	ProcessFrame()

	// Put the top-left extent of the window at (x, y) in glop-coords.
	PositionWindow(x, y int)
	AddMouseListener(func(gin.MouseEvent))
	AddInputListener(gin.Listener)

	RawTool(func(system.NativeWindowHandle) []any)
}

type testDriver struct {
	window *testWindow
}

func (d *testDriver) Click(wx, wy int) {
	// Run 'xdotool click $wx $wy'
	glog.DebugLogger().Debug("testDriver.Click>move", "wx", wx, "wy", wy, "self", d)
	xDoToolRun("mousemove", "--window", d.window.hdl, "--sync", wx, wy)
	glog.DebugLogger().Debug("testDriver.Click>click", "wx", wx, "wy", wy, "self", d)
	xDoToolRun("click", "--window", d.window.hdl, "1")
	glog.DebugLogger().Debug("testDriver.Click>done", "wx", wx, "wy", wy, "self", d)
}

func (d *testDriver) RawTool(fn func(system.NativeWindowHandle) []any) {
	xDoToolRun(fn(d.window.hdl)...)
}

func (d *testDriver) ProcessFrame() {
	d.window.sys.Think()
}

func (d *testDriver) PositionWindow(x, y int) {
	xDoToolRun("windowmove", d.window.hdl, x, y)
}

func (d *testDriver) AddMouseListener(listener func(gin.MouseEvent)) {
	d.window.sys.AddMouseListener(listener)
}

func (d *testDriver) AddInputListener(listnr gin.Listener) {
	d.window.sys.AddInputListener(listnr)
}

var _ Driver = (*testDriver)(nil)

func WithTestWindowDriver(dx, dy int, fn func(driver Driver)) {
	WithTestWindow(dx, dy, func(window Window) {
		fn(window.NewDriver())
	})
}
