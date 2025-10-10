package rendertest

import (
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
)

func BlankAndDrawRectNdc(x1, y1, x2, y2 float64) {
	render.WithFreshMatrices(func() {
		BlankAndDrawRect(x1, y1, x2, y2)
	})
}

func BlankAndDrawRect(x1, y1, x2, y2 float64) {
	render.WithBlankScreen(0, 0, 0, 1, func() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		DrawRect(x1, y1, x2, y2)
	})
}

func DrawRectNdc(x1, y1, x2, y2 float64) {
	render.WithFreshMatrices(func() {
		DrawRect(x1, y1, x2, y2)
	})
}

func DrawRect(x1, y1, x2, y2 float64) {
	render.WithoutTexturing(func() {
		render.WithColour(1, 0, 0, 1, func() {
			gl.Begin(gl.TRIANGLES)
			gl.Vertex2d(x1, y1)
			gl.Vertex2d(x1, y2)
			gl.Vertex2d(x2, y2)

			gl.Vertex2d(x1, y1)
			gl.Vertex2d(x2, y2)
			gl.Vertex2d(x2, y1)
			gl.End()
		})
	})
}
