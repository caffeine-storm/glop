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

	t.Run("IsPressed(some-key)", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		inputObj := gin.Make()

		specificKeyX := getKeyXForKeyboard0(inputObj)
		require.NotNil(specificKeyX)

		genericKeyX := getCorrespondingAnyDeviceKey(inputObj, specificKeyX)
		require.NotNil(genericKeyX)

		require.NotEqual(specificKeyX, genericKeyX)

		t.Run("supports 'any-device'", func(t *testing.T) {
			// A specific key gets pressed.
			eg := gin.EventGroup{
				Events: []gin.Event{
					{
						Key:  specificKeyX,
						Type: aggregator.Press,
					},
				},
				Timestamp: 32,
			}

			// Checking if the generic key is pressed should return true.
			assert.True(eg.IsPressed(genericKeyX.Id()))
		})
	})

	t.Run("IsPressed interactions with the 'any key'", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		inputObj := gin.Make()

		anyKey := inputObj.GetKeyById(gin.AnyAnyKey)
		require.NotNil(anyKey)

		specificKeyX := getKeyXForKeyboard0(inputObj)
		require.NotNil(specificKeyX)

		keyYId := specificKeyX.Id()
		keyYId.Index = gin.KeyY
		specificKeyY := inputObj.GetKeyById(keyYId)
		require.NotNil(specificKeyY)

		t.Run("an EventGroup denoting a key press should say 'the any has been pressed'", func(t *testing.T) {
			// A specific key gets pressed and an 'any' key is included as a
			// secondary event.
			eg := gin.EventGroup{
				Events: []gin.Event{
					{
						Key:  specificKeyX,
						Type: aggregator.Press,
					},
					{
						Key:  anyKey,
						Type: aggregator.Press,
					},
				},
				Timestamp: 32,
			}

			assert.True(eg.IsPressed(anyKey.Id()), "The 'any' key should look like it's pressed")

			assert.False(eg.IsPressed(specificKeyY.Id()), "A different key should look like it's not pressed")
		})
	})
}
