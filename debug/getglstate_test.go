package debug_test

import (
	"fmt"
	"testing"

	"github.com/caffeine-storm/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/stretchr/testify/assert"
)

func TestGlInspect(t *testing.T) {
	t.Run("GetColorMatrix", func(t *testing.T) {
		errors := []gl.GLenum{}
		testbuilder.Run(func() {
			debug.GetColorMatrix()

			for err := gl.GetError(); err != 0; err = gl.GetError() {
				errors = append(errors, err)
			}
		})

		if len(errors) > 0 {
			// TODO(tmckee): add a helper to strerror error codes from GL.
			t.Fatalf("GetColorMatrix queued %d errors: %v", len(errors), errors)
		}
	})

	t.Run("GetGlState exposes helpful data", func(t *testing.T) {
		var glstate *debug.GlState
		testbuilder.Run(func() {
			glstate = debug.GetGlState()
		})

		stringified := fmt.Sprintf("%v", glstate)

		t.Run("active texture unit", func(t *testing.T) {
			assert.Contains(t, stringified, "ACTIVE_TEXTURE")
		})

		t.Run("bindings", func(t *testing.T) {
			assert.Contains(t, stringified, "ARRAY_BUFFER_BINDING")
			assert.Contains(t, stringified, "ELEMENT_ARRAY_BUFFER_BINDING")
			assert.Contains(t, stringified, "PIXEL_PACK_BUFFER_BINDING")
			assert.Contains(t, stringified, "PIXEL_UNPACK_BUFFER_BINDING")
			assert.Contains(t, stringified, "TEXTURE_BINDING_2D")
			assert.Contains(t, stringified, "TEXTURE_COORD_ARRAY_BUFFER_BINDING")
			assert.Contains(t, stringified, "VERTEX_ARRAY_BUFFER_BINDING")
		})

		t.Run("colours", func(t *testing.T) {
			assert.Contains(t, stringified, "gl.CURRENT_COLOR")
			assert.Contains(t, stringified, "gl.COLOR_CLEAR_VALUE")
		})
	})
}
