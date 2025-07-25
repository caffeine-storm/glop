package guitest_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gin/aggregator"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/gui/guitest"
	"github.com/stretchr/testify/assert"
)

func findEvent(events []gui.EventGroup, pred func(gui.EventGroup) bool) gui.EventGroup {
	idx := -1
	for i := range events {
		if !pred(events[i]) {
			continue
		}

		if idx != -1 {
			panic(fmt.Errorf("more than one event matched the predicate: %d and %d", idx, i))
		}

		idx = i
	}

	if idx == -1 {
		panic(fmt.Errorf("no event matched the predicate"))
	}

	return events[idx]
}

func TestSynthesize(t *testing.T) {
	t.Run("WheelDown", func(t *testing.T) {
		assert := assert.New(t)
		synthesized := guitest.SynthesizeEvents().WheelDown(-42)

		mouseWheelKeyId := gin.AnyMouseWheelVertical
		mouseWheelKeyId.Device.Index = 0

		assert.True(synthesized.IsPressed(mouseWheelKeyId))
		assert.Equal(float64(-42), synthesized.PrimaryEvent().Key.FramePressTotal())
	})

	t.Run("dragging", func(t *testing.T) {
		fromPos := gui.Point{
			X: 4, Y: 4,
		}
		toPos := gui.Point{
			X: 7, Y: 42,
		}

		leftMouseButtonKeyId := gin.KeyId{
			Index: gin.MouseLButton,
			Device: gin.DeviceId{
				Index: 0,
				Type:  gin.DeviceTypeMouse,
			},
		}

		synthesized := guitest.SynthesizeEvents().DragGesture(leftMouseButtonKeyId, fromPos, toPos)

		numEvents := len(synthesized)
		assert.Greater(t, numEvents, 0, "there should be some events")

		mouseDown := findEvent(synthesized, func(ev gui.EventGroup) bool {
			if !ev.PrimaryEvent().IsPress() {
				return false
			}
			return ev.PrimaryEvent().Key.Id() == leftMouseButtonKeyId
		})
		assert.Equal(t, mouseDown.GetMousePosition(), fromPos)
		mouseUp := findEvent(synthesized, func(ev gui.EventGroup) bool {
			if !ev.PrimaryEvent().IsRelease() {
				return false
			}
			return ev.PrimaryEvent().Key.Id() == leftMouseButtonKeyId
		})
		assert.Equal(t, mouseUp.GetMousePosition(), toPos)

		t.Run("'Adjust' events are used for mouse movement", func(t *testing.T) {
			mouseMoveXToStart := findEvent(synthesized, func(ev gui.EventGroup) bool {
				if ev.PrimaryEvent().Key.Id().Index != gin.MouseXAxis {
					return false
				}
				return ev.GetMousePosition() == fromPos
			})
			assert.Equal(t, aggregator.Adjust, mouseMoveXToStart.PrimaryEvent().Type)
		})
	})
}
