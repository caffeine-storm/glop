package guitest

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gin/aggregator"
	"github.com/runningwild/glop/gui"
)

type synth struct {
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
	return &synth{}
}

func (s *synth) WheelDown(amt float64) gui.EventGroup {
	in := gin.Make()

	keyId := gin.AnyMouseWheelVertical
	// Don't pick "any" mouse wheel; pick the first.
	keyId.Device.Index = 0

	wheelDownKey := in.GetKeyById(keyId)
	wheelDownKey.KeySetPressAmt(amt, 42, gin.Event{})
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
