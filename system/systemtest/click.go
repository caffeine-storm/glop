package systemtest

import (
	"fmt"
	"os/exec"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

type Driver interface {
	Click(wx, wy int)
	ProcessFrame()
}

type testDriver struct {
	wdw *testWindow
}

type Window interface {
	NewDriver() Driver
}

type testWindow struct {
	sys system.System
}

func (self *testWindow) NewDriver() Driver {
	return &testDriver{
		wdw: self,
	}
}

var _ Window = (*testWindow)(nil)

func RunXDoTool(xdotoolArgs ...string) {
	cmd := exec.Command("xdotool", xdotoolArgs...)

	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("could not %q: %w", cmd.String(), err))
	}
}

func (d *testDriver) Click(wx, wy int) {
	// Run 'xdotool click $wx $wy'
	RunXDoTool("mousemove", fmt.Sprintf("%d", wx), fmt.Sprintf("%d", wy))
	RunXDoTool("click", "1")
}

func (d *testDriver) ProcessFrame() {
	d.wdw.sys.Think()
}

var _ Driver = (*testDriver)(nil)

func NewTestWindow(sys system.System, queue render.RenderQueueInterface) Window {
	return &testWindow{
		sys: sys,
	}
}
