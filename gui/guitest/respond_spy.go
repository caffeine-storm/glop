package guitest

import "github.com/runningwild/glop/gui"

type RespondSpy interface {
	Respond(gui *gui.Gui, evt gui.EventGroup)
	GetEvents() []gui.EventGroup
}

type respondSpy struct {
	events []gui.EventGroup
}

func (spy *respondSpy) Respond(gui *gui.Gui, evt gui.EventGroup) {
	spy.events = append(spy.events, evt)
}

func (spy *respondSpy) GetEvents() []gui.EventGroup {
	return spy.events
}

func NewRespondSpy() RespondSpy {
	return &respondSpy{}
}
