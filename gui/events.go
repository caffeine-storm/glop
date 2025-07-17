package gui

import "github.com/runningwild/glop/gin"

type EventGroup struct {
	gin.EventGroup
	DispatchedToFocussedWidget bool
}

type EventHandlingContext interface {
	UseMousePosition(grp EventGroup) (Point, bool)
	LeftButton(grp EventGroup) bool
	MiddleButton(grp EventGroup) bool
	RightButton(grp EventGroup) bool
}
