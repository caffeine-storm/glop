package rendertest

import "github.com/go-gl-legacy/gl"

func ClearScreen() {
	// Set the default clear values.
	gl.ClearColor(0, 0, 0, 1)
	gl.ClearDepth(1)
	gl.ClearStencil(0)

	// Clear all the bitplanes we care about.
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
}
