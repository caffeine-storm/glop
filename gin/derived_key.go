package gin

import (
	"fmt"

	"github.com/runningwild/glop/glog"
)

var (
	next_derived_key_index KeyIndex
)

func init() {
	next_derived_key_index = 10000
}

func genDerivedKeyIndex() KeyIndex {
	next_derived_key_index++
	return next_derived_key_index
}

// TODO: Handle removal of dependencies
func (input *Input) registerDependence(derived Key, dep KeyId) {
	input.logger.Trace("gin.Input>registerDependence", "derived", derived, "dep", dep)
	list, ok := input.id_to_deps[dep]
	if !ok {
		list = make([]Key, 0)
	}
	list = append(list, derived)
	input.id_to_deps[dep] = list
	if derived.Id() == dep {
		panic("Can't have a key depend on itself.")
	}
}

func (input *Input) BindDerivedKey(name string, bindings ...Binding) Key {
	input.logger.Trace("gin.input")
	return input.bindDerivedKeyWithIndex(
		name,
		genDerivedKeyIndex(),
		DeviceId{Index: 1, Type: DeviceTypeDerived},
		bindings...)
}

func (input *Input) bindDerivedKeyWithIndex(name string, index KeyIndex, device DeviceId, bindings ...Binding) Key {
	input.logger.Trace("gin.input")
	dk := &derivedKey{
		keyState: keyState{
			id: KeyId{
				Index:  index,
				Device: device,
			},
			name:       name,
			aggregator: &standardAggregator{},
		},
		Bindings:      bindings,
		bindings_down: make([]bool, len(bindings)),
	}

	// TODO: Decide whether or not this is true, might need to register them for
	// when the game loses focus.
	// I think it might not be necessary to register derived keys.
	// input.registerKeyIndex(dk.id.Index, &standardAggregator{}, name)

	for _, binding := range bindings {
		input.registerDependence(dk, binding.PrimaryKey)
		for _, modifier := range binding.Modifiers {
			input.registerDependence(dk, modifier)
		}
	}

	// TODO: Figure out a way to move this into Input.GetKeyById() or something.
	// It's really dirty to have these maps/slices populated in multiple places.
	input.key_map[dk.id] = dk
	input.all_keys = append(input.all_keys, dk)
	return dk
}

// A derivedKey is down if any of its bindings are down
type derivedKey struct {
	keyState
	is_down bool

	Bindings []Binding

	// We store the down state of all of our bindings so that things
	bindings_down []bool
}

var _ Key = (*derivedKey)(nil)

func (dk *derivedKey) CurPressAmt() float64 {
	sum := 0.0
	for index := range dk.Bindings {
		if dk.bindings_down[index] {
			sum += dk.Bindings[index].primaryPressAmt()
		} else {
			sum += dk.Bindings[index].CurPressAmt()
		}
	}
	return sum
}

func (dk *derivedKey) IsDown() bool {
	return dk.numBindingsDown() > 0
}

func (dk *derivedKey) numBindingsDown() int {
	count := 0
	for _, down := range dk.bindings_down {
		if down {
			count++
		}
	}
	return count
}

func (dk *derivedKey) SetPressAmt(amt float64, ms int64, cause Event) (event Event) {
	index := -1
	for i, binding := range dk.Bindings {
		if cause.Key.Id() == binding.PrimaryKey {
			index = i
		}
	}
	event.Type = NoEvent
	event.Key = &dk.keyState
	if amt == 0 && index != -1 && dk.numBindingsDown() == 1 && dk.bindings_down[index] {
		event.Type = Release
	}
	if amt != 0 && index != -1 && dk.numBindingsDown() == 0 && !dk.bindings_down[index] {
		glog.TraceLogger().Trace("Generated press event", "key", dk)
		event.Type = Press
	}
	if index != -1 {
		dk.bindings_down[index] = dk.Bindings[index].CurPressAmt() != 0
	}
	dk.keyState.aggregator.SetPressAmt(amt, ms, event.Type)
	return
}

// A Binding, b, is considered down if b.PrimaryKey is down and all of
// b.Modifiers[i].IsDown() matches b.Down[i]. Yes, KeyIds don't have an
// IsDown(), we're really talking about the Key instance identified by the
// KeyId.
type Binding struct {
	PrimaryKey KeyId
	Modifiers  []KeyId
	Down       []bool
	Input      *Input
}

func (input *Input) MakeBinding(primary KeyId, modifiers []KeyId, down []bool) Binding {
	input.logger.Trace("gin.input")
	if len(modifiers) != len(down) {
		panic("MakeBindings(primary, modifiers, down) - modifiers and down must have the same length.")
	}
	return Binding{
		PrimaryKey: primary,
		Modifiers:  modifiers,
		Down:       down,
		Input:      input,
	}
}

func (b *Binding) primaryPressAmt() float64 {
	mappedKey, ok := b.Input.key_map[b.PrimaryKey]
	if !ok {
		panic(fmt.Errorf("couldn't find PrimaryKey %+v in key_map", b.PrimaryKey))
	}
	return mappedKey.CurPressAmt()
}

func (b *Binding) CurPressAmt() float64 {
	for i := range b.Modifiers {
		if b.Input.GetKeyById(b.Modifiers[i]).IsDown() != b.Down[i] {
			return 0
		}
	}
	key, ok := b.Input.key_map[b.PrimaryKey]
	if !ok || !key.IsDown() {
		return 0
	}
	return key.CurPressAmt()
}
