package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system"
	"github.com/runningwild/glop/system/systemtest"
)

type click struct {
	x, y int
}

const windowScale = 64

func GivenANewDriver() (systemtest.Driver, func()) {
	driverChannel := make(chan systemtest.Driver)

	go systemtest.WithTestWindowDriver(windowScale, windowScale, func(driver systemtest.Driver) {
		driverChannel <- driver
		<-driverChannel
	})

	drv := <-driverChannel
	cleanup := func() {
		driverChannel <- drv
	}
	return drv, cleanup
}

func WatchForMouseEvents(drv systemtest.Driver) *[]gin.MouseEvent {
	ret := new([]gin.MouseEvent)

	drv.AddMouseListener(func(evt gin.MouseEvent) {
		*ret = append(*ret, evt)
	})

	return ret
}

func LastClick(events *[]gin.MouseEvent) (click, bool) {
	for i := len(*events) - 1; i >= 0; i-- {
		evt := (*events)[i]
		if evt.Type == gin.MouseEventTypeClick {
			return click{
				x: evt.X,
				y: evt.Y,
			}, true
		}
	}

	return click{}, false
}

func TestSystemtestDriver(t *testing.T) {
	t.Run("xdotool commands are sent to the correct native window", func(t *testing.T) {
		// Make two windows of the same size.
		driverA, cleanA := GivenANewDriver()
		defer cleanA()

		driverB, cleanB := GivenANewDriver()
		defer cleanB()

		// Move each window to the same screen position.
		driverA.PositionWindow(12, 17)
		driverB.PositionWindow(12, 17)
		driverA.ProcessFrame()
		driverB.ProcessFrame()

		mouseEventsA := WatchForMouseEvents(driverA)
		mouseEventsB := WatchForMouseEvents(driverB)

		// Click on two points within the windows' shared bounds and process their
		// events. Assert that each window sees only the click sent to it.
		// Note that we need to click 'above' 17 or else we're clicking off of the
		// screen.
		driverA.Click(4, 22)
		driverA.ProcessFrame()
		driverB.Click(9, 19)
		driverB.ProcessFrame()

		// Assert each window got their click.
		expectedClickA := click{
			x: 4,
			y: 22,
		}
		expectedClickB := click{
			x: 9,
			y: 19,
		}
		clickA, found := LastClick(mouseEventsA)
		if !found {
			t.Error("no events found for driverA")
		}
		clickB, found := LastClick(mouseEventsB)
		if !found {
			t.Error("no events found for driverB")
		}

		if clickA != expectedClickA || clickB != expectedClickB {
			t.Fatalf("click expectations failed: aclicks: %+v, bclicks: %+v", mouseEventsA, mouseEventsB)
		}
	})

	t.Run("no ordering constraints between separate drivers' ProcessFrame() calls", func(t *testing.T) {
		// Make two windows of the same size.
		driverA, cleanA := GivenANewDriver()
		defer cleanA()

		driverB, cleanB := GivenANewDriver()
		defer cleanB()

		// Move each window to the same screen position.
		driverA.PositionWindow(12, 17)
		driverB.PositionWindow(12, 17)
		driverA.ProcessFrame()
		driverB.ProcessFrame()

		mouseEventsA := WatchForMouseEvents(driverA)
		mouseEventsB := WatchForMouseEvents(driverB)

		// Click on two points within the windows' shared bounds and process their
		// events. Assert that each window sees only the click sent to it. Do both
		// clicks then both ProcessFrame calls for a regression test.
		// Note that we need to click 'above' 17 or else we're clicking off of the
		// screen.
		driverA.Click(4, 22)
		driverB.Click(9, 19)
		driverA.ProcessFrame()
		driverB.ProcessFrame()

		// Assert each window got their click.
		expectedClickA := click{
			x: 4,
			y: 22,
		}
		expectedClickB := click{
			x: 9,
			y: 19,
		}
		clickA, found := LastClick(mouseEventsA)
		if !found {
			t.Error("no events found for driverA")
		}
		clickB, found := LastClick(mouseEventsB)
		if !found {
			t.Error("no events found for driverB")
		}

		if clickA != expectedClickA || clickB != expectedClickB {
			t.Fatalf("click expectations failed: aclicks: %+v, bclicks: %+v", mouseEventsA, mouseEventsB)
		}
	})

	t.Run("clicks use an origin at the bottom-left", func(t *testing.T) {
		driver, clean := GivenANewDriver()
		defer clean()

		events := WatchForMouseEvents(driver)

		// Click lower half of the test window. Use RawTool for specifying polar
		// co-ordinates that won't suffer from our glop-vs-X origin confusion.
		driver.RawTool(func(hdl system.NativeWindowHandle) []any {
			return []any{
				// Move mouse to target; 3 pixels below the bottom of the centre of the
				// window under test. X will consider the centre to be at (32, 32) for
				// a 64x64 window.
				"mousemove", "--sync", "--window", hdl, "--polar", 180, 3,
				// Click the left mouse button
				"click", "--window", hdl, 1,
			}
		})
		driver.ProcessFrame()

		// Expect a report of (floor((width-1)/2), floor((height-1)/2) - 3).
		yexpect := ((windowScale - 1) / 2) - 3
		expectedClick := click{
			x: windowScale / 2,
			y: yexpect,
		}

		lastClick, found := LastClick(events)
		if !found {
			t.Fatalf("couldn't find the click in %v", events)
		}

		if lastClick != expectedClick {
			t.Fatalf("click expectation failed: expected: %+v, actual: %+v", expectedClick, lastClick)
		}
	})
}
