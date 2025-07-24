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
		DispatchedToFocussedWidget: false,
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

func (s *synth) MouseDown(at gui.Point) []gui.EventGroup {
	leftMouseButton := s.input.GetKeyById(gin.KeyId{
		Index: gin.MouseLButton,
		Device: gin.DeviceId{
			Index: 0,
			Type:  gin.DeviceTypeMouse,
		},
	})
	leftMouseButton.KeySetPressAmt(1, 42, gin.Event{})

	mouseDown := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  leftMouseButton,
					Type: aggregator.Press,
				},
			},
		},
	}
	mouseDown.SetMousePosition(at.X, at.Y)

	return []gui.EventGroup{
		mouseDown,
	}
}

func (s *synth) MouseUp(at gui.Point) []gui.EventGroup {
	leftMouseButton := s.input.GetKeyById(gin.KeyId{
		Index: gin.MouseLButton,
		Device: gin.DeviceId{
			Index: 0,
			Type:  gin.DeviceTypeMouse,
		},
	})
	leftMouseButton.KeySetPressAmt(0, 42, gin.Event{})

	mouseUp := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				{
					Key:  leftMouseButton,
					Type: aggregator.Release,
				},
			},
		},
	}
	mouseUp.SetMousePosition(at.X, at.Y)

	return []gui.EventGroup{
		mouseUp,
	}
}

func (s *synth) MouseDrag(fromPoint, toPoint gui.Point) []gui.EventGroup {
	totalGesture := []gui.EventGroup{}

	// Move to start
	totalGesture = append(totalGesture, s.MouseMove(fromPoint)...)

	// Mouse down
	totalGesture = append(totalGesture, s.MouseDown(fromPoint)...)

	// Move to end
	totalGesture = append(totalGesture, s.MouseMove(toPoint)...)

	// Mouse up
	totalGesture = append(totalGesture, s.MouseUp(toPoint)...)

	return totalGesture
}
