package guitest_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui/guitest"
	"github.com/stretchr/testify/assert"
)

func TestSynthesize(t *testing.T) {
	t.Run("WheelDown", func(t *testing.T) {
		assert := assert.New(t)
		synthesized := guitest.SynthesizeEvents().WheelDown(-42)

		mouseWheelKeyId := gin.AnyMouseWheelVertical
		mouseWheelKeyId.Device.Index = 0

		assert.True(synthesized.IsPressed(mouseWheelKeyId))
		assert.Equal(float64(-42), synthesized.PrimaryEvent().Key.FramePressTotal())
	})
}
