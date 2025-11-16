package gintest

import (
	"github.com/caffeine-storm/glop/gin"
	"github.com/caffeine-storm/glop/gin/aggregator"
)

func ClickEventGroup(input *gin.Input) gin.EventGroup {
	leftMouseButtonKeyId := gin.AnyMouseLButton

	leftButtonEvent := gin.Event{
		Key:  input.GetKeyById(leftMouseButtonKeyId),
		Type: aggregator.Press,
	}
	return gin.EventGroup{
		Events:      []gin.Event{leftButtonEvent},
		TimestampMs: 17,
	}
}
