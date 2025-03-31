package gin

import (
	"fmt"

	"github.com/runningwild/glop/glog"
)

type MouseEventType int

const (
	MouseEventTypeMove MouseEventType = iota
	MouseEventTypeClick
	MouseEventTypeWheel
)

// TODO(#26): need to distinguish between left/middle/right buttons
type MouseEvent struct {
	X, Y int
	Type MouseEventType
}

func (evt *MouseEvent) GetPosition() (int, int) {
	return evt.X, evt.Y
}

// TODO(#18): other input things generate gin.Event instances at their
// discretion; we might want to change the MouseListenerFunc api to report
// optional gin.Events too
type MouseListenerFunc func(MouseEvent)

// Dispatches to a list of listeners whenever a mouse event happens.
type MouseInput struct {
	listeners []MouseListenerFunc
	logger    glog.Logger
}

func (in *MouseInput) AddListener(listenerFunc MouseListenerFunc) {
	in.listeners = append(in.listeners, listenerFunc)
}

func classifyMouseEventType(event OsEvent) MouseEventType {
	// Note: this assumes a single pointer device; need to distinguish by
	// event.KeyId.Device.Index if we want to handle multiple pointers at once.
	switch event.KeyId.Index {
	case MouseXAxis:
		fallthrough
	case MouseYAxis:
		return MouseEventTypeMove

	case MouseWheelVertical:
		fallthrough
	case MouseWheelHorizontal:
		return MouseEventTypeWheel

	case MouseLButton:
		fallthrough
	case MouseMButton:
		fallthrough
	case MouseRButton:
		return MouseEventTypeClick
	}

	panic(fmt.Errorf("Unmapped KeyId: %v (should have had a mouse-specific KeyId.Index)", event.KeyId))
}

func (in *MouseInput) Handle(event OsEvent, group *EventGroup) {
	if in.logger == nil {
		in.logger = glog.VoidLogger()
	}
	in.logger.Trace("Handling a mouse event", "rawx", event.X, "rawy", event.Y)

	mevt := MouseEvent{
		X:    event.X,
		Y:    event.Y,
		Type: classifyMouseEventType(event),
	}

	for _, fn := range in.listeners {
		fn(mevt)
	}

	// TODO(#18): here is where we could add to 'group.Events' if we decide to
	// support optional MouseListenerFunc 'feedback'.
	// group.events = append(group.events, Event{kinda-thingy})
}

func (in *Input) AddMouseListener(listenerFunc MouseListenerFunc) {
	in.mouse.AddListener(listenerFunc)
}
