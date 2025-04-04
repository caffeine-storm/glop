package gin_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
)

func TestEventGroup(t *testing.T) {
	t.Run("event groups have x-y co-ordinates", func(t *testing.T) {
		eg := gin.EventGroup{}
		eg.X = eg.Y
		if eg.X != eg.Y {
			t.Fatalf("X and Y should behave like fields")
		}
	})
}
