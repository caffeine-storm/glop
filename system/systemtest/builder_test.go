package systemtest_test

import (
	"testing"

	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system/systemtest"
)

func TestSystemtestBuilder(t *testing.T) {
	t.Run("can wrap a rendertest builder", func(t *testing.T) {
		systemtest.TestBuilder(rendertest.GlTest()).Run(func(systemtest.Window) {
			rendertest.AssertOnRenderThread(t)
		})
	})
}
