package gin

import (
	"fmt"

	"github.com/runningwild/glop/gin/aggregator"
	"github.com/runningwild/glop/glog"
)

type Key interface {
	// Human readable name
	String() string

	// Raw name, usable for lookup by GetKeyByName()
	Name() string

	// Unique Id
	Id() KeyId

	// Sets the instantaneous press amount for this key at a specific time and
	// returns the event generated, if any.
	KeySetPressAmt(amt float64, ms int64, cause Event) Event

	// A Key may return true, amt from KeyThink() to indicate that a fake event
	// should be generated to set its press amount to amt.
	KeyThink(ms int64) (bool, float64)

	aggregator.SubAggregator
}

type KeyIndex int
type KeyId struct {
	Device DeviceId
	Index  KeyIndex
}

func (kid KeyId) MustValidate() {
	if kid.Device.Type >= DeviceTypeMax || kid.Device.Type < 0 {
		panic(fmt.Errorf("invalid device type: %d", kid.Device.Type))
	}
	if kid.Device.Type == DeviceTypeAny && kid.Device.Index != DeviceIndexAny {
		panic(fmt.Errorf("DeviceTypeAny requires DeviceIndexAny but got %v", kid.Device))
	}
}

type DeviceId struct {
	Type  DeviceType
	Index DeviceIndex
}

type DeviceIndex int

const (
	DeviceIndexAny DeviceIndex = -1
)

type DeviceType int

const (
	DeviceTypeAny DeviceType = iota
	DeviceTypeKeyboard
	DeviceTypeMouse
	DeviceTypeController
	DeviceTypeDerived
	DeviceTypeMax
)

func (dt DeviceType) String() string {
	switch dt {
	case DeviceTypeAny:
		return "any"
	case DeviceTypeKeyboard:
		return "keyboard"
	case DeviceTypeMouse:
		return "mouse"
	case DeviceTypeController:
		return "controller"
	case DeviceTypeDerived:
		return "derived"
	case DeviceTypeMax:
		return "max"
	}

	panic(fmt.Errorf("bad DeviceType: %d", int(dt)))
}

func (kid KeyId) String() string {
	// Unfortunately, KeyId values are overloaded to also support 'querying';
	// sometimes things have a sentinel value in order to control lookups.
	device := "any"
	devicetype := "any"
	if kid.Index == AnyKey {
		return "any-key"
	}
	index := fmt.Sprintf("%d", kid.Index)

	if kid.Device.Type != DeviceTypeAny {
		devicetype = fmt.Sprintf("%v", kid.Device.Type)
	}

	if kid.Device.Index != DeviceIndexAny {
		device = fmt.Sprintf("%d", kid.Device.Index)
	}

	return fmt.Sprintf("{device: %s, devicetype: %s, index: %s}", device, devicetype, index)
}

// KeyIds support a quasi-wildcard form where a single ID can represent a
// family of keys. Matches returns true iff the set of Keys identified by each
// KeyId has a non-empty intersection.
func (lhs KeyId) Matches(rhs KeyId) bool {
	if lhs.Index != AnyKey && rhs.Index != AnyKey {
		// If neither key represents 'any-key-index', the indices have to match
		if lhs.Index != rhs.Index {
			return false
		}
	}

	if lhs.Device.Type != DeviceTypeAny && rhs.Device.Type != DeviceTypeAny {
		if lhs.Device.Type != rhs.Device.Type {
			return false
		}
	}

	if lhs.Device.Index != DeviceIndexAny && rhs.Device.Index != DeviceIndexAny {
		if lhs.Device.Index != rhs.Device.Index {
			return false
		}
	}

	return true
}

// natural keys and derived keys all embed a keyState
type keyState struct {
	id   KeyId  // Unique id among all keys ever
	name string // Human readable name for the key, 'Right Shift', 'q', 'Space Bar', etc...

	aggregator.Aggregator
}

var _ Key = (*keyState)(nil)

func (ks *keyState) GetAggregator() *aggregator.Aggregator {
	return &ks.Aggregator
}

func (ks *keyState) KeyThink(ms int64) (bool, float64) {
	return ks.Aggregator.AggregatorThink(ms)
}

func (ks *keyState) String() string {
	return fmt.Sprintf("{keyState: %q id: %v agg: %v}", ks.name, ks.id, ks.Aggregator)
}

func (ks *keyState) Name() string {
	return ks.name
}

func (ks *keyState) Id() KeyId {
	return ks.id
}

// Tells this key that it was pressed, by how much and at what time. Times must
// be monotonically increasing. If this press was caused by another event (as
// is the case with derived keys), then cause is the event that made this
// happen.
func (ks *keyState) KeySetPressAmt(amt float64, ms int64, cause Event) (event Event) {
	glog.TraceLogger().Trace("KeySetPressAmt", "keyid", ks.id, "amt", amt, "ks.agg", ks.Aggregator)
	event.Type = aggregator.NoEvent
	event.Key = ks
	if (ks.CurPressAmt() == 0) != (amt == 0) {
		if amt == 0 {
			event.Type = aggregator.Release
		} else {
			event.Type = aggregator.Press
		}
	} else {
		if ks.CurPressAmt() != 0 && ks.CurPressAmt() != amt {
			event.Type = aggregator.Adjust
		} else if ks.SendAllNonZero() {
			event.Type = aggregator.Adjust
		}
	}
	ks.Aggregator.AggregatorSetPressAmt(amt, ms, event.Type)
	return
}
