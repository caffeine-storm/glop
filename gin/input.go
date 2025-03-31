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

type OsEvent struct {
	KeyId     KeyId
	Press_amt float64
	Timestamp int64
	X, Y      int
}

// Everything 'global' is put inside a struct so that tests can be run without
// stepping on each other.
type Input struct {
	// TODO(tmckee): I think 'all_keys' is a misnomer; we only put derived keys
	// in this collection... Seems to be done so that they can be prodded with a
	// 'Think' on _each frame_ :(
	all_keys []Key
	key_map  map[KeyId]Key

	// map from keyId to list of derived Keys and general derived Keys that
	// depend on it in some way
	id_to_deps map[KeyId][]Key

	// Mapping from KeyIndex to list of derived key families that depend on it.
	// This map should only be keyed by indices generated for derived keys;
	// otherwise we'd have to include info to distinguish between devices too!
	index_to_family_deps map[KeyIndex][]derivedKeyFamily

	// Mapping from KeyIndex to the derivedKeyFamily that it represents, if any.
	// This map should only be keyed by indices generated for derived keys;
	// otherwise we'd have to include info to distinguish between devices too!
	index_to_family map[KeyIndex]derivedKeyFamily

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

	// Delegate for Mouse events/handling. Assumes only one 'pointer' device at a
	// time. To handle multiple pointers, we'll need a collection distinguished
	// by device ID.
	mouse MouseInput

	// Optional logger instance to trace calls to Input.
	logger glog.Logger
}

func (input *Input) SetLogger(logger glog.Logger) {
	if logger == nil {
		logger = glog.VoidLogger()
	}
	input.logger = logger
	input.mouse.logger = logger
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
	input.id_to_deps = make(map[KeyId][]Key, 16)
	input.index_to_agg_type = make(map[KeyIndex]aggregatorType)
	input.index_to_name = make(map[KeyIndex]string)
	input.index_to_family_deps = make(map[KeyIndex][]derivedKeyFamily)
	input.index_to_family = make(map[KeyIndex]derivedKeyFamily)
	input.SetLogger(logger)
	input.mouse.logger = logger

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

	// input.bindDerivedKeyWithId("Shift", EitherShift, input.MakeBinding(LeftShift, nil, nil), input.MakeBinding(RightShift, nil, nil))
	// input.bindDerivedKeyWithId("Control", EitherControl, input.MakeBinding(LeftControl, nil, nil), input.MakeBinding(RightControl, nil, nil))
	// input.bindDerivedKeyWithId("Alt", EitherAlt, input.MakeBinding(LeftAlt, nil, nil), input.MakeBinding(RightAlt, nil, nil))
	// input.bindDerivedKeyWithId("Gui", EitherGui, input.MakeBinding(LeftGui, nil, nil), input.MakeBinding(RightGui, nil, nil))
	// input.bindDerivedKeyWithId("ShiftTab", ShiftTab, input.MakeBinding(Tab, []KeyId{EitherShift}, []bool{true}))
	// input.bindDerivedKeyWithId("DeleteOrBackspace", DeleteOrBackspace, input.MakeBinding(KeyDelete, nil, nil), input.MakeBinding(Backspace, nil, nil))
	return input
}

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

