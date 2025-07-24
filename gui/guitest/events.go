package guitest

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gin/aggregator"
	"github.com/runningwild/glop/gui"
)

type synth struct {
	input *gin.Input
}

type dontCareType struct {
	MousePos  *gin.MousePosition
	Timestamp int64
}

var dontCare = dontCareType{
	MousePos: &gin.MousePosition{
		X: 24, Y: 42,
	},
	Timestamp: 1337,
}

func SynthesizeEvents() *synth {
	return &synth{
		input: gin.Make(),
	}
}

func (s *synth) WheelDown(amt float64) gui.EventGroup {
	keyId := gin.AnyMouseWheelVertical
	// Don't pick "any" mouse wheel; pick the first.
	keyId.Device.Index = 0

	wheelDownKey := s.input.GetKeyById(keyId)
	wheelDownKey.KeySetPressAmt(amt, 42, gin.Event{})
	wheelDownKey.KeyThink(42)
	return gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  wheelDownKey,
					Type: aggregator.Press,
				},
				{
					Key:  wheelDownKey,
					Type: aggregator.Release,
				},
			},
			Timestamp: dontCare.Timestamp,
		},
	}
}

func (s *synth) MouseMove(target gui.Point) []gui.EventGroup {
	wheelXAxis := s.input.GetKeyById(gin.KeyId{
		Index: gin.MouseXAxis,
		Device: gin.DeviceId{
			Index: 0,
			Type:  gin.DeviceTypeMouse,
		},
	})
	wheelYAxis := s.input.GetKeyById(gin.KeyId{
		Index: gin.MouseYAxis,
		Device: gin.DeviceId{
			Index: 0,
			Type:  gin.DeviceTypeMouse,
		},
	})

	wheelXAxis.KeySetPressAmt(-3, 42, gin.Event{})
	wheelYAxis.KeySetPressAmt(+4, 42, gin.Event{})

	xMove := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  wheelXAxis,
					Type: aggregator.Press,
				},
			},
		},
	}
	yMove := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  wheelYAxis,
					Type: aggregator.Press,
				},
			},
		},
	}
	xMove.SetMousePosition(target.X, target.Y)
	yMove.SetMousePosition(target.X, target.Y)

	return []gui.EventGroup{xMove, yMove}
}

func (s *synth) KeyDown(keyid gin.KeyId, at gui.Point) []gui.EventGroup {
	key := s.input.GetKeyById(keyid)
	key.KeySetPressAmt(1, 42, gin.Event{})

	keyDown := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  key,
					Type: aggregator.Press,
				},
			},
		},
	}
	keyDown.SetMousePosition(at.X, at.Y)

	return []gui.EventGroup{
		keyDown,
	}
}

func (s *synth) KeyUp(keyid gin.KeyId, at gui.Point) []gui.EventGroup {
	key := s.input.GetKeyById(keyid)
	key.KeySetPressAmt(0, 42, gin.Event{})

	keyUp := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  key,
					Type: aggregator.Release,
				},
			},
		},
	}
	keyUp.SetMousePosition(at.X, at.Y)

	return []gui.EventGroup{
		keyUp,
	}
}

func (s *synth) KeyDrag(buttonId gin.KeyId, fromPoint, toPoint gui.Point) []gui.EventGroup {
	totalGesture := []gui.EventGroup{}

	// Move to start
	totalGesture = append(totalGesture, s.MouseMove(fromPoint)...)

	// Mouse down
	totalGesture = append(totalGesture, s.KeyDown(buttonId, fromPoint)...)

	// Move to end
	totalGesture = append(totalGesture, s.MouseMove(toPoint)...)

	// Mouse up
	totalGesture = append(totalGesture, s.KeyUp(buttonId, toPoint)...)

	return totalGesture
}
