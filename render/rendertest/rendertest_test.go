package rendertest_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/stretchr/testify/assert"
)

func TestFailureDoesNotCascade(t *testing.T) {
	assert.Panics(t, func() {
		testbuilder.New().Run(func() {
			panic(fmt.Errorf("yup; that's a panic"))
		})
	})
	testbuilder.New().Run(func() {
		// must not panic
	})

	// TODO(#37): won't need this test once deprecated things are removed.
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
		shouldAlsoNotGetHere := false

		assert.Panics(func() {
			testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
				shouldGetHere = true

				queue.Queue(func(st render.RenderQueueState) {
					panic(fmt.Errorf("yup; that's a panic"))
				})
				queue.Purge()

				shouldAlsoNotGetHere = true
			})
		})

		assert.True(shouldGetHere)
		assert.False(shouldAlsoNotGetHere)
	})
}
