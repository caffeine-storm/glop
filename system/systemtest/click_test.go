package systemtest_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/system/systemtest"
)

func TestE2EClickHelper(t *testing.T) {
	systemtest.WithTestWindowDriver(64, 64, func(driver systemtest.Driver) {
		expectedX, expectedY := 10, 42

		driver.Click(expectedX, expectedY)
		driver.ProcessFrame()

		// Check that gin saw it.
		allEvents := driver.GetEvents()
		if len(allEvents) == 0 {
			t.Fatalf("didn't see any events!")
		}

		actualX, actualY := driver.GetLastClick()

		if actualX != expectedX || actualY != expectedY {
			t.Fatalf("click co-ordinates didn't match! expected: %s actual %s",
				fmt.Sprintf("(%d, %d)", expectedX, expectedY),
				fmt.Sprintf("(%d, %d)", actualX, actualY))
		}
	})
}
