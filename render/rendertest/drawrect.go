package rendertest

import (
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
)

func BlankAndDrawRectNdc(x1, y1, x2, y2 float64) {
	gl.ClearColor(0.5, 0.5, 0.5, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	DrawRectNdc(x1, y1, x2, y2)
}

func DrawRectNdc(x1, y1, x2, y2 float64) {
	render.WithFreshMatrices(func() {
		gl.Begin(gl.TRIANGLES)
		gl.Color3f(1, 0, 0)
		gl.Vertex2d(x1, y1)
		gl.Vertex2d(x1, y2)
		gl.Vertex2d(x2, y2)

		gl.Vertex2d(x1, y1)
		gl.Vertex2d(x2, y2)
		gl.Vertex2d(x2, y1)
		gl.End()
	})
}
