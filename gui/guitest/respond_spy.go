package guitest

import "github.com/runningwild/glop/gui"

type Responder interface {
	Respond(gui *gui.Gui, evt gui.EventGroup)
}

type RespondSpy struct {
	events []gui.EventGroup
}

func (spy *RespondSpy) Respond(gui *gui.Gui, evt gui.EventGroup) {
	spy.events = append(spy.events, evt)
}

func (spy *RespondSpy) GetEvents() []gui.EventGroup {
	return spy.events
}

func NewRespondSpy() *RespondSpy {
	return &RespondSpy{}
}
