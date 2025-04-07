package gin

import (
	"fmt"

	"github.com/runningwild/glop/glog"
)

var (
	AnyAnyKey               = KeyId{Index: AnyKey, Device: DeviceId{Type: DeviceTypeAny, Index: DeviceIndexAny}}
	AnySpace                = KeyId{Index: Space, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyBackspace            = KeyId{Index: Backspace, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyTab                  = KeyId{Index: Tab, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyReturn               = KeyId{Index: Return, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyEscape               = KeyId{Index: Escape, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyA                 = KeyId{Index: KeyA, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyB                 = KeyId{Index: KeyB, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyC                 = KeyId{Index: KeyC, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyD                 = KeyId{Index: KeyD, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyE                 = KeyId{Index: KeyE, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyF                 = KeyId{Index: KeyF, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyG                 = KeyId{Index: KeyG, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyH                 = KeyId{Index: KeyH, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyI                 = KeyId{Index: KeyI, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyJ                 = KeyId{Index: KeyJ, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyK                 = KeyId{Index: KeyK, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyL                 = KeyId{Index: KeyL, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyM                 = KeyId{Index: KeyM, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyN                 = KeyId{Index: KeyN, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyO                 = KeyId{Index: KeyO, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyP                 = KeyId{Index: KeyP, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyQ                 = KeyId{Index: KeyQ, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyR                 = KeyId{Index: KeyR, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyS                 = KeyId{Index: KeyS, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyT                 = KeyId{Index: KeyT, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyU                 = KeyId{Index: KeyU, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyV                 = KeyId{Index: KeyV, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyW                 = KeyId{Index: KeyW, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyX                 = KeyId{Index: KeyX, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyY                 = KeyId{Index: KeyY, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyZ                 = KeyId{Index: KeyZ, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF1                   = KeyId{Index: F1, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF2                   = KeyId{Index: F2, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF3                   = KeyId{Index: F3, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF4                   = KeyId{Index: F4, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF5                   = KeyId{Index: F5, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF6                   = KeyId{Index: F6, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF7                   = KeyId{Index: F7, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF8                   = KeyId{Index: F8, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF9                   = KeyId{Index: F9, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF10                  = KeyId{Index: F10, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF11                  = KeyId{Index: F11, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyF12                  = KeyId{Index: F12, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyCapsLock             = KeyId{Index: CapsLock, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyNumLock              = KeyId{Index: NumLock, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyScrollLock           = KeyId{Index: ScrollLock, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyPrintScreen          = KeyId{Index: PrintScreen, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyPause                = KeyId{Index: Pause, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyLeftShift            = KeyId{Index: LeftShift, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyRightShift           = KeyId{Index: RightShift, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyLeftControl          = KeyId{Index: LeftControl, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyRightControl         = KeyId{Index: RightControl, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyLeftAlt              = KeyId{Index: LeftAlt, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyRightAlt             = KeyId{Index: RightAlt, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyLeftGui              = KeyId{Index: LeftGui, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyRightGui             = KeyId{Index: RightGui, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyRight                = KeyId{Index: Right, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyLeft                 = KeyId{Index: Left, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyUp                   = KeyId{Index: Up, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyDown                 = KeyId{Index: Down, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPadDivide         = KeyId{Index: KeyPadDivide, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPadMultiply       = KeyId{Index: KeyPadMultiply, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPadSubtract       = KeyId{Index: KeyPadSubtract, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPadAdd            = KeyId{Index: KeyPadAdd, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPadEnter          = KeyId{Index: KeyPadEnter, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPadDecimal        = KeyId{Index: KeyPadDecimal, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPadEquals         = KeyId{Index: KeyPadEquals, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad0              = KeyId{Index: KeyPad0, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad1              = KeyId{Index: KeyPad1, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad2              = KeyId{Index: KeyPad2, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad3              = KeyId{Index: KeyPad3, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad4              = KeyId{Index: KeyPad4, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad5              = KeyId{Index: KeyPad5, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad6              = KeyId{Index: KeyPad6, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad7              = KeyId{Index: KeyPad7, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad8              = KeyId{Index: KeyPad8, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPad9              = KeyId{Index: KeyPad9, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyDelete            = KeyId{Index: KeyDelete, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyHome              = KeyId{Index: KeyHome, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyInsert            = KeyId{Index: KeyInsert, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyEnd               = KeyId{Index: KeyEnd, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPageUp            = KeyId{Index: KeyPageUp, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyKeyPageDown          = KeyId{Index: KeyPageDown, Device: DeviceId{Type: DeviceTypeKeyboard, Index: DeviceIndexAny}}
	AnyMouseXAxis           = KeyId{Index: MouseXAxis, Device: DeviceId{Type: DeviceTypeMouse, Index: DeviceIndexAny}}
	AnyMouseYAxis           = KeyId{Index: MouseYAxis, Device: DeviceId{Type: DeviceTypeMouse, Index: DeviceIndexAny}}
	AnyMouseWheelVertical   = KeyId{Index: MouseWheelVertical, Device: DeviceId{Type: DeviceTypeMouse, Index: DeviceIndexAny}}
	AnyMouseWheelHorizontal = KeyId{Index: MouseWheelHorizontal, Device: DeviceId{Type: DeviceTypeMouse, Index: DeviceIndexAny}}
	AnyMouseLButton         = KeyId{Index: MouseLButton, Device: DeviceId{Type: DeviceTypeMouse, Index: DeviceIndexAny}}
	AnyMouseRButton         = KeyId{Index: MouseRButton, Device: DeviceId{Type: DeviceTypeMouse, Index: DeviceIndexAny}}
	AnyMouseMButton         = KeyId{Index: MouseMButton, Device: DeviceId{Type: DeviceTypeMouse, Index: DeviceIndexAny}}
)

const (
	AnyKey               KeyIndex = 0
	Space                         = 32
	Backspace                     = 8
	Tab                           = 9
	Return                        = 13
	Escape                        = 27
	KeyA                          = 97
	KeyB                          = 98
	KeyC                          = 99
	KeyD                          = 100
	KeyE                          = 101
	KeyF                          = 102
	KeyG                          = 103
	KeyH                          = 104
	KeyI                          = 105
	KeyJ                          = 106
	KeyK                          = 107
	KeyL                          = 108
	KeyM                          = 109
	KeyN                          = 110
	KeyO                          = 111
	KeyP                          = 112
	KeyQ                          = 113
	KeyR                          = 114
	KeyS                          = 115
	KeyT                          = 116
	KeyU                          = 117
	KeyV                          = 118
	KeyW                          = 119
	KeyX                          = 120
	KeyY                          = 121
	KeyZ                          = 122
	F1                            = 129
	F2                            = 130
	F3                            = 131
	F4                            = 132
	F5                            = 133
	F6                            = 134
	F7                            = 135
	F8                            = 136
	F9                            = 137
	F10                           = 138
	F11                           = 139
	F12                           = 140
	CapsLock                      = 150
	NumLock                       = 151
	ScrollLock                    = 152
	PrintScreen                   = 153
	Pause                         = 154
	LeftShift                     = 155
	RightShift                    = 156
	LeftControl                   = 157
	RightControl                  = 158
	LeftAlt                       = 159
	RightAlt                      = 160
	LeftGui                       = 161
	RightGui                      = 162
	Right                         = 166
	Left                          = 167
	Up                            = 168
	Down                          = 169
	KeyPadDivide                  = 170
	KeyPadMultiply                = 171
	KeyPadSubtract                = 172
	KeyPadAdd                     = 173
	KeyPadEnter                   = 174
	KeyPadDecimal                 = 175
	KeyPadEquals                  = 176
	KeyPad0                       = 177
	KeyPad1                       = 178
	KeyPad2                       = 179
	KeyPad3                       = 180
	KeyPad4                       = 181
	KeyPad5                       = 182
	KeyPad6                       = 183
	KeyPad7                       = 184
	KeyPad8                       = 185
	KeyPad9                       = 186
	KeyDelete                     = 190
	KeyHome                       = 191
	KeyInsert                     = 192
	KeyEnd                        = 193
	KeyPageUp                     = 194
	KeyPageDown                   = 195
	MouseXAxis                    = 300
	MouseYAxis                    = 301
	MouseWheelVertical            = 302
	MouseWheelHorizontal          = 303
	MouseLButton                  = 304
	MouseRButton                  = 305
	MouseMButton                  = 306

	// standard derived keys start here
	EitherShift = 1000 + iota
	EitherControl
	EitherAlt
	EitherGui
	ShiftTab
	DeleteOrBackspace
)

// Everything 'global' is put inside a struct so that tests can be run without
// stepping on each other.
type Input struct {
	// TODO(tmckee): This exists so that every key can be prodded with a
	// 'KeyThink' on _each frame_.
	all_keys []Key
	key_map  map[KeyId]Key

	// map from some keyId to a list of Keys that are observing keyId's presses.
	// e.g. 'S' -> []{'save-key', 'screenshot-key', ...}
	cause_to_effect map[KeyId][]Key

	// Mapping from KeyIndex to an aggregator of the appropriate type for that index.
	// This allows us to construct Keys for devices as the events happen, rather
	// than needing to know what the devices are during initialization.
	index_to_agg_type map[KeyIndex]aggregatorType

	// map from KeyIndex to a human-readable name for that key
	index_to_name map[KeyIndex]string

	// The listeners will receive all events immediately after those events have
	// been used to update all key states. The order in which listeners are
	// notified of a particular event group can change from group to group.
	listeners []Listener

	// Optional logger instance to trace calls to Input.
	logger glog.Logger
}

func (input *Input) SetLogger(logger glog.Logger) {
	if logger == nil {
		logger = glog.VoidLogger()
	}
	input.logger = logger
}

// The standard input object
var input_obj *Input

func init() {
	input_obj = Make()
}

// TODO: You messed up, the name of this function should be Input, and it
// should return an interface or something that is not called Input
func In() *Input {
	return input_obj
}

func Make() *Input {
	return MakeLogged(nil)
}

// Creates a new input object, mostly for testing. Most users will just query
// gin.Input, which is created during initialization
func MakeLogged(logger glog.Logger) *Input {
	input := new(Input)
	input.all_keys = make([]Key, 0, 512)
	input.key_map = make(map[KeyId]Key, 512)
	input.cause_to_effect = make(map[KeyId][]Key, 16)
	input.index_to_agg_type = make(map[KeyIndex]aggregatorType)
	input.index_to_name = make(map[KeyIndex]string)
	input.SetLogger(logger)

	input.registerKeyIndex(AnyKey, aggregatorTypeStandard, "AnyKey")
	for c := 'a'; c <= 'z'; c++ {
		name := fmt.Sprintf("Key %c", c+'A'-'a')
		input.registerKeyIndex(KeyIndex(c), aggregatorTypeStandard, name)
	}
	for _, c := range "0123456789`[]\\-=;',./" {
		name := fmt.Sprintf("Key %c", c)
		input.registerKeyIndex(KeyIndex(c), aggregatorTypeStandard, name)
	}
	input.registerKeyIndex(Space, aggregatorTypeStandard, "Space")
	input.registerKeyIndex(Backspace, aggregatorTypeStandard, "Backspace")
	input.registerKeyIndex(Tab, aggregatorTypeStandard, "Tab")
	input.registerKeyIndex(Return, aggregatorTypeStandard, "Return")
	input.registerKeyIndex(Escape, aggregatorTypeStandard, "Escape")
	input.registerKeyIndex(F1, aggregatorTypeStandard, "F1")
	input.registerKeyIndex(F2, aggregatorTypeStandard, "F2")
	input.registerKeyIndex(F3, aggregatorTypeStandard, "F3")
	input.registerKeyIndex(F4, aggregatorTypeStandard, "F4")
	input.registerKeyIndex(F5, aggregatorTypeStandard, "F5")
	input.registerKeyIndex(F6, aggregatorTypeStandard, "F6")
	input.registerKeyIndex(F7, aggregatorTypeStandard, "F7")
	input.registerKeyIndex(F8, aggregatorTypeStandard, "F8")
	input.registerKeyIndex(F9, aggregatorTypeStandard, "F9")
	input.registerKeyIndex(F10, aggregatorTypeStandard, "F10")
	input.registerKeyIndex(F11, aggregatorTypeStandard, "F11")
	input.registerKeyIndex(F12, aggregatorTypeStandard, "F12")
	input.registerKeyIndex(CapsLock, aggregatorTypeStandard, "CapsLock")
	input.registerKeyIndex(NumLock, aggregatorTypeStandard, "NumLock")
	input.registerKeyIndex(ScrollLock, aggregatorTypeStandard, "ScrollLock")
	input.registerKeyIndex(PrintScreen, aggregatorTypeStandard, "PrintScreen")
	input.registerKeyIndex(Pause, aggregatorTypeStandard, "Pause")
	input.registerKeyIndex(LeftShift, aggregatorTypeStandard, "LeftShift")
	input.registerKeyIndex(RightShift, aggregatorTypeStandard, "RightShift")
	input.registerKeyIndex(LeftControl, aggregatorTypeStandard, "LeftControl")
	input.registerKeyIndex(RightControl, aggregatorTypeStandard, "RightControl")
	input.registerKeyIndex(LeftAlt, aggregatorTypeStandard, "LeftAlt")
	input.registerKeyIndex(RightAlt, aggregatorTypeStandard, "RightAlt")
	input.registerKeyIndex(LeftGui, aggregatorTypeStandard, "LeftGui")
	input.registerKeyIndex(RightGui, aggregatorTypeStandard, "RightGui")
	input.registerKeyIndex(Right, aggregatorTypeStandard, "Right")
	input.registerKeyIndex(Left, aggregatorTypeStandard, "Left")
	input.registerKeyIndex(Up, aggregatorTypeStandard, "Up")
	input.registerKeyIndex(Down, aggregatorTypeStandard, "Down")
	input.registerKeyIndex(KeyPadDivide, aggregatorTypeStandard, "KeyPadDivide")
	input.registerKeyIndex(KeyPadMultiply, aggregatorTypeStandard, "KeyPadMultiply")
	input.registerKeyIndex(KeyPadSubtract, aggregatorTypeStandard, "KeyPadSubtract")
	input.registerKeyIndex(KeyPadAdd, aggregatorTypeStandard, "KeyPadAdd")
	input.registerKeyIndex(KeyPadEnter, aggregatorTypeStandard, "KeyPadEnter")
	input.registerKeyIndex(KeyPadDecimal, aggregatorTypeStandard, "KeyPadDecimal")
	input.registerKeyIndex(KeyPadEquals, aggregatorTypeStandard, "KeyPadEquals")
	input.registerKeyIndex(KeyPad0, aggregatorTypeStandard, "KeyPad0")
	input.registerKeyIndex(KeyPad1, aggregatorTypeStandard, "KeyPad1")
	input.registerKeyIndex(KeyPad2, aggregatorTypeStandard, "KeyPad2")
	input.registerKeyIndex(KeyPad3, aggregatorTypeStandard, "KeyPad3")
	input.registerKeyIndex(KeyPad4, aggregatorTypeStandard, "KeyPad4")
	input.registerKeyIndex(KeyPad5, aggregatorTypeStandard, "KeyPad5")
	input.registerKeyIndex(KeyPad6, aggregatorTypeStandard, "KeyPad6")
	input.registerKeyIndex(KeyPad7, aggregatorTypeStandard, "KeyPad7")
	input.registerKeyIndex(KeyPad8, aggregatorTypeStandard, "KeyPad8")
	input.registerKeyIndex(KeyPad9, aggregatorTypeStandard, "KeyPad9")
	input.registerKeyIndex(KeyDelete, aggregatorTypeStandard, "KeyDelete")
	input.registerKeyIndex(KeyHome, aggregatorTypeStandard, "KeyHome")
	input.registerKeyIndex(KeyInsert, aggregatorTypeStandard, "KeyInsert")
	input.registerKeyIndex(KeyEnd, aggregatorTypeStandard, "KeyEnd")
	input.registerKeyIndex(KeyPageUp, aggregatorTypeStandard, "KeyPageUp")
	input.registerKeyIndex(KeyPageDown, aggregatorTypeStandard, "KeyPageDown")

	input.registerKeyIndex(MouseXAxis, aggregatorTypeAxis, "X Axis")
	input.registerKeyIndex(MouseYAxis, aggregatorTypeAxis, "Y Axis")
	input.registerKeyIndex(MouseWheelVertical, aggregatorTypeWheel, "MouseWheel")
	input.registerKeyIndex(MouseLButton, aggregatorTypeStandard, "MouseLButton")
	input.registerKeyIndex(MouseRButton, aggregatorTypeStandard, "MouseRButton")
	input.registerKeyIndex(MouseMButton, aggregatorTypeStandard, "MouseMButton")

	// TODO(#28): bind these 'default' derived keys
	// input.bindDerivedKeyWithId("Shift", EitherShift, input.MakeBinding(LeftShift, nil, nil), input.MakeBinding(RightShift, nil, nil))
	// input.bindDerivedKeyWithId("Control", EitherControl, input.MakeBinding(LeftControl, nil, nil), input.MakeBinding(RightControl, nil, nil))
	// input.bindDerivedKeyWithId("Alt", EitherAlt, input.MakeBinding(LeftAlt, nil, nil), input.MakeBinding(RightAlt, nil, nil))
	// input.bindDerivedKeyWithId("Gui", EitherGui, input.MakeBinding(LeftGui, nil, nil), input.MakeBinding(RightGui, nil, nil))
	// input.bindDerivedKeyWithId("ShiftTab", ShiftTab, input.MakeBinding(Tab, []KeyId{EitherShift}, []bool{true}))
	// input.bindDerivedKeyWithId("DeleteOrBackspace", DeleteOrBackspace, input.MakeBinding(KeyDelete, nil, nil), input.MakeBinding(Backspace, nil, nil))
	return input
}

func (input *Input) registerKeyIndex(index KeyIndex, agg_type aggregatorType, name string) {
	input.logger.Trace("gin.Input")
	if index < 0 {
		panic(fmt.Errorf("cannot register a key with a negative index: %d", index))
	}
	if prev, ok := input.index_to_name[index]; ok {
		panic(fmt.Errorf("cannot overwrite key registration: index %d, new name %q, old name %q", index, name, prev))
	}
	input.index_to_agg_type[index] = agg_type
	input.index_to_name[index] = name
}

func (input *Input) GetKeyByParts(key_index KeyIndex, device_type DeviceType, device_index DeviceIndex) Key {
	input.logger.Trace("gin.Input")
	return input.GetKeyById(KeyId{
		Index: key_index,
		Device: DeviceId{
			Index: device_index,
			Type:  device_type,
		},
	})
}

func (input *Input) GetKeyById(id KeyId) Key {
	input.logger.Trace("gin.Input")
	id.MustValidate()
	key, ok := input.key_map[id]
	if !ok {
		if id.Index == AnyKey || id.Device.Type == DeviceTypeAny || id.Device.Index == DeviceIndexAny {
			// If we're looking for a general key we know how to create those
			input.key_map[id] = &generalDerivedKey{
				keyState: keyState{
					id:         id,
					name:       "Name me?",
					aggregator: &standardAggregator{},
				},
				input: input,
			}
			key = input.key_map[id]
			input.all_keys = append(input.all_keys, key)
		} else {
			// Check if the index is valid, if it is then we can just create a new
			// key the appropriate device.
			agg_type, ok := input.index_to_agg_type[id.Index]
			if !ok {
				panic(fmt.Errorf("no key registered with id == %v", id))
			}
			input.key_map[id] = &keyState{
				id:         id,
				name:       input.index_to_name[id.Index],
				aggregator: aggregatorForType(agg_type),
			}
			key = input.key_map[id]
			input.all_keys = append(input.all_keys, key)
		}
	}
	return key
}

func (input *Input) GetKeyByName(name string) Key {
	input.logger.Trace("gin.Input")
	for _, key := range input.key_map {
		if key.Name() == name {
			return key
		}
	}
	return nil
}

// The Input object can have multiple Listener instances registered with it.
// Each Listener will receive event groups as they are processed. Each Listener
// will also get a .Think() call once per frame after all input events for the
// frame have been processed.
func (input *Input) RegisterEventListener(listener Listener) {
	input.logger.Trace("gin.Input")
	input.listeners = append(input.listeners, listener)
}

// Returns true if triggering 'cause' will trigger 'effect'.
func (input *Input) willTrigger(cause, effect KeyId) bool {
	if cause == effect {
		return true
	}

	// input.id_to_deps encodes a DAG of KeyId interdependence. Start at the
	// 'from' node and BFS for 'to' in the set of descendents.
	visited := map[KeyId]bool{}
	workQueue := []KeyId{cause}

	for len(workQueue) > 0 {
		nextCause := workQueue[0]
		workQueue = workQueue[1:]

		if visited[nextCause] {
			continue
		}

		visited[nextCause] = true
		for _, nextEffect := range input.cause_to_effect[nextCause] {
			if nextEffect.Id() == effect {
				return true
			}
			if !visited[nextEffect.Id()] {
				// Check for transitive effects by treating effects as new causes.
				workQueue = append(workQueue, nextEffect.Id())
			}
		}

		nextCauseIgnoringDeviceIndex := nextCause
		nextCauseIgnoringDeviceIndex.Device.Index = DeviceIndexAny
		for _, nextEffect := range input.cause_to_effect[nextCauseIgnoringDeviceIndex] {
			if nextEffect.Id() == effect {
				return true
			}
			if !visited[nextEffect.Id()] {
				// Check for transitive effects by treating effects as new causes.
				workQueue = append(workQueue, nextEffect.Id())
			}
		}
	}

	return false
}

func (input *Input) addCauseEffect(cause KeyId, effect Key) {
	input.logger.Trace("gin.Input>addObserver", "derived", effect, "dep", cause)

	if input.willTrigger(effect.Id(), cause) {
		panic(fmt.Errorf("depedency cycle detected: %v depends on %v", cause, effect.Id()))
	}

	list := input.cause_to_effect[cause]
	list = append(list, effect)
	input.cause_to_effect[cause] = list
}

func (input *Input) removeCauseEffect(cause KeyId, effect Key) {
	list, ok := input.cause_to_effect[cause]
	if !ok {
		panic(fmt.Errorf("no effects known for cause %v", cause))
	}

	newList := make([]Key, len(list))
	out := 0
	for i := 0; i < len(list); i++ {
		if list[i] == effect {
			continue
		}

		newList[out] = list[i]
		out++
	}

	// If there was no match, the caller is probably confused.
	if out == len(list) {
		panic(fmt.Errorf("no cause/effect existed for %v/%v", cause, effect))
	}

	newList = newList[:out]
	input.cause_to_effect[cause] = newList
}

// Returns the Keys that need to be notified when the given KeyId is triggered.
func (input *Input) findKeyIdObservers(id KeyId) []Key {
	id_ignoring_device_index := id
	id_ignoring_device_index.Device.Index = DeviceIndexAny

	// Direct dependencies are recorded in input.id_to_deps
	keysToPress := input.cause_to_effect[id]

	// TODO(tmckee): consider using a set instead of a list for the keys to
	// press... if presses are idempotent, we don't need to press them again (so
	// don't bother walking them) _OR_ if they're not idempotent, how in the heck
	// would we manage to get a reasonable "press amount" after all this!?

	// Dependencies for keys organized by the same 'KeyIndex' but not pinned to a
	// particular device instance, though the 'DeviceType' does need to match
	// 🤔...
	for _, dep := range input.cause_to_effect[id_ignoring_device_index] {
		keysToPress = append(keysToPress, dep)
	}

	return keysToPress
}

func (input *Input) pressKey(k Key, amt float64, cause Event, group *EventGroup) {
	event := k.KeySetPressAmt(amt, group.Timestamp, cause)
	keysToPress := input.findKeyIdObservers(event.Key.Id())
	if event.Type != NoEvent {
		group.Events = append(group.Events, event)
	}
	for _, dep := range keysToPress {
		input.pressKey(dep, dep.CurPressAmt(), event, group)
	}

	// Press synthetic keys (like, the 'Any' key)
	if k.Id().Index != AnyKey && k.Id().Device.Type != DeviceTypeAny && k.Id().Device.Type != DeviceTypeDerived && k.Id().Device.Index != DeviceIndexAny {
		general_keys := []Key{
			input.GetKeyByParts(AnyKey, k.Id().Device.Type, k.Id().Device.Index),
			input.GetKeyByParts(AnyKey, k.Id().Device.Type, DeviceIndexAny),
			input.GetKeyByParts(AnyKey, DeviceTypeAny, DeviceIndexAny),
			input.GetKeyByParts(k.Id().Index, k.Id().Device.Type, DeviceIndexAny),
			input.GetKeyByParts(k.Id().Index, DeviceTypeAny, DeviceIndexAny),
		}
		for _, general_key := range general_keys {
			input.pressKey(general_key, amt, cause, group)
		}
	}
}

func (input *Input) Think(t int64, os_events []OsEvent) []EventGroup {
	// Generate all key events here. Derived keys are handled through pressKey
	// and all events are aggregated into one array. Events in this array will
	// necessarily be in sorted order.
	var groups []EventGroup
	for _, os_event := range os_events {
		group := EventGroup{
			Timestamp: os_event.Timestamp,
		}

		// Whether this was a keyboard keystroke or actually a mouse thing, still
		// update the x/y mouse position. Imagine, for example, a hotkey that
		// behaves differently depending on where the mouse is; it will need to
		// have a way to find current mouse position. Don't worry, native code is
		// expected to populate cursor_x, cursor_y for all OsEvents.
		group.SetMousePosition(os_event.X, os_event.Y)

		input.pressKey(
			input.GetKeyById(os_event.KeyId),
			os_event.Press_amt,
			Event{},
			&group)

		if len(group.Events) > 0 {
			groups = append(groups, group)
			for _, listener := range input.listeners {
				listener.HandleEventGroup(group)
			}
		}
	}

	for _, key := range input.all_keys {
		synthesizeNewEvent, amt := key.KeyThink(t)
		if !synthesizeNewEvent {
			continue
		}
		glog.TraceLogger().Trace("synthetic event", "source key", key)

		// Here, we don't set a mouse position because it wouldn't make sense for
		// synthetic keys.
		group := EventGroup{
			Timestamp: t,
		}
		input.pressKey(key, amt, Event{}, &group)
		if len(group.Events) > 0 {
			groups = append(groups, group)
			for _, listener := range input.listeners {
				listener.HandleEventGroup(group)
			}
		}
	}

	for _, listener := range input.listeners {
		listener.Think(t)
	}
	return groups
}
