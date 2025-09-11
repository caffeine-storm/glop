package guitest

import "github.com/runningwild/glop/gui"

type Responder interface {
	Respond(gui *gui.Gui, evt gui.EventGroup) bool
}

type RespondSpy struct {
	events []gui.EventGroup
}

func (spy *RespondSpy) Respond(gui *gui.Gui, evt gui.EventGroup) bool {
	spy.events = append(spy.events, evt)
	return false
}

func (spy *RespondSpy) GetEvents() []gui.EventGroup {
	return spy.events
}

func NewRespondSpy() *RespondSpy {
	return &RespondSpy{}
}
