package rendertest_test

import (
	"log"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithGl(t *testing.T) {
	rendertest.GlTest().Run(func() {
		versionString := gl.GetString(gl.VERSION)
		log.Printf("versionString: %q\n", versionString)

		if versionString == "" {
			t.Error("gl.GetString(gl.VERSION) must not return the empty string once OpenGL is initialized")
		}
	})
}

func TestCrossTalkPrevention(t *testing.T) {
	// Before and after a test, there should be certain invariants otherwise
	// tests are susceptible to cross-talk.
	var initialState map[string]int
	rendertest.GlTest().Run(func() {
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

	var taintedState map[string]int
	t.Run("tainted state should fail", func(t *testing.T) {
		assert.Panics(t, func() {
			rendertest.GlTest().Run(func() {
				// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
				buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
				buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
				taintedState = debug.GetBindingsSet()
			})
		})
	})
	require.NotNil(t, taintedState)
	assert.NotEqual(t, initialState, taintedState)

	var nextState map[string]int
	t.Run("the deprecated helpers merely warn", func(t *testing.T) {
		rendertest.DeprecatedWithGl(func() {
			// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
			buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
			buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
			taintedState = debug.GetBindingsSet()
		})

		assert.Equal(t, taintedState, nextState)
	})
}
