package gin

import (
	"fmt"

	"github.com/runningwild/glop/gin/aggregator"
)

// A view over the data that comes back from native code.
type OsEvent struct {
	KeyId     KeyId
	Press_amt float64
	Timestamp int64
	X, Y      int
}

type Event struct {
	Key  Key
	Type aggregator.EventType
}

func (e Event) String() string {
	if e.Key == nil || e.Type == aggregator.NoEvent {
		return fmt.Sprintf("NoEvent")
	}
	return fmt.Sprintf("'%v %v'", e.Type, e.Key)
}

func (e Event) IsPress() bool {
	return e.Type == aggregator.Press
}

func (e Event) IsRelease() bool {
	return e.Type == aggregator.Release
}

type MousePosition struct {
	X, Y int
}

// An EventGroup is a series of events that were all created by a single
// OsEvent.
type EventGroup struct {
	Events    []Event
	mousePos  *MousePosition
	Timestamp int64
}

func (eg EventGroup) String() string {
	mouseInfo := "<mousePos: nil>"
	if eg.HasMousePosition() {
		x, y := eg.GetMousePosition()
		mouseInfo = fmt.Sprintf("{%d %d}", x, y)
	}

	return fmt.Sprintf("{%v %s %v}", eg.Events, mouseInfo, eg.Timestamp)
}

// Returns a bool indicating whether an event corresponding to the given KeyId
// is present in the EventGroup, and if so the Event returned is a copy of that
// event.
func (eg *EventGroup) FindEvent(id KeyId) (Event, bool) {
	for _, evt := range eg.Events {
		if id.Contains(evt.Key.Id()) {
			return evt, true
		}
	}
	return Event{}, false
}

// Returns true if the given KeyId is considered 'Pressed' within this event
// group.
func (eg *EventGroup) IsPressed(id KeyId) bool {
	ev, found := eg.FindEvent(id)
	if !found {
		return false
	}
	return ev.Type == aggregator.Press
}

// Returns true if the given KeyId is considered 'Released' within this event
// group.
func (eg *EventGroup) IsReleased(id KeyId) bool {
	ev, found := eg.FindEvent(id)
	if !found {
		return false
	}
	return ev.Type == aggregator.Release
}

// Returns the root-cause event of the EventGroup. Useful for classifying the
// group as a whole.
func (eg *EventGroup) PrimaryEvent() Event {
	if len(eg.Events) == 0 {
		panic(fmt.Errorf("no (primary) event for eventgroup"))
	}
	return eg.Events[0]
}

func (eg *EventGroup) HasMousePosition() bool {
	return eg.mousePos != nil
}

func (eg *EventGroup) GetMousePosition() (int, int) {
	if !eg.HasMousePosition() {
		panic(fmt.Errorf("can't GetMousePosition when it's nil"))
	}
	return eg.mousePos.X, eg.mousePos.Y
}

func (eg *EventGroup) SetMousePosition(x, y int) {
	eg.mousePos = &MousePosition{
		X: x,
		Y: y,
	}
}

func (eg *EventGroup) IsMouseMove() bool {
	if !eg.HasMousePosition() {
		return false
	}

	if eg.PrimaryEvent().Key.Id().Device.Type != DeviceTypeMouse {
		return false
	}

	return eg.PrimaryEvent().Type == aggregator.Adjust
}

// During HandleEventGroup a Listener can query keys as to their current state
// (i.e.  with Cur*() methods) and these will accurately report their state.
//
// Frame*() methods on keys will report state from last frame.
//
// Listener.Think() will be called after all the events for a frame have been
// processed.
type Listener interface {
	HandleEventGroup(EventGroup)
	Think(int64)
}

type EventDispatcher interface {
	RegisterEventListener(Listener)
}
