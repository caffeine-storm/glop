package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system/systemtest"
)

type click struct {
	x, y int
}

func GivenANewDriver() (systemtest.Driver, func()) {
	driverChannel := make(chan systemtest.Driver)

	go systemtest.WithTestWindowDriver(64, 64, func(driver systemtest.Driver) {
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
	if len(*events) == 0 {
		return click{}, false
	}

	evt := (*events)[len(*events)-1]

	return click{
		x: evt.X,
		y: evt.Y,
	}, true
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
		// TODO(tmckee): it seems clickA => clickB => procA => procB fails... it
		// should not fail.
		driverA.Click(4, 5)
		driverA.ProcessFrame()
		driverB.Click(9, 2)
		driverB.ProcessFrame()

		// Assert each window got their click.
		expectedClickA := click{
			x: 4,
			y: 5,
		}
		expectedClickB := click{
			x: 9,
			y: 2,
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
}
