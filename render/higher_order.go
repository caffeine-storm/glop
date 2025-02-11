package render

import "github.com/go-gl-legacy/gl"

func GetCurrentMatrixMode() gl.GLenum {
	var matmode [1]int32
	gl.GetIntegerv(gl.MATRIX_MODE, matmode[:])
	return gl.GLenum(matmode[0])
}

// TODO(tmckee): can we use the type system to prevent bad enums?
func WithMatrixMode(mode gl.GLenum, fn func()) {
	oldMode := GetCurrentMatrixMode()
	gl.MatrixMode(mode)
	defer gl.MatrixMode(oldMode)

	fn()
}
