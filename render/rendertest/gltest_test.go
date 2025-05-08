package rendertest_test

import (
	"testing"

	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func assertOnRenderThread(*testing.T) {
	render.MustBeOnRenderThread()
}

func assertOffRenderThread(t *testing.T) {
	assert.Panics(t, func() {
		render.MustBeOnRenderThread()
	})
}

func TestGlTestHelpers(t *testing.T) {
	t.Run("default builder runs on render thread", func(t *testing.T) {
		rendertest.GlTest().Run(func() {
			assertOnRenderThread(t)
		})
	})

	t.Run("can run off of render thread", func(t *testing.T) {
		rendertest.GlTest().WithQueue().Run(func(queue render.RenderQueueInterface) {
			assertOffRenderThread(t)

			queue.Queue(func(render.RenderQueueState) {
				assertOnRenderThread(t)
			})
			queue.Purge()
		})
	})

	t.Run("can specify dimensions", func(t *testing.T) {
		t.Run("with literal sizes", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.GlTest().WithSize(64, 128).Run(func() {
				assertOnRenderThread(t)
				_, _, dx, dy := debug.GetViewport()
				assert.Equal(dx, uint32(64))
				assert.Equal(dy, uint32(128))
			})
		})

		t.Run("and get the queue after", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.GlTest().WithSize(64, 128).WithQueue().Run(func(queue render.RenderQueueInterface) {
				assertOffRenderThread(t)
				queue.Queue(func(render.RenderQueueState) {
					_, _, dx, dy := debug.GetViewport()
					assert.Equal(dx, uint32(64))
					assert.Equal(dy, uint32(128))
				})
				queue.Purge()
			})
		})

		t.Run("and get the queue first", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.GlTest().WithQueue().WithSize(64, 128).Run(func(queue render.RenderQueueInterface) {
				assertOffRenderThread(t)
				queue.Queue(func(render.RenderQueueState) {
					_, _, dx, dy := debug.GetViewport()
					assert.Equal(dx, uint32(64))
					assert.Equal(dy, uint32(128))
				})
				queue.Purge()
			})
		})
	})
}
