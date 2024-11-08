package debug_test

import (
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render/rendertest"
)

func TestGlInspect(t *testing.T) {
	t.Run("GetColorMatrix", func(t *testing.T) {
		// TODO(tmckee): clean: add an abstraction for running a func under a
		// 'fresh' GL instance.
		_, render := rendertest.InitGlForTest(50, 50)
		errors := []gl.GLenum{}
		render.Queue(func() {
			// Purge any queued GL errors so that other tests don't affect this one.
			debug.LogAndClearGlErrors(glog.VoidLogger())

			debug.GetColorMatrix()

			for err := gl.GetError(); err != 0; err = gl.GetError() {
				errors = append(errors, err)
				err = gl.GetError()
			}
		})
		render.Purge()

		if len(errors) > 0 {
			// TODO(tmckee): add a helper to strerror error codes from GL.
			t.Fatalf("GetColorMatrix queued %d errors: %v", len(errors), errors)
		}
	})
}
