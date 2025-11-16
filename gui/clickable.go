package gui

import "github.com/caffeine-storm/glop/gin"

// Embed a Clickable object to run a specified function when the widget
// is clicked and run a specified function.
type Clickable struct {
	on_click func(EventHandlingContext, int64)
}

func (c Clickable) DoRespond(ctx EventHandlingContext, event_group EventGroup) (bool, bool) {
	if event_group.IsPressed(gin.AnyMouseLButton) {
		c.on_click(ctx, event_group.TimestampMs)
		return true, false
	}
	return false, false
}
