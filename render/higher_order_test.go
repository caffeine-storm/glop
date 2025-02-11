package render_test

import (
	"fmt"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func getCurrentMatrixMode() gl.GLenum {
	var matmode [1]int32
	gl.GetIntegerv(gl.MATRIX_MODE, matmode[:])
	return gl.GLenum(matmode[0])
}

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
		firstMode = getCurrentMatrixMode()
		secondMode := pickADifferentMode(firstMode)
		render.WithMatrixMode(secondMode, func() {
			actual = getCurrentMatrixMode()
		})
	})

	assert.NotEqual(t, actual, firstMode)
}
