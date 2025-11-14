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
	mouseWheelKeyId := gin.AnyMouseWheelVertical
	mouseWheelKeyId.Device.Index = 0

	t.Run("WheelDown", func(t *testing.T) {
		assert := assert.New(t)
		synthesized := guitest.SynthesizeEvents().WheelDown(-42)

		assert.True(synthesized.IsPressed(mouseWheelKeyId))
		// The 'current press total' is just the running sum while we process a
		// frame. In between calls to "process the whole frame", the counter should
		// be 0.
		assert.Equal(float64(0), synthesized.PrimaryEvent().Key.CurPressTotal())
	})

	t.Run("dragging", func(t *testing.T) {
		fromPos := gui.PointAt(4, 4)
		toPos := gui.PointAt(7, 42)

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

	t.Run("can hook into 'Respond' phase", func(t *testing.T) {
		listener := guitest.NewRespondSpy()
		guitest.SynthesizeEvents(listener).WheelDown(5)

		events := listener.GetEvents()
		for _, evt := range events {
			if evt.PrimaryEvent().Key.Id() == mouseWheelKeyId {
				// Want more checks; make sure there's a Press and a Release(?)
				return
			}
		}

		t.Fatalf("didn't see events from the mouse wheel during the 'Respond' phase; saw %v", events)
	})
}
