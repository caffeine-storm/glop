package rendertest

import "github.com/go-gl-legacy/gl"

func WithClearColour(r, g, b, a gl.GLclampf, fn func()) {
	oldClear := [4]float32{0, 0, 0, 0}

	gl.GetFloatv(gl.COLOR_CLEAR_VALUE, oldClear[:])

	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	defer func() {
		gl.ClearColor(
			gl.GLclampf(oldClear[0]),
			gl.GLclampf(oldClear[1]),
			gl.GLclampf(oldClear[2]),
			gl.GLclampf(oldClear[3]))
	}()
	fn()
}
