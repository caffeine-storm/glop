package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system/systemtest"
)

type click struct {
	x, y int
}

func GivenANewDriver() chan systemtest.Driver {
	driverChannel := make(chan systemtest.Driver)

	go systemtest.WithTestWindowDriver(64, 64, func(driver systemtest.Driver) {
		driverChannel <- driver
		<-driverChannel
	})

	return driverChannel
}

func WatchForClicks(drv systemtest.Driver) *[]gin.MouseEvent {
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
		driverAChan := GivenANewDriver()
		driverA := <-driverAChan
		defer func() {
			driverAChan <- driverA
		}()
		driverBChan := GivenANewDriver()
		driverB := <-driverBChan
		defer func() {
			driverBChan <- driverB
		}()

		// Move each window to the same screen position.
		driverA.PositionWindow(12, 17)
		driverB.PositionWindow(12, 17)
		driverA.ProcessFrame()
		driverB.ProcessFrame()

		aclicks := WatchForClicks(driverA)
		bclicks := WatchForClicks(driverB)

		// Click on two points within the windows' shared bounds and process their
		// events. Assert that each window sees only the click sent to it.
		driverA.Click(16, 22)
		driverB.Click(21, 19)
		driverA.ProcessFrame()
		driverB.ProcessFrame()

		// Assert each window got their click.
		expectedClickA := click{
			x: 16,
			y: 22,
		}
		expectedClickB := click{
			x: 21,
			y: 19,
		}
		clickA, found := LastClick(aclicks)
		if !found {
			t.Error("no events found for driverA")
		}
		clickB, found := LastClick(bclicks)
		if !found {
			t.Error("no events found for driverB")
		}

		if clickA != expectedClickA || clickB != expectedClickB {
			t.Fatalf("click expectations failed: aclicks: %+v, bclicks: %+v", aclicks, bclicks)
		}
	})
}
