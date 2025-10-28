package gin_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gin/aggregator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getKeyXForKeyboard0(in *gin.Input) gin.Key {
	specificId := gin.KeyId{
		Index: gin.KeyX,
		Device: gin.DeviceId{
			Index: 0,
			Type:  gin.DeviceTypeKeyboard,
		},
	}

	return in.GetKeyById(specificId)
}

func getCorrespondingAnyDeviceKey(in *gin.Input, referenceKey gin.Key) gin.Key {
	genericId := referenceKey.Id()
	genericId.Device.Index = gin.DeviceIndexAny

	return in.GetKeyById(genericId)
}

type testcasedata struct {
	name          string
	eventType     aggregator.EventType
	funcUnderTest func(gin.EventGroup, gin.KeyId) bool
}

func TestEventGroup(t *testing.T) {
	t.Run("event groups have optional x-y co-ordinates", func(t *testing.T) {
		eg := gin.EventGroup{}
		if eg.HasMousePosition() {
			t.Fatalf("zero-valued groups must have non-existent co-ordinates")
		}

		eg.SetMousePosition(14, 44)

		if !eg.HasMousePosition() {
			t.Fatalf("there must be a position after setting one")
		}

		x, y := eg.GetMousePosition()
		if x != 14 || y != 44 {
			t.Fatalf("reading the mouse position should return whatever was set")
		}
	})
	t.Run("events have a useful API", func(t *testing.T) {
		ev := gin.Event{}
		ev.IsPress()
		ev.IsRelease()
	})
	t.Run("EventGroup.String() should be useful", func(t *testing.T) {
		eg := gin.EventGroup{}

		stringified := fmt.Sprintf("%v", eg)
		if !strings.Contains(stringified, "mousePos: nil") {
			t.Fatalf("non-mouse event groups need to report 'nil' for mouse position but got %q", stringified)
		}

		eg.SetMousePosition(14, 44)
		stringified = fmt.Sprintf("%v", eg)

		if !strings.Contains(stringified, "14 44") {
			t.Fatalf("a mouse-event group should show the mouse x/y when stringified but got %q", stringified)
		}

		pointerStringified := fmt.Sprintf("%v", &eg)
		if stringified != pointerStringified {
			t.Fatalf("format difference when showing value vs pointer-to-value: %q %q", stringified, pointerStringified)
		}
	})

	testtable := []testcasedata{
		{
			name:      "IsPressed",
			eventType: aggregator.Press,
			funcUnderTest: func(grp gin.EventGroup, kid gin.KeyId) bool {
				return grp.IsPressed(kid)
			},
		},
		{
			name:      "IsReleased",
			eventType: aggregator.Release,
			funcUnderTest: func(grp gin.EventGroup, kid gin.KeyId) bool {
				return grp.IsReleased(kid)
			},
		},
	}

	for _, testcase := range testtable {
		t.Run(testcase.name, func(t *testing.T) {
			t.Run("some-key", func(t *testing.T) {
				assert := assert.New(t)
				require := require.New(t)

				inputObj := gin.Make()

				anyKey := inputObj.GetKeyById(gin.AnyAnyKey)
				require.NotNil(anyKey)

				specificKeyX := getKeyXForKeyboard0(inputObj)
				require.NotNil(specificKeyX)

				xkeyId := specificKeyX.Id()
				specificKeyY := inputObj.GetKeyByParts(gin.KeyY, xkeyId.Device.Type, xkeyId.Device.Index)
				require.NotNil(specificKeyX)

				genericKeyX := getCorrespondingAnyDeviceKey(inputObj, specificKeyX)
				require.NotNil(genericKeyX)

				require.NotEqual(specificKeyX, genericKeyX)

				t.Run("supports 'any-device'", func(t *testing.T) {
					// A specific key gets used.
					eg := gin.EventGroup{
						Events: []gin.Event{
							{
								Key:  specificKeyX,
								Type: testcase.eventType,
							},
						},
						TimestampMs: 32,
					}

					// Check that the generic key is in the right state.
					assert.True(testcase.funcUnderTest(eg, genericKeyX.Id()))
				})
				t.Run("an EventGroup denoting a key event should say 'the any has that event'", func(t *testing.T) {
					// If a specific key gets toggled, an 'any' key is included as a
					// secondary event.
					eg := gin.EventGroup{
						Events: []gin.Event{
							{
								Key:  specificKeyX,
								Type: testcase.eventType,
							},
							{
								Key:  anyKey,
								Type: testcase.eventType,
							},
						},
						TimestampMs: 32,
					}

					assert.True(testcase.funcUnderTest(eg, anyKey.Id()), "The 'any' key should look like it's been toggled")
					assert.False(testcase.funcUnderTest(eg, specificKeyY.Id()), "A different key should look like it's not involved")
				})
			})
		})
	}

	t.Run("IsMouseMove", func(t *testing.T) {
		input := gin.Make()
		xAxis := input.GetKeyByParts(gin.MouseXAxis, gin.DeviceTypeMouse, 0)
		require.NotNil(t, xAxis)

		eg := gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  xAxis,
					Type: aggregator.Adjust,
				},
			},
		}
		eg.SetMousePosition(13, 42)
		t.Run("returns true for a mouse move", func(t *testing.T) {
			assert.True(t, eg.IsMouseMove())
		})

		eg.Events[0].Type = aggregator.Press
		t.Run("returns false for a press", func(t *testing.T) {
			assert.False(t, eg.IsMouseMove())
		})

		eg.Events[0].Type = aggregator.Release
		t.Run("returns false for a release", func(t *testing.T) {
			assert.False(t, eg.IsMouseMove())
		})

		t.Run("calling KeySetPressAmt on an axis key will return a mouse move", func(t *testing.T) {
			result := xAxis.KeySetPressAmt(42, dontCare.Timestamp, dontCare.NoEvent)
			assert.Equal(t, aggregator.Adjust, result.Type)
		})
	})
}
