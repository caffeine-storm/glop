package systemtest

import "github.com/runningwild/glop/gin"

type Driver interface {
	Click(wx, wy int)
	ProcessFrame()
	AddMouseListener(func(gin.MouseEvent))
}

type testDriver struct {
	window *testWindow
}

func (d *testDriver) Click(wx, wy int) {
	// Run 'xdotool click $wx $wy'
	runXDoTool("mousemove", wx, wy)
	runXDoTool("click", "1")
}

func (d *testDriver) ProcessFrame() {
	d.window.sys.Think()
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
