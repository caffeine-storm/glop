package gin_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGinSpecs(t *testing.T) {
	Convey("Input", t, func() {
		Convey("NaturalKeySpec", NaturalKeySpec)
		Convey("DerivedKeySpec", DerivedKeySpec)
		Convey("DeviceSpec", DeviceSpec)
		Convey("NestedDerivedKeySpec", NestedDerivedKeySpec)
		Convey("EventSpec", EventSpec)
		Convey("AxisSpec", AxisSpec)
		Convey("EventListenerSpec", EventListenerSpec)
	})
}

func injectEvent(events *[]gin.OsEvent, key_index gin.KeyIndex, device_index gin.DeviceIndex, device_type gin.DeviceType, amt float64, timestamp int64) {
	*events = append(*events,
		gin.OsEvent{
			KeyId: gin.KeyId{
				Index: key_index,
				Device: gin.DeviceId{
					Index: device_index,
					Type:  device_type,
				},
			},
			Press_amt: amt,
			Timestamp: timestamp,
		},
	)
}

var dontCare = struct {
	Amount    float64
	Timestamp int64
}{
	Amount:    1337,
	Timestamp: 1337,
}

func injectMouseClick(events *[]gin.OsEvent, target gui.Point) {
	clickEvent := gin.OsEvent{
		KeyId: gin.KeyId{
			Index: gin.MouseLButton,
			Device: gin.DeviceId{
				Index: 0, // First mouse
				Type:  gin.DeviceTypeMouse,
			},
		},
		X:         target.X,
		Y:         target.Y,
		Press_amt: dontCare.Amount,
		Timestamp: dontCare.Timestamp,
	}
	*events = append(*events, clickEvent)
}

type testMouseListener struct {
	clicks []gui.Point
}

var _ gin.Listener = (*testMouseListener)(nil)

func (self *testMouseListener) HandleEventGroup(group gin.EventGroup) {
	for _, evt := range group.Events {
		kid := evt.Key.Id()
		if kid.Index == gin.MouseLButton {
			pt := gui.Point{
				X: group.X,
				Y: group.Y,
			}
			fmt.Printf("X: %d Y: %d\n", group.X, group.Y)
			self.clicks = append(self.clicks, pt)
			return
		}
	}
}
func (*testMouseListener) Think(int64) {}

func (self *testMouseListener) ExpectClicks(pts []gui.Point) {
	if self.clicks == nil {
		self.clicks = []gui.Point{}
	}
	So(pts, ShouldEqual, self.clicks)
}

func NaturalKeySpec() {
	input := gin.Make()
	keya := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyByParts(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	Convey("Single key press or release per frame sets basic keyState values properly.", func() {

		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 5)
		input.Think(10, events)
		So(keya.FramePressCount(), ShouldEqual, 1)
		So(keya.FrameReleaseCount(), ShouldEqual, 0)
		So(keyb.FramePressCount(), ShouldEqual, 0)
		So(keyb.FrameReleaseCount(), ShouldEqual, 0)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 15)
		input.Think(20, events)
		So(keya.FramePressCount(), ShouldEqual, 0)
		So(keya.FrameReleaseCount(), ShouldEqual, 0)
		So(keyb.FramePressCount(), ShouldEqual, 1)
		So(keyb.FrameReleaseCount(), ShouldEqual, 0)

		events = events[0:0]
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 25)
		input.Think(30, events)
		So(keya.FramePressCount(), ShouldEqual, 0)
		So(keya.FrameReleaseCount(), ShouldEqual, 1)
		So(keyb.FramePressCount(), ShouldEqual, 0)
		So(keyb.FrameReleaseCount(), ShouldEqual, 0)
	})

	Convey("Multiple key presses in a single frame work.", func() {
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 5)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 6)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 7)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 8)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 9)
		input.Think(10, events)
		So(keya.FramePressCount(), ShouldEqual, 3)
		So(keya.FrameReleaseCount(), ShouldEqual, 2)
		So(keyb.FramePressCount(), ShouldEqual, 1)
		So(keyb.FrameReleaseCount(), ShouldEqual, 0)
	})

	Convey("Redundant events don't generate redundant events.", func() {
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 5)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 6)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 7)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 8)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 9)
		input.Think(10, events)
		So(keya.FramePressCount(), ShouldEqual, 1)
		So(keya.FrameReleaseCount(), ShouldEqual, 1)
	})

	Convey("Key.FramePressSum() works.", func() {
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 3)
		input.Think(10, events)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 14)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 16)
		input.Think(20, events)
		So(keya.FramePressSum(), ShouldEqual, 8.0)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 22)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 24)
		input.Think(30, events)
		So(keyb.FramePressSum(), ShouldEqual, 2.0)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 35)
		input.Think(40, events)
		So(keyb.FramePressSum(), ShouldEqual, 5.0)
	})

	Convey("Key.FramePressAvg() works.", func() {
		events := make([]gin.OsEvent, 0)
		input.Think(10, events)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 10)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 12)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 14)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 16)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 18)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 20)
		input.Think(20, events)
		So(keya.FramePressAvg(), ShouldEqual, 0.6)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 25)
		input.Think(30, events)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 35)
		input.Think(40, events)
		So(keyb.FramePressAvg(), ShouldEqual, 0.5)
	})
}

