package gui_test

import (
	"bytes"
	"testing"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/gui/guitest"
)

var dims = gui.Dims{13, 42}

func TestGui(t *testing.T) {
	t.Run("Make", func(t *testing.T) {
		_ = guitest.MakeStubbedGui(gui.Dims{13, 42})
	})

	t.Run("make with logger", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger := glog.New(&glog.Opts{
			Output: buffer,
		})

		val, err := gui.MakeLogged(dims, guitest.MakeStubbedEventDispatcher(), logger)
		if err != nil {
			t.Fatalf("got unexpected error while gui.Make'ing: %v", err)
		}

		if val == nil {
			t.Fatalf("got no error from Make but got no gui either!")
		}
	})
}
