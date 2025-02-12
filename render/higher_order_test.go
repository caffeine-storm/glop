package render_test

import (
	"fmt"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func pickADifferentMode(someMatrixMode render.MatrixMode) render.MatrixMode {
	switch someMatrixMode {
	case render.MatrixModeModelView:
		return render.MatrixModeProjection
	case render.MatrixModeProjection:
		fallthrough
	case render.MatrixModeTexture:
		fallthrough
	case render.MatrixModeColour:
		return render.MatrixModeModelView
	default:
		panic(fmt.Errorf("bad matrixmode: %d", someMatrixMode))
	}
}

func pickADifferentMatrix(someMatrix render.Matrix) render.Matrix {
	notIdentity := render.Matrix{
		0, 1, 0, 0,
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	someMatrix.Multiply(&notIdentity)
	return someMatrix
}

func TestWithMatrixMode(t *testing.T) {
	var beforeMode, duringMode, targetMode, afterMode render.MatrixMode
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

func TestWithMatrixInMode(t *testing.T) {
	var beforeMode, duringMode, targetMode, afterMode render.MatrixMode
	var beforeMat, duringMat, targetMat, afterMat render.Matrix

	rendertest.WithGl(func() {
		beforeMode = render.GetCurrentMatrixMode()
		targetMode = pickADifferentMode(beforeMode)
		if beforeMode == targetMode {
			panic(fmt.Errorf("bad test; need to find a _different_ mode"))
		}

		beforeMat = render.GetCurrentMatrix(beforeMode)
		targetMat = pickADifferentMatrix(beforeMat)
		if render.MatrixIsEqual(beforeMat, targetMat) {
			panic(fmt.Errorf("bad test; need to find a _different_ matrix"))
		}

		render.WithMatrixInMode(&targetMat, targetMode, func() {
			duringMode = render.GetCurrentMatrixMode()
			duringMat = render.GetCurrentMatrix(duringMode)

			// clobbering the matrix mode shouldn't break anything
			gl.MatrixMode(gl.GLenum(pickADifferentMode(duringMode)))
		})

		afterMode = render.GetCurrentMatrixMode()
		afterMat = render.GetCurrentMatrix(afterMode)
	})

	assert.Equal(t, duringMode, targetMode)
	assert.Equal(t, afterMode, beforeMode)

	assert.True(t, render.MatrixIsEqual(duringMat, targetMat))
	assert.True(t, render.MatrixIsEqual(afterMat, beforeMat))
}