func DerivedKeySpec() {
	input := gin.Make()
	keya := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyByParts(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	keyc := input.GetKeyByParts(gin.KeyC, gin.DeviceTypeKeyboard, 1)
	keye := input.GetKeyByParts(gin.KeyE, gin.DeviceTypeKeyboard, 1)
	keyf := input.GetKeyByParts(gin.KeyF, gin.DeviceTypeKeyboard, 1)
	ABc_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id(), keyc.Id()}, []bool{true, false})
	Ef_binding := input.MakeBinding(keye.Id(), []gin.KeyId{keyf.Id()}, []bool{false})
	ABc_Ef := input.BindDerivedKey("ABc_Ef", ABc_binding, Ef_binding)

	// ABc_Ef should be down if either ab and not c, or e and not f (or both)
	// That is to say the following (and no others) should all trigger it:
	// (Capital letter indicates the key is down, lowercase indicate it is not)
	// A B c e f
	// A B c e F
	// A B c E f
	// A B c E F
	// a b c E f
	// a b C E f
	// a B c E f
	// a B C E f
	// A b c E f
	// A b C E f
	// A B C E f

	Convey("Derived key presses happen only when a primary key is pressed after all modifiers are set.", func() {

		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(10, events)
		So(ABc_Ef.FramePressAmt(), ShouldEqual, 1.0)
		So(ABc_Ef.IsDown(), ShouldEqual, true)
		So(ABc_Ef.FramePressCount(), ShouldEqual, 1)
		events = events[0:0]

		Convey("Release happens once primary key is released", func() {
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 11)
			input.Think(20, events)
			So(ABc_Ef.IsDown(), ShouldEqual, false)
			So(ABc_Ef.FramePressCount(), ShouldEqual, 0)
			So(ABc_Ef.FrameReleaseCount(), ShouldEqual, 1)
		})

		Convey("Key remains down when when a down modifier is released", func() {
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 11)
			input.Think(20, events)
			So(ABc_Ef.IsDown(), ShouldEqual, true)
			So(ABc_Ef.FramePressCount(), ShouldEqual, 0)
			So(ABc_Ef.FrameReleaseCount(), ShouldEqual, 0)
		})

		Convey("Key remains down when an up modifier is pressed", func() {
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 11)
			input.Think(20, events)
			So(ABc_Ef.IsDown(), ShouldEqual, true)
			So(ABc_Ef.FramePressCount(), ShouldEqual, 0)
			So(ABc_Ef.FrameReleaseCount(), ShouldEqual, 0)
		})
		Convey("Release isn't affect by bindings changing states first", func() {
			Convey("releasing b", func() {
				injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 11)
				injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 11)
				input.Think(20, events)
				So(ABc_Ef.IsDown(), ShouldEqual, false)
				So(ABc_Ef.FramePressCount(), ShouldEqual, 0)
				So(ABc_Ef.FrameReleaseCount(), ShouldEqual, 1)
			})
			Convey("pressing c", func() {
				injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 11)
				injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 11)
				input.Think(20, events)
				So(ABc_Ef.IsDown(), ShouldEqual, false)
				So(ABc_Ef.FramePressCount(), ShouldEqual, 0)
				So(ABc_Ef.FrameReleaseCount(), ShouldEqual, 1)
			})
		})

		Convey("Pressing a second binding should not generate another press on the derived key", func() {
			injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 1, 11)
			input.Think(20, events)
			So(ABc_Ef.IsDown(), ShouldEqual, true)
			So(ABc_Ef.FramePressCount(), ShouldEqual, 0)
			So(ABc_Ef.FrameReleaseCount(), ShouldEqual, 0)
		})

		// Reset keys
		events = events[0:0]
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 21)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 21)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 0, 21)
		injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 0, 21)
		input.Think(30, events)
		So(ABc_Ef.FramePressAmt(), ShouldEqual, 0.0)
		So(ABc_Ef.IsDown(), ShouldEqual, false)

		// Test that second binding can trigger a press
		events = events[0:0]
		injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 1, 31)
		input.Think(40, events)
		So(ABc_Ef.FramePressAmt(), ShouldEqual, 1.0)
		So(ABc_Ef.IsDown(), ShouldEqual, true)
		So(ABc_Ef.FramePressCount(), ShouldEqual, 1)

		// Reset keys
		events = events[0:0]
		injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 0, 41)
		input.Think(50, events)

		// Test that first binding doesn't trigger a press if modifiers aren't set first
		events = events[0:0]
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 51)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 51)
		input.Think(60, events)
		So(ABc_Ef.IsDown(), ShouldEqual, false)
		So(ABc_Ef.FramePressCount(), ShouldEqual, 0)
	})
}

