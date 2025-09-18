package gui

type Responder interface {
	// Returns true if this or any of its children consumed the event group.
	Respond(*Gui, EventGroup) bool
}