// An EventGroup is a series of events that were all created by a single
// OsEvent.
type EventGroup struct {
	Events    []Event
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

func (input *Input) registerKeyIndex(index KeyIndex, agg_type aggregatorType, name string) {
	input.logger.Trace("gin.Input")
	if index < 0 {
		panic(fmt.Sprintf("Cannot register a key with index %d, indexes must be greater than 0.", index))
	}
	if prev, ok := input.index_to_name[index]; ok {
		panic(fmt.Sprintf("Cannot register key index %d, it has already been registered with the name %s and aggregator %v.", index, prev, agg_type))
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
	if id.Device.Type >= DeviceTypeMax || id.Device.Type < 0 {
		panic(fmt.Sprintf("Specified invalid DeviceType, %d.", id.Device))
	}
	key, ok := input.key_map[id]
	if !ok {
		if family, ok := input.index_to_family[id.Index]; ok {
			// If the index indicates a family but the key doesn't exist, go ahead
			// and have the family create it.
			input.key_map[id] = family.GetKey(id.Device)
			key = input.key_map[id]

			// TODO(tmckee): there are three blocks here and they all add a key to
			// input.all_keys, but this one does it implicitly through
			// family.GetKey(). We should find a way to avoid this and have all
			// additions to all_keys be in the same place.
			// input.all_keys = append(input.all_keys, key)
		} else if id.Index == AnyKey || id.Device.Type == DeviceTypeAny || id.Device.Index == DeviceIndexAny {
			// If we're looking for a general key we know how to create those
			if id.Device.Type == DeviceTypeAny && id.Device.Index != DeviceIndexAny {
				panic("Cannot specify a Device Index but not a Device Type.")
			}
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
				panic(fmt.Sprintf("No key registered with id == %v.", id))
			}
			var agg aggregator
			switch agg_type {
			case aggregatorTypeStandard:
				agg = &standardAggregator{}
			case aggregatorTypeAxis:
				agg = &axisAggregator{}
			case aggregatorTypeWheel:
				agg = &wheelAggregator{}
			default:
				panic(fmt.Sprintf("Unknown aggregator type specified: %T.", agg_type))
			}
			input.key_map[id] = &keyState{
				id:         id,
				name:       input.index_to_name[id.Index],
				aggregator: agg,
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

// Look for Keys related to the event's Key and notify them as needed.
func (input *Input) informDeps(event Event, group *EventGroup) {
	input.logger.Trace("gin.Input")
	id := event.Key.Id()

	id_ignoring_device_index := id
	id_ignoring_device_index.Device.Index = DeviceIndexAny

	// Direct dependencies are recorded in input.id_to_deps
	keysToPress := input.id_to_deps[id]

	// TODO(tmckee): consider using a set instead of a list for the keys to
	// press... if presses are idempotent, we don't need to press them again (so
	// don't bother walking them) _OR_ if they're not idempotent, how in the heck
	// would we manage to get a reasonable "press amount" after all this!?

	// Dependencies for keys organized by the same 'KeyIndex' but not pinned to a
	// particular device instance, though the 'DeviceType' does need to match
	// 🤔...
	for _, dep := range input.id_to_deps[id_ignoring_device_index] {
		keysToPress = append(keysToPress, dep)
	}

	// Skip over notifying relevant 'key families' for derived keys (why???) or
	// keys that do not distinguish between device instances (why???).
	if id.Device.Type != DeviceTypeDerived && id.Device.Index != DeviceIndexAny {
		// Select each {key-that-has-multiple-triggers} that should be triggered by
		// the current key for pressing.
		for _, family_dep := range input.index_to_family_deps[id.Index] {
			key := family_dep.GetKey(id.Device)
			keysToPress = append(keysToPress, key)
		}
	}
	for _, dep := range keysToPress {
		input.pressKey(dep, dep.CurPressAmt(), event, group)
	}
	if event.Type != NoEvent {
		group.Events = append(group.Events, event)
	}
}

func (input *Input) pressKey(k Key, amt float64, cause Event, group *EventGroup) {
	input.logger.Trace("gin.Input", "group.Events", group.Events)
	event := k.SetPressAmt(amt, group.Timestamp, cause)
	input.informDeps(event, group)

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

// The Input object can have a single Listener registered with it. This object
// will receive event groups as they are processed. During HandleEventGroup a
// listener can query keys as to their current state (i.e. with Cur*() methods)
// and these will accurately report their state given that the current event
// group has happened and no future events have happened yet.
//
// Frame*() methods on keys will report state from last frame.
//
// Listener.Think() will be called after all the events for a frame have been
// processed.
type EventHandler interface {
	HandleEventGroup(EventGroup)
}
type Listener interface {
	EventHandler
	Think(int64)
}
type EventDispatcher interface {
	RegisterEventListener(Listener)
}

func (input *Input) RegisterEventListener(listener Listener) {
	input.logger.Trace("gin.Input")
	input.listeners = append(input.listeners, listener)
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

		if os_event.KeyId.Device.Type == DeviceTypeMouse {
			input.mouse.Handle(os_event, &group)
		}

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
		gen, amt := key.Think(t)
		if !gen {
			continue
		}
		group := EventGroup{Timestamp: t}
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
