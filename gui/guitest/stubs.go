package guitest

import (
	"fmt"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
)

type stubEventDispatcher struct{}

func (*stubEventDispatcher) RegisterEventListener(gin.Listener)     {}
func (*stubEventDispatcher) AddMouseListener(gin.MouseListenerFunc) {}

func MakeStubbedEventDispatcher() gin.EventDispatcher {
	return &stubEventDispatcher{}
}

func MakeStubbedGui(dims gui.Dims) *gui.Gui {
	ret, err := gui.Make(dims, MakeStubbedEventDispatcher())
	if err != nil {
		panic(fmt.Errorf("couldn't gui.Make: %w", err))
	}
	return ret
}
