package gin

import agg "github.com/runningwild/glop/gin/aggregator"

// A generalDerivedKey represents a group of natural keys. A key is specified
// with a KeyID which is a tuple of (key index, device type, device index).
// Given these components, the following are possible:
// (specific, specific, specific) - These are natural keys
// (specific, specific, general) - Specific key on any device of a specific type
// (specific, general, general) - Specific key on any device at all
// (general, specific, specific) - Any key on a specific device
// (general, specific, general) - Any key on any device of a specific type
// (general, general, general) - Any key on any device at all
// Note: It never makes sense to specify a device index without specifying the
// device type - doing so will cause glop to panic.
type generalDerivedKey struct {
	keyState
	press_amt float64

	// We need the input object itself so that we can get at all of the keys that
	// we depend on.
	input *Input
}

func (gdk *generalDerivedKey) CurPressAmt() float64 {
	// TODO(#28): this needs to go away; we've got keys indexed by a many of
	// their KeyID components so we should be doing lookups instead.
	sum := 0.0
	for _, key := range gdk.input.all_keys {
		if key.Id().Index == AnyKey ||
			key.Id().Device.Type == DeviceTypeAny ||
			key.Id().Device.Type == DeviceTypeDerived ||
			key.Id().Device.Index == DeviceIndexAny {
			continue
		}
		if gdk.Id().Index != AnyKey && key.Id().Index != gdk.Id().Index {
			// Not the appropriate key index
			continue
		}
		if gdk.Id().Device.Type != DeviceTypeAny &&
			key.Id().Device.Type != gdk.Id().Device.Type {
			// Not the appropriate device type
			continue
		}
		if gdk.Id().Device.Index != DeviceIndexAny &&
			key.Id().Device.Index != gdk.Id().Device.Index {
			// Not the appropriate device index
			continue
		}
		sum += key.CurPressAmt()
	}
	return sum
}

func (gdk *generalDerivedKey) IsDown() bool {
	return gdk.press_amt > 0
}

func (gdk *generalDerivedKey) KeySetPressAmt(amt float64, ms int64, cause Event) (event Event) {
	event.Type = agg.NoEvent
	event.Key = &gdk.keyState
	old_press_amt := gdk.press_amt
	gdk.press_amt = gdk.CurPressAmt()
	if (old_press_amt == 0) == (gdk.press_amt == 0) {
		return
	}
	if gdk.press_amt > 0 {
		event.Type = agg.Press
	} else {
		event.Type = agg.Release
	}
	gdk.keyState.Aggregator.AggregatorSetPressAmt(gdk.press_amt, ms, event.Type)
	return
}
