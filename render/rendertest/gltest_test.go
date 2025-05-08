package rendertest_test

import (
	"testing"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
)

func TestGlTestHelpers(t *testing.T) {
	t.Run("default builder runs on render thread", func(t *testing.T) {
		rendertest.GlTest().Run(func() {
			render.MustBeOnRenderThread()
		})
	})
}
