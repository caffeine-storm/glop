package rendertest_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/gloptest"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/stretchr/testify/assert"
)

func clearUnsetValues(mp map[string]int) {
	todelete := map[string]bool{}
	for key, val := range mp {
		if val == 0 {
			todelete[key] = true
		}
	}

	for key := range todelete {
		delete(mp, key)
	}
}

func TestCrossTalkPrevention(t *testing.T) {
	// Before and after a test, there should be certain invariants otherwise
	// tests are susceptible to cross-talk.
	var initialState map[string]int
	testbuilder.New().Run(func() {
		initialState = debug.GetBindingsSet()
	})

	t.Run("initial state", func(t *testing.T) {
		for bindingName, boundVal := range initialState {
			if boundVal != 0 {
				t.Logf("found initially bound state: name: %q, val: %d\n", bindingName, boundVal)
				t.Fail()
			}
		}
	})
	clearUnsetValues(initialState)
	assert.Empty(t, initialState, "tests should be allowed to expect a clean initial state")

	var taintedState map[string]int
	t.Run("testbuilder helpers send errors if state-change is leaked", func(t *testing.T) {
		var expectedError error = nil
		assert.Panics(t, func() {
			testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
				queue.AddErrorCallback(func(_ render.RenderQueueInterface, e error) {
					expectedError = e
				})
				queue.Queue(func(render.RenderQueueState) {
					// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
					buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
					buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
					taintedState = debug.GetBindingsSet()
				})
				queue.Purge()

				// We haven't run the "cleanup" phase of this testcase so state leakage
				// should not be checked yet.
				assert.Nil(t, expectedError)
			})
		})
		assert.NotNil(t, expectedError)

		clearUnsetValues(taintedState)
		assert.NotEqual(t, slices.Sorted(maps.Keys(initialState)), slices.Sorted(maps.Keys(taintedState)))
	})

	var nextState map[string]int

	// TODO(#37): won't need this test once deprecated things are removed.
	t.Run("the deprecated helpers merely warn", func(t *testing.T) {
		output := gloptest.CollectOutput(func() {
			rendertest.DeprecatedWithGl(func() {
				// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
				buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
				buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
				nextState = debug.GetBindingsSet()
			})
		})

		clearUnsetValues(nextState)
		assert.NotEqual(t, slices.Sorted(maps.Keys(initialState)), slices.Sorted(maps.Keys(nextState)))

		assert.Contains(t, strings.Join(output, "\n"), "state leakage")
	})
}
