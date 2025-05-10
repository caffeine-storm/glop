package rendertest_test

import (
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func TestGlTestHelpers(t *testing.T) {
	t.Run("default builder runs on render thread", func(t *testing.T) {
		rendertest.GlTest().Run(func() {
			rendertest.AssertOnRenderThread(t)
		})
	})

	t.Run("can run off of render thread", func(t *testing.T) {
		rendertest.GlTest().WithQueue().Run(func(queue render.RenderQueueInterface) {
			rendertest.AssertOffRenderThread(t)

			queue.Queue(func(render.RenderQueueState) {
				rendertest.AssertOnRenderThread(t)
			})
			queue.Purge()
		})
	})

	t.Run("can specify dimensions", func(t *testing.T) {
		t.Run("with literal sizes", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.GlTest().WithSize(64, 128).Run(func() {
				rendertest.AssertOnRenderThread(t)
				_, _, dx, dy := debug.GetViewport()
				assert.Equal(dx, uint32(64))
				assert.Equal(dy, uint32(128))
			})
		})

		t.Run("and get the queue after", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.GlTest().WithSize(64, 128).WithQueue().Run(func(queue render.RenderQueueInterface) {
				rendertest.AssertOffRenderThread(t)
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
				rendertest.AssertOffRenderThread(t)
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

func TestGlStateLeakage(t *testing.T) {
	t.Run("GlTest should complain upon leakage", func(t *testing.T) {
		assert.Panics(t, func() {
			rendertest.GlTest().Run(func() {
				// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
				buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
				buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
			})
		})
	})
}
