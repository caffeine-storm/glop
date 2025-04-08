package gintesting

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gin/aggregator"
)

func ClickEventGroup(input *gin.Input) gin.EventGroup {
	leftMouseButtonKeyId := gin.AnyMouseLButton

	leftButtonEvent := gin.Event{
		Key:  input.GetKeyById(leftMouseButtonKeyId),
		Type: aggregator.Press,
	}
	return gin.EventGroup{
		Events:    []gin.Event{leftButtonEvent},
		Timestamp: 17,
	}
}