// Check that derived keys can properly differentiate between the same key
// pressed on different devices.
func DeviceSpec() {
	input := gin.Make()
	keya1 := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keya2 := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 2)
	keya3 := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 3)
	A1_binding := input.MakeBinding(keya1.Id(), nil, nil)
	A2_binding := input.MakeBinding(keya2.Id(), nil, nil)
	A3_binding := input.MakeBinding(keya3.Id(), nil, nil)
	A1 := input.BindDerivedKey("A1", A1_binding)
	A2 := input.BindDerivedKey("A2", A2_binding)
	A3 := input.BindDerivedKey("A3", A3_binding)

	keya_any := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, gin.DeviceIndexAny)
	AAny_binding := input.MakeBinding(keya_any.Id(), nil, nil)
	AAny := input.BindDerivedKey("AAny", AAny_binding)
	keyb_any := input.GetKeyByParts(gin.KeyB, gin.DeviceTypeKeyboard, gin.DeviceIndexAny)
	BAny_binding := input.MakeBinding(keyb_any.Id(), nil, nil)
	BAny := input.BindDerivedKey("BAny", BAny_binding)

	any_key_on_1 := input.GetKeyByParts(gin.AnyKey, gin.DeviceTypeKeyboard, 1)
	any_key_on_1_binding := input.MakeBinding(any_key_on_1.Id(), nil, nil)
	Any1 := input.BindDerivedKey("Any1", any_key_on_1_binding)
	any_key_on_2 := input.GetKeyByParts(gin.AnyKey, gin.DeviceTypeKeyboard, 2)
	any_key_on_2_binding := input.MakeBinding(any_key_on_2.Id(), nil, nil)
	Any2 := input.BindDerivedKey("Any2", any_key_on_2_binding)
	any_key_on_3 := input.GetKeyByParts(gin.AnyKey, gin.DeviceTypeKeyboard, 3)
	any_key_on_3_binding := input.MakeBinding(any_key_on_3.Id(), nil, nil)
	Any3 := input.BindDerivedKey("Any3", any_key_on_3_binding)

	the_any_key := input.GetKeyByParts(gin.AnyKey, gin.DeviceTypeAny, gin.DeviceIndexAny)
	the_any_key_binding := input.MakeBinding(the_any_key.Id(), nil, nil)
	Any_key := input.BindDerivedKey("Any Key", the_any_key_binding)

	Convey("Derived keys trigger from the specified devices and no others.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, events)
		So(A1.IsDown(), ShouldEqual, true)
		So(A2.IsDown(), ShouldEqual, false)
		So(A3.IsDown(), ShouldEqual, false)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 3)
		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 1, 4)
		input.Think(5, events)
		So(A1.IsDown(), ShouldEqual, false)
		So(A2.IsDown(), ShouldEqual, true)
		So(A3.IsDown(), ShouldEqual, false)

		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 0, 6)
		injectEvent(&events, gin.KeyA, 3, gin.DeviceTypeKeyboard, 1, 7)
		input.Think(8, events)
		So(A1.IsDown(), ShouldEqual, false)
		So(A2.IsDown(), ShouldEqual, false)
		So(A3.IsDown(), ShouldEqual, true)
		events = events[0:0]
	})

	Convey("Derived key can specify a specific key on any device.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, events)
		So(AAny.IsDown(), ShouldEqual, true)
		So(BAny.IsDown(), ShouldEqual, false)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 3)
		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 1, 4)
		input.Think(5, events)
		So(AAny.IsDown(), ShouldEqual, true)
		So(BAny.IsDown(), ShouldEqual, false)

		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 0, 6)
		injectEvent(&events, gin.KeyB, 1, gin.DeviceTypeKeyboard, 1, 7)
		input.Think(8, events)
		So(AAny.IsDown(), ShouldEqual, false)
		So(BAny.IsDown(), ShouldEqual, true)

		injectEvent(&events, gin.KeyB, 1, gin.DeviceTypeKeyboard, 0, 9)
		injectEvent(&events, gin.KeyB, 2, gin.DeviceTypeKeyboard, 1, 10)
		input.Think(11, events)
		So(AAny.IsDown(), ShouldEqual, false)
		So(BAny.IsDown(), ShouldEqual, true)
		events = events[0:0]
	})

	Convey("Derived key can specify any key on a specific device.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, events)
		So(Any1.IsDown(), ShouldEqual, true)
		So(Any2.IsDown(), ShouldEqual, false)
		So(Any3.IsDown(), ShouldEqual, false)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 3)
		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 1, 3)
		input.Think(4, events)
		So(Any1.IsDown(), ShouldEqual, false)
		So(Any2.IsDown(), ShouldEqual, true)
		So(Any3.IsDown(), ShouldEqual, false)

		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 0, 5)
		injectEvent(&events, gin.KeyA, 3, gin.DeviceTypeKeyboard, 1, 5)
		input.Think(6, events)
		So(Any1.IsDown(), ShouldEqual, false)
		So(Any2.IsDown(), ShouldEqual, false)
		So(Any3.IsDown(), ShouldEqual, true)

		events = events[0:0]
	})

	Convey("Derived key can specify the any key.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, events)
		So(Any_key.IsDown(), ShouldEqual, true)
		So(Any_key.FramePressCount(), ShouldEqual, 1)
		So(Any_key.FrameReleaseCount(), ShouldEqual, 0)

		injectEvent(&events, gin.KeyB, 2, gin.DeviceTypeKeyboard, 1, 3)
		input.Think(4, events)
		So(Any_key.IsDown(), ShouldEqual, true)
		So(Any_key.FramePressCount(), ShouldEqual, 0)
		So(Any_key.FrameReleaseCount(), ShouldEqual, 0)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 5)
		input.Think(6, events)
		So(Any_key.IsDown(), ShouldEqual, true)
		So(Any_key.FramePressCount(), ShouldEqual, 0)
		So(Any_key.FrameReleaseCount(), ShouldEqual, 0)

		injectEvent(&events, gin.KeyB, 2, gin.DeviceTypeKeyboard, 0, 7)
		input.Think(8, events)
		So(Any_key.IsDown(), ShouldEqual, false)
		So(Any_key.FramePressCount(), ShouldEqual, 0)
		So(Any_key.FrameReleaseCount(), ShouldEqual, 1)
		events = events[0:0]
	})
}

