package rendertest

import (
	"fmt"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/system"
)

// TODO(#37): prefer testbuilder.New()
func DeprecatedWithGlAndHandleForTest(width, height int, fn func(system.System, system.NativeWindowHandle, render.RenderQueueInterface)) {
	RunDeprecatedTestWithCachedContext(width, height, fn)
}

// TODO(#37): prefer testbuilder.New()
func DeprecatedWithGlForTest(width, height int, fn func(system.System, render.RenderQueueInterface)) {
	DeprecatedWithGlAndHandleForTest(width, height, func(sys system.System, _ system.NativeWindowHandle, queue render.RenderQueueInterface) {
		fn(sys, queue)
	})
}

// TODO(#37): prefer testbuilder.New()
func DeprecatedWithGl(fn func()) {
	DeprecatedWithGlForTest(50, 50, func(sys system.System, renderQueue render.RenderQueueInterface) {
		logger := glog.ErrorLogger()

		errors := []gl.GLenum{}
		renderQueue.Queue(func(render.RenderQueueState) {
			// Clear out GL's error queue so that a leaky test doesn't break us by
			// accident.
			for err := gl.GetError(); err != 0; err = gl.GetError() {
				logger.Error("glErrors before given func", "error_code", err)
				err = gl.GetError()
			}

			fn()

			for err := gl.GetError(); err != 0; err = gl.GetError() {
				errors = append(errors, err)
				err = gl.GetError()
			}
		})
		renderQueue.Purge()

		// If there were GL errors _caused_ by the given func, fail!
		if len(errors) > 0 {
			// TODO(tmckee): add a helper to strerror error codes from GL.
			panic(fmt.Errorf("WithGl ran a func that produced %d GL errors: %v", len(errors), errors))
		}
	})
}
