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

// TODO: rename this file to mouse_test.go
func TestE2EMouseWheelHelper(t *testing.T) {
	testcases := []struct {
		name           string
		expectedScroll int
	}{
		{
			name:           "scroll up",
			expectedScroll: 7,
		},
		{
			name:           "scroll down",
			expectedScroll: -10,
		},
	}
	for _, testcase := range testcases {
		systemtest.WithTestWindowDriver(64, 64, func(driver systemtest.Driver) {
			driver.Scroll(testcase.expectedScroll)
			driver.ProcessFrame()

			// Check that gin saw it.
			allEvents := driver.GetEvents()
			if len(allEvents) == 0 {
				t.Fatalf("didn't see any events!")
			}

			actualScroll := driver.GetLastScroll()

			if actualScroll != testcase.expectedScroll {
				t.Fatalf("scroll amount didn't match! expected: %d actual %d", testcase.expectedScroll, actualScroll)
			}
		})
	}
}
