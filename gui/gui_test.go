package gui_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
)

type stubEventDispatcher struct{}

func (*stubEventDispatcher) RegisterEventListener(gin.Listener) {}

func TestGui(t *testing.T) {
	t.Run("Make", func(t *testing.T) {
		dispatch := &stubEventDispatcher{}
		_, err := gui.Make(dispatch, gui.Dims{42, 13})
		if err != nil {
			t.Fatalf("couldn't gui.Make: %v", err)
		}
	})
}
