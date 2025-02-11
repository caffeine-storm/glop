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
	var firstMode, actual gl.GLenum
	rendertest.WithGl(func() {
		firstMode = render.GetCurrentMatrixMode()
		secondMode := pickADifferentMode(firstMode)
		render.WithMatrixMode(secondMode, func() {
			actual = render.GetCurrentMatrixMode()
		})
	})

	assert.NotEqual(t, actual, firstMode)
}
