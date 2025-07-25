package aggregator

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

func DecideEventType(curPressAmount, newPressAmount float64, agg Aggregator) EventType {
	if curPressAmount == newPressAmount {
		// Nothing's really changing but some keys need to report an 'adjust' here
		// anyways.
		if agg.SendAllNonZero() {
			return Adjust
		}
		return NoEvent
	}

	if curPressAmount == 0 {
		// We should only return 'Press' if we're transitioning from 0 to not-0.
		return Press
	}

	if newPressAmount == 0 {
		// We should only return 'Release' if we're transitioning from not-0 to 0.
		return Release
	}

	// The key is pressed before and after but at different amounts; sounds like
	// an adjustment to me!
	return Adjust
}
