package rendertest_test

import (
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvariantFailureMessages(t *testing.T) {
	t.Run("should use human readable identifiers", func(t *testing.T) {
		var err error
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = e.(error)
				}
			}()
			testbuilder.New().Run(func() {
				// break an invariant
				buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
				buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
			})

			// The above test should fail because it leaks state.
			t.Fatalf("unreachable")
		}()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "ELEMENT_ARRAY_BUFFER")
	})
}
