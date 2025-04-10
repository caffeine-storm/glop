package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system/systemtest"
)

func TestWindowExposesAQueue(t *testing.T) {
	systemtest.WithTestWindow(64, 64, func(window systemtest.Window) {
		var _ render.RenderQueueInterface = window.GetQueue()
	})
}

func TestWindowExposesDimensions(t *testing.T) {
	systemtest.WithTestWindow(64, 64, func(window systemtest.Window) {
		var _ gui.Dims = window.GetDims()
	})
}
