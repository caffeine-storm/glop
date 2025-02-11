package render

import "github.com/go-gl-legacy/gl"

func GetCurrentMatrixMode() MatrixMode {
	var matmode [1]int32
	gl.GetIntegerv(gl.MATRIX_MODE, matmode[:])
	return MatrixMode(matmode[0])
}

type MatrixMode int32

const (
	MatrixModeModelView  MatrixMode = gl.MODELVIEW
	MatrixModeProjection MatrixMode = gl.PROJECTION
	MatrixModeTexture    MatrixMode = gl.TEXTURE
	MatrixModeColour     MatrixMode = gl.COLOR
)

func WithMatrixMode(mode MatrixMode, fn func()) {
	oldMode := GetCurrentMatrixMode()
	gl.MatrixMode(gl.GLenum(mode))
	defer gl.MatrixMode(gl.GLenum(oldMode))

	fn()
}
