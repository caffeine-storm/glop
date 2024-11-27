package gui_test

import (
	"testing"

	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/gui/guitest"
)

func TestGui(t *testing.T) {
	t.Run("Make", func(t *testing.T) {
		_ = guitest.MakeStubbedGui(gui.Dims{13, 42})
	})
}
