package gui_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
)

type stubEventDispatcher struct{}

func (*stubEventDispatcher) RegisterEventListener(gin.Listener) {}

func MakeStubbedEventDispatcher() gin.EventDispatcher {
	return &stubEventDispatcher{}
}

func MakeStubbedGui() *gui.Gui {
	ret, err := gui.Make(MakeStubbedEventDispatcher(), gui.Dims{42, 13})
	if err != nil {
		panic(fmt.Errorf("couldn't gui.Make: %w", err))
	}
	return ret
}

func TestGui(t *testing.T) {
	t.Run("Make", func(t *testing.T) {
		_ = MakeStubbedGui()
	})
}
