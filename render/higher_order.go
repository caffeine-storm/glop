package render

import (
	"github.com/MobRulesGames/mathgl"
	"github.com/go-gl-legacy/gl"
)

type MatrixMode int32

const (
	MatrixModeModelView  MatrixMode = gl.MODELVIEW
	MatrixModeProjection MatrixMode = gl.PROJECTION
	MatrixModeTexture    MatrixMode = gl.TEXTURE
	MatrixModeColour     MatrixMode = gl.COLOR
)

func GetCurrentMatrixMode() MatrixMode {
	var matmode [1]int32
	gl.GetIntegerv(gl.MATRIX_MODE, matmode[:])
	return MatrixMode(matmode[0])
}

type Matrix = mathgl.Mat4

func GetCurrentMatrix(mode MatrixMode) Matrix {
	var getKey gl.GLenum
	switch mode {
	case MatrixModeModelView:
		getKey = gl.MODELVIEW_MATRIX
	case MatrixModeProjection:
		getKey = gl.PROJECTION_MATRIX
	case MatrixModeTexture:
		getKey = gl.TEXTURE_MATRIX
	case MatrixModeColour:
		getKey = gl.COLOR_MATRIX
	}

	var mat mathgl.Mat4
	gl.GetFloatv(getKey, mat[:])
	return mat
}

func MatrixIsEqual(lhs, rhs Matrix) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i := range lhs {
		if lhs[i] != rhs[i] {
			return false
		}
	}

	return true
}

func WithMatrixMode(mode MatrixMode, fn func()) {
	oldMode := GetCurrentMatrixMode()
	gl.MatrixMode(gl.GLenum(mode))
	gl.PushMatrix()
	defer func() {
		gl.MatrixMode(gl.GLenum(mode))
		gl.PopMatrix()

		gl.MatrixMode(gl.GLenum(oldMode))
	}()

	fn()
}

func WithMatrixInMode(mat *Matrix, mode MatrixMode, fn func()) {
	WithMatrixMode(mode, func() {
		gl.LoadMatrixf((*[16]float32)(mat))

		fn()
	})
}
