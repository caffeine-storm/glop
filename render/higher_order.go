package render

import (
	"github.com/MobRulesGames/mathgl"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glew"
)

type MatrixMode int32

const (
	MatrixModeModelView  MatrixMode = gl.MODELVIEW
	MatrixModeProjection MatrixMode = gl.PROJECTION
	MatrixModeTexture    MatrixMode = gl.TEXTURE
	MatrixModeColour     MatrixMode = gl.COLOR
)

func GetCurrentMatrixMode() MatrixMode {
	return MatrixMode(gl.GetInteger(gl.MATRIX_MODE))
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

func WithMultMatrixInMode(mat *Matrix, mode MatrixMode, fn func()) {
	WithMatrixMode(mode, func() {
		gl.MultMatrixf((*[16]float32)(mat))

		fn()
	})
}

func WithFreshMatrices(fn func()) {
	ident := &Matrix{}
	ident.Identity()

	var colourMat *Matrix
	if glew.GL_ARB_imaging {
		colourMat = ident
	}

	WithMatrixInMode(ident, MatrixModeModelView, func() {
		WithMatrixInMode(ident, MatrixModeProjection, func() {
			WithMatrixInMode(ident, MatrixModeTexture, func() {
				if colourMat != nil {
					WithMatrixInMode(ident, MatrixModeColour, func() {
						fn()
					})
				} else {
					fn()
				}
			})
		})
	})
}

func WithTexture2DSetting(enableT2D bool, fn func()) {
	oldval := []bool{false}
	gl.GetBooleanv(gl.TEXTURE_2D, oldval)

	defer func() {
		if oldval[0] {
			gl.Enable(gl.TEXTURE_2D)
		} else {
			gl.Disable(gl.TEXTURE_2D)
		}
	}()

	if enableT2D {
		gl.Enable(gl.TEXTURE_2D)
	} else {
		gl.Disable(gl.TEXTURE_2D)
	}
	fn()
}

func WithTexturing(fn func()) {
	WithTexture2DSetting(true, fn)
}

func WithoutTexturing(fn func()) {
	WithTexture2DSetting(false, fn)
}

func WithColour(r, g, b, a float32, fn func()) {
	assertNormalized(r, g, b, a)

	oldColour := GetCurrentForegroundColour()
	oldr := ByteToNormalizedColour(oldColour.R)
	oldg := ByteToNormalizedColour(oldColour.G)
	oldb := ByteToNormalizedColour(oldColour.B)
	olda := ByteToNormalizedColour(oldColour.A)

	gl.Color4f(r, g, b, a)
	defer func() {
		gl.Color4f(oldr, oldg, oldb, olda)
	}()

	fn()
}

func WithBlankScreen(r, g, b, a gl.GLclampf, fn func()) {
	assertNormalized(r, g, b, a)

	gl.PushAttrib(gl.ACCUM_BUFFER_BIT | gl.CURRENT_BIT)
	defer gl.PopAttrib()

	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	fn()
}
