package gin_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runningwild/glop/gin"
)

func TestEventGroup(t *testing.T) {
	t.Run("event groups have optional x-y co-ordinates", func(t *testing.T) {
		eg := gin.EventGroup{}
		if eg.HasMousePosition() {
			t.Fatalf("zero-valued groups must have non-existent co-ordinates")
		}

		eg.SetMousePosition(14, 44)

		if !eg.HasMousePosition() {
			t.Fatalf("there must be a position after setting one")
		}

		x, y := eg.GetMousePosition()
		if x != 14 || y != 44 {
			t.Fatalf("reading the mouse position should return whatever was set")
		}
	})
	t.Run("events have a useful API", func(t *testing.T) {
		ev := gin.Event{}
		ev.IsPress()
		ev.IsRelease()
	})
	t.Run("EventGroup.String() should be useful", func(t *testing.T) {
		eg := gin.EventGroup{}
		eg.SetMousePosition(14, 44)

		stringified := fmt.Sprintf("%v", eg)

		if !strings.Contains(stringified, "14 44") {
			t.Fatalf("the event group should show the mouse x/y when stringified but got %q", stringified)
		}
	})
}
