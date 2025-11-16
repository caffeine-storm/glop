package systemtest_test

import (
	"testing"

	"github.com/caffeine-storm/glop/render/rendertest"
	"github.com/caffeine-storm/glop/render/rendertest/testbuilder"
	"github.com/caffeine-storm/glop/system/systemtest"
)

func TestSystemtestBuilder(t *testing.T) {
	t.Run("can wrap a rendertest builder", func(t *testing.T) {
		systemtest.TestBuilder(testbuilder.New()).Run(func(systemtest.Window) {
			rendertest.AssertOffRenderThread(t)
		})
	})
}
