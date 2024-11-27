package guitest

import (
	"fmt"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
)

type stubEventDispatcher struct{}

func (*stubEventDispatcher) RegisterEventListener(gin.Listener) {}

func MakeStubbedEventDispatcher() gin.EventDispatcher {
	return &stubEventDispatcher{}
}

func MakeStubbedGui(dims gui.Dims) *gui.Gui {
	ret, err := gui.Make(MakeStubbedEventDispatcher(), dims)
	if err != nil {
		panic(fmt.Errorf("couldn't gui.Make: %w", err))
	}
	return ret
}
