package rendertest_test

import (
	"testing"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	"github.com/stretchr/testify/assert"
)

func TestGlContext(t *testing.T) {
	t.Run("uses a stubbed clock in tests", func(t *testing.T) {
		assert := assert.New(t)
		rendertest.RunTestWithCachedContext(64, 64, func(sys system.System, hdl system.NativeWindowHandle, queue render.RenderQueueInterface) {
			t1 := sys.Think()
			assert.NotEqual(0, t1, "'time' should not start at 0; it would break a bunch of assumptions in Haunts T_T")

			t2 := sys.Think()
			assert.Equal(t1, t2, "time should only advance if we tell it to")

			fiveMinutesInMilliseconds := int64(5 * 60 * 1000)
			rendertest.AdvanceTime(sys, fiveMinutesInMilliseconds)

			t3 := sys.Think()
			assert.Equal(t2+fiveMinutesInMilliseconds, t3, "'time' should have advanced by the given step")
		})
	})
}
