package rendertest_test

import (
	"testing"

	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func TestGlTestHelpers(t *testing.T) {
	t.Run("default builder runs on render thread", func(t *testing.T) {
		rendertest.GlTest().Run(func() {
			render.MustBeOnRenderThread()
		})
	})

	t.Run("can run off of render thread", func(t *testing.T) {
		assert := assert.New(t)
		rendertest.GlTest().WithQueue().Run(func(queue render.RenderQueueInterface) {
			assert.Panics(func() {
				render.MustBeOnRenderThread()
			})

			queue.Queue(func(render.RenderQueueState) {
				render.MustBeOnRenderThread()
			})
			queue.Purge()
		})
	})

	t.Run("can specify dimensions", func(t *testing.T) {
		t.Run("with literal sizes", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.GlTest().WithSize(64, 128).Run(func() {
				render.MustBeOnRenderThread()
				_, _, dx, dy := debug.GetViewport()
				assert.Equal(dx, uint32(64))
				assert.Equal(dy, uint32(128))
			})
		})
	})
}
