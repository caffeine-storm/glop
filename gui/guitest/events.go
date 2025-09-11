package guitest

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
)

var xaxis = gin.KeyId{
	Index: gin.MouseXAxis,
	Device: gin.DeviceId{
		Index: 0,
		Type:  gin.DeviceTypeMouse,
	},
}

var yaxis = gin.KeyId{
	Index: gin.MouseYAxis,
	Device: gin.DeviceId{
		Index: 0,
		Type:  gin.DeviceTypeMouse,
	},
}

var wheel = gin.KeyId{
	Index: gin.MouseWheelVertical,
	Device: gin.DeviceId{
		Index: 0,
		Type:  gin.DeviceTypeMouse,
	},
}

type synth struct {
	input   *gin.Input
	rootGui *gui.Gui
	spies   []RespondSpy
}

type dontCareType struct {
	MousePos   *gin.MousePosition
	MousePoint gui.Point
	Timestamp  int64
	NoEvent    gin.Event
}

var dontCare = dontCareType{
	MousePos: &gin.MousePosition{
		X: 24, Y: 42,
	},
	MousePoint: gui.Point{
		X: 48, Y: 84,
	},
	Timestamp: 1337,
	NoEvent:   gin.Event{},
}

func (s *synth) emulateRespondPhase(eg gui.EventGroup) {
	for _, spy := range s.spies {
		spy.Respond(s.rootGui, eg)
	}
}

func (s *synth) makeEventGroup(keyid gin.KeyId, at gui.Point, pressAmt float64) gui.EventGroup {
	key := s.input.GetKeyById(keyid)
	evt := key.KeySetPressAmt(pressAmt, dontCare.Timestamp, dontCare.NoEvent)

	ret := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events: []gin.Event{
				evt,
			},
		},
	}
	ret.SetMousePosition(at.X, at.Y)

	return ret
}

func (s *synth) synthesizeEventGroup(keyid gin.KeyId, at gui.Point, pressAmt float64) gui.EventGroup {
	ret := s.makeEventGroup(keyid, at, pressAmt)
	s.emulateRespondPhase(ret)
	return ret
}

func (s *synth) press(keyid gin.KeyId, at gui.Point) gui.EventGroup {
	ret := s.synthesizeEventGroup(keyid, at, 1)
	ret.Events[0].Key.KeyThink(dontCare.Timestamp)
	return ret
}

func (s *synth) release(keyid gin.KeyId, at gui.Point) gui.EventGroup {
	ret := s.synthesizeEventGroup(keyid, at, 0)
	ret.Events[0].Key.KeyThink(dontCare.Timestamp)
	return ret
}

func SynthesizeEvents(listeners ...RespondSpy) *synth {
	return &synth{
		input:   gin.Make(),
		rootGui: MakeStubbedGui(gui.Dims{Dx: 16, Dy: 16}),
		spies:   listeners,
	}
}

func (s *synth) WheelDown(amt float64) gui.EventGroup {
	start := s.synthesizeEventGroup(wheel, dontCare.MousePoint, amt)
	end := s.synthesizeEventGroup(wheel, dontCare.MousePoint, 0)

	events := []gin.Event{}
	events = append(events, start.Events...)
	events = append(events, end.Events...)

	eg := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events:    events,
			Timestamp: dontCare.Timestamp,
		},
	}
	eg.SetMousePosition(dontCare.MousePoint.X, dontCare.MousePoint.Y)

	eg.Events[0].Key.KeyThink(dontCare.Timestamp)

	return eg
}

func (s *synth) MouseMove(target gui.Point) []gui.EventGroup {
	return []gui.EventGroup{
		s.press(xaxis, target),
		s.press(yaxis, target),
	}
}

func (s *synth) KeyDown(keyid gin.KeyId, at gui.Point) []gui.EventGroup {
	return []gui.EventGroup{
		s.press(keyid, at),
	}
}

func (s *synth) KeyUp(keyid gin.KeyId, at gui.Point) []gui.EventGroup {
	return []gui.EventGroup{
		s.release(keyid, at),
	}
}

func (s *synth) DragGesture(buttonId gin.KeyId, fromPoint, toPoint gui.Point) []gui.EventGroup {
	totalGesture := []gui.EventGroup{}

	// Move to start
	totalGesture = append(totalGesture, s.MouseMove(fromPoint)...)

	// Key down
	totalGesture = append(totalGesture, s.KeyDown(buttonId, fromPoint)...)

	// Move to end
	totalGesture = append(totalGesture, s.MouseMove(toPoint)...)

	// Key up
	totalGesture = append(totalGesture, s.KeyUp(buttonId, toPoint)...)

	return totalGesture
}
