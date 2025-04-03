package gin

import (
	"fmt"
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

	// A very select set of keys should always send events when their press amt
	// is non-zero. These are typically not your ordinary keys, mouse wheels,
	// mouse pointers, etc...
	SendAllNonZero() bool

	// A Key may return true, amt from KeyThink() to indicate that a fake event
	// should be generated to set its press amount to amt.
	KeyThink(ms int64) (bool, float64)

	subAggregator
}

type subAggregator interface {
	IsDown() bool
	FramePressCount() int
	FrameReleaseCount() int
	FramePressAmt() float64
	FramePressSum() float64
	FramePressAvg() float64
	CurPressCount() int
	CurReleaseCount() int
	CurPressAmt() float64
	CurPressSum() float64
}

type aggregator interface {
	subAggregator
	AggregatorThink(ms int64) (bool, float64)
	AggregatorSetPressAmt(amt float64, ms int64, event_type EventType)
	SendAllNonZero() bool
}

type aggregatorType int

const (
	aggregatorTypeStandard aggregatorType = iota
	aggregatorTypeAxis
	aggregatorTypeWheel
)

func aggregatorForType(tp aggregatorType) aggregator {
	switch tp {
	case aggregatorTypeStandard:
		return &standardAggregator{}
	case aggregatorTypeAxis:
		return &axisAggregator{}
	case aggregatorTypeWheel:
		return &wheelAggregator{}
	}

	panic(fmt.Errorf("unknown aggregatorType: %d", tp))
}

func (at aggregatorType) MustValidate() {
	switch at {
	case aggregatorTypeStandard:
	case aggregatorTypeAxis:
	case aggregatorTypeWheel:
	default:
		panic(fmt.Errorf("invalid aggregatorType: %d", at))
	}
}

// Simple struct that aggregates presses and press_amts during a frame so they
// can be viewed between KeyThink()s
type keyStats struct {
	press_count   int
	release_count int
	press_amt     float64
	press_sum     float64
	press_avg     float64
}

type baseAggregator struct {
	this, prev keyStats
}

func (a *baseAggregator) FramePressCount() int {
	return a.prev.press_count
}

func (a *baseAggregator) FrameReleaseCount() int {
	return a.prev.release_count
}

func (a *baseAggregator) FramePressAmt() float64 {
	return a.prev.press_amt
}

func (a *baseAggregator) FramePressSum() float64 {
	return a.prev.press_sum
}

func (a *baseAggregator) FramePressAvg() float64 {
	return a.prev.press_avg
}

func (a *baseAggregator) CurPressCount() int {
	return a.this.press_count
}

func (a *baseAggregator) CurReleaseCount() int {
	return a.this.release_count
}

func (a *baseAggregator) CurPressAmt() float64 {
	return a.this.press_amt
}

func (a *baseAggregator) CurPressSum() float64 {
	return a.this.press_sum
}

func (a *baseAggregator) updateCounts(event_type EventType) {
	switch event_type {
	case Press:
		a.this.press_count++
	case Release:
		a.this.release_count++
	}
}

func (a *baseAggregator) SendAllNonZero() bool {
	return false
}

// The standardAggregator's sum is an integral of the press_amt over time
type standardAggregator struct {
	baseAggregator
	last_press int64
	last_think int64
}

func (sa *standardAggregator) IsDown() bool {
	return sa.this.press_amt != 0
}

func (sa *standardAggregator) AggregatorSetPressAmt(amt float64, ms int64, event_type EventType) {
	sa.this.press_sum += sa.this.press_amt * float64(ms-sa.last_press)
	sa.this.press_amt = amt
	sa.last_press = ms
	sa.updateCounts(event_type)
}

func (sa *standardAggregator) AggregatorThink(ms int64) (bool, float64) {
	sa.this.press_sum += sa.this.press_amt * float64(ms-sa.last_press)
	if ms != sa.last_think {
		sa.this.press_avg = sa.this.press_sum / float64(ms-sa.last_think)
	} else {
		sa.this.press_avg = 0
	}
	sa.prev = sa.this
	sa.this = keyStats{
		press_amt: sa.prev.press_amt,
	}
	sa.last_press = ms
	sa.last_think = ms
	return false, 0
}

// The axisAggregator's sum is the sum of all press amounts specified by
// SetPressAmt(). FramePressAvg() returns the same value as FramePressSum().
type axisAggregator struct {
	baseAggregator
	is_down bool
}

func (aa *axisAggregator) IsDown() bool {
	return aa.is_down
}

func (aa *axisAggregator) AggregatorSetPressAmt(amt float64, ms int64, event_type EventType) {
	aa.this.press_sum += amt
	aa.this.press_amt = amt
	if amt != 0 {
		aa.is_down = true
	}
	aa.updateCounts(event_type)
}

func (aa *axisAggregator) AggregatorThink(ms int64) (bool, float64) {
	was_down := aa.prev.press_amt != 0
	aa.prev = aa.this
	aa.this = keyStats{}
	aa.prev.press_avg = aa.prev.press_sum
	if aa.prev.press_amt == 0 {
		aa.is_down = false
		if was_down {
			return true, 0
		}
	}
	return false, 0
}

// A wheelAggregator is just like a standardAggregator except for two things:
// - It sends Adjust events for *all* non-zero press amounts
// - If a frame goes by without it receiving any input it creates a Release
// event
type wheelAggregator struct {
	standardAggregator
	event_received bool
}

func (wa *wheelAggregator) SendAllNonZero() bool {
	return true
}

func (wa *wheelAggregator) AggregatorSetPressAmt(amt float64, ms int64, event_type EventType) {
	wa.event_received = wa.last_press < wa.last_think
	wa.standardAggregator.AggregatorSetPressAmt(amt, ms, event_type)
}

func (wa *wheelAggregator) AggregatorThink(ms int64) (bool, float64) {
	if b, _ := wa.standardAggregator.AggregatorThink(ms); b {
		panic("standardAggregator should not generate an event on AggregatorThink()")
	}
	if wa.CurPressAmt() != 0 {
		if wa.event_received {
			wa.event_received = false
			return true, wa.CurPressAmt()
		} else {
			return true, 0
		}
	}
	return false, 0
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

// natural keys and derived keys all embed a keyState
type keyState struct {
	id   KeyId  // Unique id among all keys ever
	name string // Human readable name for the key, 'Right Shift', 'q', 'Space Bar', etc...

	aggregator
}

var _ Key = (*keyState)(nil)

func (ks *keyState) KeyThink(ms int64) (bool, float64) {
	return ks.aggregator.AggregatorThink(ms)
}

func (ks *keyState) String() string {
	return fmt.Sprintf("{keyState:%q id: %v}", ks.name, ks.id)
}

func (ks *keyState) Name() string {
	return ks.name
}

func (ks *keyState) Id() KeyId {
	return ks.id
}

// Tells this key that how much it was pressed at a particular time. Times must
// be monotonically increasing. If this press was caused by another event (as
// is the case with derived keys), then cause is the event that made this
// happen.
func (ks *keyState) KeySetPressAmt(amt float64, ms int64, cause Event) (event Event) {
	event.Type = NoEvent
	event.Key = ks
	if (ks.CurPressAmt() == 0) != (amt == 0) {
		if amt == 0 {
			event.Type = Release
		} else {
			event.Type = Press
		}
	} else {
		if ks.CurPressAmt() != 0 && ks.CurPressAmt() != amt {
			event.Type = Adjust
		} else if ks.SendAllNonZero() {
			event.Type = Adjust
		}
	}
	ks.aggregator.AggregatorSetPressAmt(amt, ms, event.Type)
	return
}
