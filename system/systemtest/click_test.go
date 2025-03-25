package systemtest_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system/systemtest"
)

func TestE2EClickHelper(t *testing.T) {
	systemtest.WithTestWindowDriver(64, 64, func(driver systemtest.Driver) {
		expectedX, expectedY := 10, 42
		mouseEvents := []gin.MouseEvent{}

		driver.AddMouseListener(func(mouseEvent gin.MouseEvent) {
			mouseEvents = append(mouseEvents, mouseEvent)
		})

		driver.Click(expectedX, expectedY)

		// TODO(tmckee:#19): this might be clicking some other window or not
		// waiting long enough for the native code to pull the move/click events
		// from X. Perhaps we need to send the click to a specific window by id?
		driver.ProcessFrame()

		// Check that gin saw it.
		if len(mouseEvents) == 0 {
			t.Fatalf("didn't see a click!")
		}

		lastEvent := mouseEvents[len(mouseEvents)-1]
		actualX, actualY := lastEvent.GetPosition()

		if actualX != expectedX || actualY != expectedY {
			t.Fatalf("click co-ordinates didn't match! expected: %s actual %s",
				fmt.Sprintf("(%d, %d)", expectedX, expectedY),
				fmt.Sprintf("(%d, %d)", actualX, actualY))
		}
	})
}