func NestedDerivedKeySpec() {
	input := gin.Make()
	keya := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyByParts(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	keyc := input.GetKeyByParts(gin.KeyC, gin.DeviceTypeKeyboard, 1)
	AB_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id()}, []bool{true})
	AB := input.BindDerivedKey("AB", AB_binding)
	AB_C_binding := input.MakeBinding(keyc.Id(), []gin.KeyId{AB.Id()}, []bool{true})
	AB_C := input.BindDerivedKey("AB_C", AB_C_binding)
	events := make([]gin.OsEvent, 0)

	check := func(order string) {
		input.Think(10, events)
		if strings.Index(order, "b") < strings.Index(order, "a") {
			So(AB.IsDown(), ShouldEqual, true)
			So(AB.FramePressCount(), ShouldEqual, 1)
		} else {
			So(AB.IsDown(), ShouldEqual, false)
			So(AB.FramePressCount(), ShouldEqual, 0)
		}
		if order == "bac" {
			So(AB_C.IsDown(), ShouldEqual, true)
			So(AB_C.FramePressCount(), ShouldEqual, 1)
		} else {
			So(AB_C.IsDown(), ShouldEqual, false)
			So(AB_C.FramePressCount(), ShouldEqual, 0)
		}
	}

	Convey("Nested derived keys work like normal derived keys.", func() {
		Convey("Testing order 'bac'.", func() {
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("bac")
		})
		Convey("Testing order 'abc'.", func() {
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("abc")
		})
		Convey("Testing order 'acb'.", func() {
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("acb")
		})
		Convey("Testing order 'bca'.", func() {
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("bca")
		})
		Convey("Testing order 'cab'.", func() {
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("cab")
		})
		Convey("Testing order 'cba'.", func() {
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("cba")
		})
	})
}

