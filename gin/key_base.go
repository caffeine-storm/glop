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

func (ki KeyIndex) Contains(other KeyIndex) bool {
	if ki == AnyKey {
		return true
	}

	// If ki is a 'normal' key, it Contains exactly one KeyIndex: itself
	if ki < DerivedKeysRangeStart || ki >= DerivedKeysRangeEnd {
		return ki == other
	}

	switch ki {
	case EitherShift:
		return other == LeftShift || other == RightShift
	case EitherControl:
		return other == LeftControl || other == RightControl
	case EitherAlt:
		return other == LeftAlt || other == RightAlt
	case EitherGui:
		return other == LeftGui || other == RightGui
	case ShiftTab:
		return other == LeftShift || other == RightShift
	case DeleteOrBackspace:
		return other == KeyDelete || other == Backspace
	}

	panic(fmt.Errorf("KeyIndex.Contains: unknown KeyIndex: %d", int(ki)))
}

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

func (di DeviceId) Contains(other DeviceId) bool {
	return di.Index.Contains(other.Index) && di.Type.Contains(other.Type)
}

type DeviceIndex int

const (
	DeviceIndexAny DeviceIndex = -1
)

func (di DeviceIndex) Contains(other DeviceIndex) bool {
	if di == DeviceIndexAny {
		return true
	}
	return di == other
}

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

func (dt DeviceType) Contains(other DeviceType) bool {
	if dt == DeviceTypeAny {
		return true
	}
	return dt == other
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

// KeyIds support a quasi-wildcard form where an event for a single ID can
// 'cascade' over a set of other keys. Contains returns true iff the set of
// Keys covered by the 'cascade' includes the given KeyId.
func (lhs KeyId) Contains(rhs KeyId) bool {
	return lhs.Index.Contains(rhs.Index) && lhs.Device.Contains(rhs.Device)
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

	event.Key = ks
	event.Type = aggregator.DecideEventType(ks.CurPressAmt(), amt, ks.Aggregator)

	ks.Aggregator.AggregatorSetPressAmt(amt, ms, event.Type)
	return
}
