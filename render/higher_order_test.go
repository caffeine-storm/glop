package render_test

import (
	"fmt"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glew"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
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

func assertFreshState(t *testing.T, st *debug.GlState) {
	ident64 := make([]float64, 16)
	for i := range 4 {
		ident64[4*i+i] = 1.0
	}

	for i, stateComponent := range [][]float64{
		st.ModelViewMatrix,
		st.ProjectionMatrix,
		st.TextureMatrix,
	} {
		if !slices.Equal(ident64, stateComponent) {
			t.Logf("mismatch: component#%d, %v vs %v", i, ident64, stateComponent)
			t.Fail()
		}
	}

	expectedColorMat := ident64
	if !glew.GL_ARB_imaging {
		expectedColorMat = nil
	}

	if !slices.Equal(expectedColorMat, st.ColorMatrix) {
		t.Logf("mismatch: colormatrix, %v vs %v", ident64, st.ColorMatrix)
		t.Fail()
	}
}

func TestWithMatrixMode(t *testing.T) {
	var beforeMode, duringMode, targetMode, afterMode render.MatrixMode
	rendertest.DeprecatedWithGl(func() {
		beforeMode = render.GetCurrentMatrixMode()
		targetMode = pickADifferentMode(beforeMode)

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

	rendertest.DeprecatedWithGl(func() {
		beforeMode = render.GetCurrentMatrixMode()
		targetMode = pickADifferentMode(beforeMode)

		beforeMat = render.GetCurrentMatrix(beforeMode)
		targetMat = pickADifferentMatrix(beforeMat)

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

func TestWithFreshMatrices(t *testing.T) {
	ident := render.Matrix{}
	ident.Identity()
	notIdent := pickADifferentMatrix(ident)

	var beforeState, entryState, exitState, afterState *debug.GlState

	rendertest.DeprecatedWithGl(func() {
		// Start out with matrics that are 'not-fresh'.
		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadMatrixf((*[16]float32)(&notIdent))

		beforeState = debug.GetGlState()

		render.WithFreshMatrices(func() {
			entryState = debug.GetGlState()

			// clobbering the current matrix must not persist
			curMode := render.GetCurrentMatrixMode()
			notCurMode := pickADifferentMode(curMode)
			curMat := render.GetCurrentMatrix(curMode)
			notCurMat := pickADifferentMatrix(curMat)
			gl.LoadMatrixf((*[16]float32)(&notCurMat))

			// clobbering the matrix mode must not persist
			gl.MatrixMode(gl.GLenum(notCurMode))

			exitState = debug.GetGlState()
			assert.NotEqual(t, entryState, exitState)
		})

		afterState = debug.GetGlState()
	})

	assert.Equal(t, beforeState, afterState)
	assert.NotEqual(t, entryState, beforeState)

	assertFreshState(t, entryState)
}

func TestTexture2DHelpers(t *testing.T) {
	rendertest.DeprecatedWithGl(func() {
		testcase := func() {
			flags := []bool{false, false, false}
			gl.GetBooleanv(gl.TEXTURE_2D, flags[0:])

			render.WithTexture2DSetting(!flags[0], func() {
				gl.GetBooleanv(gl.TEXTURE_2D, flags[1:])
			})

			gl.GetBooleanv(gl.TEXTURE_2D, flags[2:])

			if flags[0] != flags[2] {
				t.Fatalf("mismatch texture2d state before: %v, in: %v, after: %v", flags[0], flags[1], flags[2])
			}

			if flags[0] == flags[1] {
				t.Fatalf("didn't toggle texture2d state before: %v, in: %v, after: %v", flags[0], flags[1], flags[2])
			}
		}
		gl.Disable(gl.TEXTURE_2D)
		testcase()
		gl.Enable(gl.TEXTURE_2D)
		testcase()
	})
}