func EventSpec() {
	input := gin.Make()
	keya := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyByParts(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	keyc := input.GetKeyByParts(gin.KeyC, gin.DeviceTypeKeyboard, 1)
	keyd := input.GetKeyByParts(gin.KeyD, gin.DeviceTypeKeyboard, 1)
	AB_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id()}, []bool{true})
	AB := input.BindDerivedKey("AB", AB_binding)
	CD_binding := input.MakeBinding(keyc.Id(), []gin.KeyId{keyd.Id()}, []bool{true})
	CD := input.BindDerivedKey("CD", CD_binding)
	AB_CD_binding := input.MakeBinding(AB.Id(), []gin.KeyId{CD.Id()}, []bool{true})
	_ = input.BindDerivedKey("AB CD", AB_CD_binding)
	events := make([]gin.OsEvent, 0)

	check := func(lengths ...int) {
		groups := input.Think(10, events)
		So(len(groups), ShouldEqual, len(lengths))
		for i, length := range lengths {
			// To make it more clear what each test is doing we only check for the
			// number of natural key and derived key events generated on each
			// press/release, i.e. we don't count general key events.
			natural_events := 0
			for _, event := range groups[i].Events {
				if event.Key.Id().Index == gin.AnyKey ||
					event.Key.Id().Device.Type == gin.DeviceTypeAny ||
					event.Key.Id().Device.Index == gin.DeviceIndexAny {
					continue
				}
				natural_events++
			}
			So(natural_events, ShouldEqual, length)
		}
	}

	Convey("Testing order 'abcd'.", func() {
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 1, 1, 1)
	})

	Convey("Testing order 'dbca'.", func() {
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 1, 2, 3)
	})

	Convey("Testing order 'dcba'.", func() {
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 2, 1, 3)
	})

	Convey("Testing order 'bcda'.", func() {
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 1, 1, 2)
	})

	// This test also checks that a derived key stays down until the primary key
	// is released CD is used here after D is released to trigger AB_CD.
	Convey("Testing order 'dcbDad'.", func() {
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 0, 4)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 5)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 6)
		check(1, 2, 1, 1, 3, 1)
	})
}

