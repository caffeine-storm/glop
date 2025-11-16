package guitest

import (
	"fmt"

	"github.com/caffeine-storm/glop/gin"
	"github.com/caffeine-storm/glop/gui"
)

type stubEventDispatcher struct{}

func (*stubEventDispatcher) RegisterEventListener(gin.Listener) {}

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
