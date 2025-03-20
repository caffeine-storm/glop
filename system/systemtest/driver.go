package systemtest

import "fmt"

type Driver interface {
	Click(wx, wy int)
	ProcessFrame()
}

type testDriver struct {
	window *testWindow
}

func (d *testDriver) Click(wx, wy int) {
	// Run 'xdotool click $wx $wy'
	RunXDoTool("mousemove", fmt.Sprintf("%d", wx), fmt.Sprintf("%d", wy))
	RunXDoTool("click", "1")
}

func (d *testDriver) ProcessFrame() {
	d.window.sys.Think()
}

var _ Driver = (*testDriver)(nil)

func WithTestWindowDriver(dx, dy int, fn func(driver Driver)) {
	WithTestWindow(dx, dy, func(window Window) {
		fn(window.NewDriver())
	})
}
