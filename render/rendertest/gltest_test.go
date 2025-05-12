package rendertest_test

import (
	"fmt"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/stretchr/testify/assert"
)

func TestGlTestHelpers(t *testing.T) {
	t.Run("default builder runs on render thread", func(t *testing.T) {
		testbuilder.New().Run(func() {
			rendertest.AssertOnRenderThread(t)
		})
	})

	t.Run("can run off of render thread", func(t *testing.T) {
		testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
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
			testbuilder.New().WithSize(64, 128).Run(func() {
				rendertest.AssertOnRenderThread(t)
				_, _, dx, dy := debug.GetViewport()
				assert.Equal(dx, uint32(64))
				assert.Equal(dy, uint32(128))
			})
		})

		t.Run("and get the queue after", func(t *testing.T) {
			assert := assert.New(t)
			testbuilder.New().WithSize(64, 128).WithQueue().Run(func(queue render.RenderQueueInterface) {
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
			testbuilder.New().WithQueue().WithSize(64, 128).Run(func(queue render.RenderQueueInterface) {
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
			testbuilder.New().Run(func() {
				// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
				buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
				buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
			})
		})
	})
}

func TestFailureDoesNotCascade(t *testing.T) {
	assert.Panics(t, func() {
		testbuilder.New().Run(func() {
			panic(fmt.Errorf("yup; that's a panic"))
		})
	})
	testbuilder.New().Run(func() {
		// must not panic
	})

	t.Run("even with the deprecated helpers", func(t *testing.T) {
		assert.Panics(t, func() {
			rendertest.DeprecatedWithGl(func() {
				panic(fmt.Errorf("yup; that's a panic"))
			})
		})
		rendertest.DeprecatedWithGl(func() {
			// must not panic
		})
	})

	t.Run("render thread failures fail-fast", func(t *testing.T) {
		assert := assert.New(t)

		shouldGetHere := false
		shouldNotGetHere := false
		shouldAlsoNotGetHere := false

		assert.Panics(func() {
			testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
				shouldGetHere = true

				queue.Queue(func(st render.RenderQueueState) {
					panic(fmt.Errorf("yup; that's a panic"))
					shouldNotGetHere = true
				})
				queue.Purge()

				shouldAlsoNotGetHere = true
			})
		})

		assert.True(shouldGetHere)
		assert.False(shouldNotGetHere)
		assert.False(shouldAlsoNotGetHere)
	})
}
