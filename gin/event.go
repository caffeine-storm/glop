package gin

import "fmt"

type EventType int

const (
	NoEvent EventType = iota
	Press
	Release
	Adjust // The key was and is down, but the value of it has changed
)

func (event EventType) String() string {
	switch event {
	case Press:
		return "press"
	case Release:
		return "release"
	case NoEvent:
		return "noevent"
	case Adjust:
		return "adjust"
	}
	panic(fmt.Errorf("%d is not a valid EventType", event))
}

// A view over the data that comes back from native code.
type OsEvent struct {
	KeyId     KeyId
	Press_amt float64
	Timestamp int64
	X, Y      int
}

// TODO: Consider making a Timestamp type (int64)
type Event struct {
	Key  Key
	Type EventType
}

func (e Event) String() string {
	if e.Key == nil || e.Type == NoEvent {
		return fmt.Sprintf("NoEvent")
	}
	return fmt.Sprintf("'%v %v'", e.Type, e.Key)
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

// Returns a bool indicating whether an event corresponding to the given KeyId
// is present in the EventGroup, and if so the Event returned is a copy of that
// event.
func (eg *EventGroup) FindEvent(id KeyId) (bool, Event) {
	for i := range eg.Events {
		if eg.Events[i].Key.Id() == id {
			return true, eg.Events[i]
		}
	}
	return false, Event{}
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

// During HandleEventGroup a Listener can query keys as to their current state
// (i.e.  with Cur*() methods) and these will accurately report their state.
//
// Frame*() methods on keys will report state from last frame.
//
// Listener.Think() will be called after all the events for a frame have been
// processed.
//
// TODO(tmckee:#20) Instead of having every Listener (maybe, Thinker?) also
// implement 'EventHandler', don't couple them. Just register the same object
// that implmenets both interfaces with two registration calls.
type Listener interface {
	HandleEventGroup(EventGroup)
	Think(int64)
}

type EventDispatcher interface {
	RegisterEventListener(Listener)
}
