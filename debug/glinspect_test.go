package debug_test

import (
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render/rendertest"
)

func TestGlInspect(t *testing.T) {
	t.Run("GetColorMatrix", func(t *testing.T) {
		errors := []gl.GLenum{}
		rendertest.WithGl(func() {
			debug.GetColorMatrix()

			for err := gl.GetError(); err != 0; err = gl.GetError() {
				errors = append(errors, err)
				err = gl.GetError()
			}
		})

		if len(errors) > 0 {
			// TODO(tmckee): add a helper to strerror error codes from GL.
			t.Fatalf("GetColorMatrix queued %d errors: %v", len(errors), errors)
		}
	})
}
