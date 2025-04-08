package systemtest

import (
	"fmt"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system"

	agg "github.com/runningwild/glop/gin/aggregator"
)

type Driver interface {
	Click(wx, wy int)
	Scroll(dy float64)
	ProcessFrame()

	// Put the top-left extent of the window at (x, y) in glop-coords.
	PositionWindow(x, y int)
	AddInputListener(gin.Listener)

	RawTool(func(system.NativeWindowHandle) []any)

	GetEvents() []gin.EventGroup

	// Panics if there were no clicks
	GetLastClick() (int, int)

	// Panics if there were no scrolls
	GetLastScroll() float64

	gin.Listener
}

type testDriver struct {
	window *testWindow

	// Each testDriver listens for input events from gin and records each event
	// group here.
	eventGroups []gin.EventGroup
}

func (d *testDriver) glopToX(glopX, glopY int) (int, int) {
	height := d.window.getWindowHeight()
	return glopX, height - 1 - glopY
}

func (d *testDriver) Click(glopX, glopY int) {
	xorgX, xorgY := d.glopToX(glopX, glopY)
	xDoToolRun("mousemove", "--window", d.window.hdl, "--sync", xorgX, xorgY)
	xDoToolRun("click", "--window", d.window.hdl, "1")
}

func (d *testDriver) Scroll(dy float64) {
	if dy == 0 {
		panic(fmt.Errorf("can't scroll by a distance of 0"))
	}
	x, y := d.glopToX(5, 5)
	xDoToolRun("mousemove", "--window", d.window.hdl, "--sync", x, y)
	btn := 4 // 'scroll up' is button4 in X parlance
	if dy < 0 {
		btn = 5 // 'scroll down' is button5 in X parlance
	}
	xDoToolRun("click", "--window", d.window.hdl, btn)
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

func (d *testDriver) HandleEventGroup(grp gin.EventGroup) {
	d.eventGroups = append(d.eventGroups, grp)
}

func (d *testDriver) GetLastClick() (int, int) {
	for i := len(d.eventGroups) - 1; i > 0; i-- {
		each := d.eventGroups[i]
		if !each.HasMousePosition() {
			continue
		}
		switch each.PrimaryEvent().Key.Id().Index {
		case gin.MouseLButton:
			fallthrough
		case gin.MouseMButton:
			fallthrough
		case gin.MouseRButton:
			return each.GetMousePosition()
		}
	}

	panic(fmt.Errorf("couldn't find click in events: %v", d.eventGroups))
}

func (d *testDriver) GetLastScroll() float64 {
	for i := len(d.eventGroups) - 1; i > 0; i-- {
		each := d.eventGroups[i]
		if !each.HasMousePosition() {
			continue
		}
		ev := each.PrimaryEvent()
		if ev.Type != agg.Press {
			continue
		}

		switch ev.Key.Id().Index {
		case gin.MouseWheelVertical:
			// TODO: this counts number of scrolls; not direction/distance of
			// scroll...
			return float64(ev.Key.FramePressCount())
		}
	}

	panic(fmt.Errorf("couldn't find MouseWheelVertical in events: %v", d.eventGroups))
}

func (d *testDriver) GetEvents() []gin.EventGroup {
	return d.eventGroups
}

func (d *testDriver) Think(int64) {}

func (d *testDriver) AddInputListener(l gin.Listener) {
	d.window.sys.AddInputListener(l)
}

var _ Driver = (*testDriver)(nil)

func WithTestWindowDriver(dx, dy int, fn func(driver Driver)) {
	WithTestWindow(dx, dy, func(window Window) {
		fn(window.NewDriver())
	})
}
