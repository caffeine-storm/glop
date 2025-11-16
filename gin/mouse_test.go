package gin_test

import (
	"testing"

	"github.com/caffeine-storm/glop/gin"
	"github.com/caffeine-storm/glop/glog"
	"github.com/stretchr/testify/assert"
)

var verticalWheelKeyId = gin.KeyId{
	Index: gin.MouseWheelVertical,
	Device: gin.DeviceId{
		Index: 0,
		Type:  gin.DeviceTypeMouse,
	},
}

func scrollUpSequence() (int64, []gin.OsEvent) {
	events := []gin.OsEvent{
		{
			KeyId:       verticalWheelKeyId,
			Press_amt:   17,
			TimestampMs: 13,
			X:           5,
			Y:           7,
		},
	}
	return 42, events
}

func TestMouseInput(t *testing.T) {
	t.Run("mouse wheel", func(t *testing.T) {
		t.Run("vertical wheel", func(t *testing.T) {
			assert := assert.New(t)

			logger := glog.New(&glog.Opts{
				Level: glog.LevelTrace,
			})
			inputObj := gin.MakeLogged(logger)
			wheelUp := inputObj.GetKeyById(verticalWheelKeyId)

			assert.Equal(wheelUp.FramePressAmt(), float64(0), "a fresh gin.Key should have no press amount")

			inputObj.Think(scrollUpSequence())

			assert.Greater(wheelUp.FramePressAmt(), float64(0), "after a scroll up, the 'MouseWheelVertical' key should have a positive press amount")
		})
	})
}
