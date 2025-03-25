package systemtest

type Driver interface {
	Click(wx, wy int)
	ProcessFrame()
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

var _ Driver = (*testDriver)(nil)

func WithTestWindowDriver(dx, dy int, fn func(driver Driver)) {
	WithTestWindow(dx, dy, func(window Window) {
		fn(window.NewDriver())
	})
}
