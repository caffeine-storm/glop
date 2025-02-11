package render

import "github.com/go-gl-legacy/gl"

// TODO(tmckee): can we use the type system to prevent bad enums?
func WithMatrixMode(mode gl.GLenum, fn func()) {
	fn()
}