func AxisSpec() {
	input := gin.Make()

	// TODO: This is the mouse x axis key, we need a constant for this or
	// something.
	x := input.GetKeyByParts(gin.MouseXAxis, gin.DeviceTypeMouse, 1)
	events := make([]gin.OsEvent, 0)

	Convey("Axes aggregate press amts and report IsDown() properly.", func() {
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, 1, 5)
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, 10, 6)
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, -3, 7)
		input.Think(10, events)
		So(x.FramePressAmt(), ShouldEqual, -3.0)
		So(x.FramePressSum(), ShouldEqual, 8.0)
	})

	Convey("Axes can sum to zero and still be down.", func() {
		input.Think(0, events)
		events = events[0:0]
		So(x.FramePressSum(), ShouldEqual, 0.0)
		So(x.IsDown(), ShouldEqual, false)

		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, 5, 5)
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, -5, 6)
		input.Think(10, events)
		events = events[0:0]
		So(x.FramePressSum(), ShouldEqual, 0.0)
		So(x.IsDown(), ShouldEqual, true)

		input.Think(20, events)
		So(x.FramePressSum(), ShouldEqual, 0.0)
		So(x.IsDown(), ShouldEqual, false)
	})
}

type listener struct {
	input  *gin.Input
	key_id gin.KeyId

	press_count   []int
	release_count []int
	press_amt     []float64
}

func (l *listener) ExpectPressCounts(v ...int) {
	l.press_count = v
}
func (l *listener) ExpectReleaseCounts(v ...int) {
	l.release_count = v
}
func (l *listener) ExpectPressAmts(v ...float64) {
	l.press_amt = v
}
func (l *listener) HandleEventGroup(eg gin.EventGroup) {
	k := l.input.GetKeyById(l.key_id)
	So(k.CurPressCount(), ShouldEqual, l.press_count[0])
	So(k.CurReleaseCount(), ShouldEqual, l.release_count[0])
	So(k.CurPressAmt(), ShouldEqual, l.press_amt[0])
	l.press_count = l.press_count[1:]
	l.release_count = l.release_count[1:]
	l.press_amt = l.press_amt[1:]
}
func (l *listener) Think(ms int64) {
	So(len(l.press_count), ShouldEqual, 0)
	So(len(l.release_count), ShouldEqual, 0)
	So(len(l.press_amt), ShouldEqual, 0)
}

func EventListenerSpec() {
	input := gin.Make()
	keya := input.GetKeyByParts(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyByParts(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	AB_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id()}, []bool{true})
	AB := input.BindDerivedKey("AB", AB_binding)
	events := make([]gin.OsEvent, 0)

	Convey("Check keys report state properly while handling events", func() {
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 2)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 5)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 6)

		Convey("Test a", func() {
			la := &listener{
				input:  input,
				key_id: keya.Id(),
			}
			input.RegisterEventListener(la)
			la.ExpectPressCounts(1, 1, 1, 2, 2, 2)
			la.ExpectReleaseCounts(0, 1, 1, 1, 1, 2)
			la.ExpectPressAmts(1, 0, 0, 1, 1, 0)
			input.Think(0, events)
		})
		Convey("Test b", func() {
			lb := &listener{
				input:  input,
				key_id: keyb.Id(),
			}
			input.RegisterEventListener(lb)
			lb.ExpectPressCounts(0, 0, 1, 1, 1, 1)
			lb.ExpectReleaseCounts(0, 0, 0, 0, 1, 1)
			lb.ExpectPressAmts(0, 0, 1, 1, 0, 0)
			input.Think(0, events)
		})
		Convey("Test ab", func() {
			lab := &listener{
				input:  input,
				key_id: AB.Id(),
			}
			input.RegisterEventListener(lab)
			lab.ExpectPressCounts(0, 0, 0, 1, 1, 1)
			lab.ExpectReleaseCounts(0, 0, 0, 0, 0, 1)
			lab.ExpectPressAmts(0, 0, 0, 1, 1, 0)
			input.Think(0, events)
		})
	})
	Convey("Event Listeners get notified of mouse events", func() {
		l := &testMouseListener{}
		input.RegisterEventListener(l)

		clickPoint := gui.Point{
			X: 17,
			Y: 42,
		}

		events := events[0:0]
		injectMouseClick(&events, clickPoint)

		// No clicks yet because we haven't input.Think'd
		l.ExpectClicks([]gui.Point{})

		input.Think(0, events)

		l.ExpectClicks([]gui.Point{
			clickPoint,
		})
	})
}
