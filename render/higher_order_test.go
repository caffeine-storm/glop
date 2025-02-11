package render_test

import (
	"fmt"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func pickADifferentMode(someMatrixMode gl.GLenum) gl.GLenum {
	switch someMatrixMode {
	case gl.MODELVIEW:
		return gl.PROJECTION
	case gl.PROJECTION:
		fallthrough
	case gl.TEXTURE:
		fallthrough
	case gl.COLOR:
		return gl.MODELVIEW
	default:
		panic(fmt.Errorf("bad matrixmode: %d", someMatrixMode))
	}
}

func TestWithMatrixMode(t *testing.T) {
	var beforeMode, duringMode, targetMode, afterMode gl.GLenum
	rendertest.WithGl(func() {
		beforeMode = render.GetCurrentMatrixMode()
		targetMode = pickADifferentMode(beforeMode)

		if beforeMode == targetMode {
			panic(fmt.Errorf("bad test; need to find a _different_ mode"))
		}

		render.WithMatrixMode(targetMode, func() {
			duringMode = render.GetCurrentMatrixMode()
		})

		afterMode = render.GetCurrentMatrixMode()
	})

	assert.Equal(t, duringMode, targetMode)
	assert.Equal(t, afterMode, beforeMode)
}
