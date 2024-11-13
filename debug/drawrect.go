package debug

import "github.com/go-gl-legacy/gl"

func BlankAndDrawRectNdc(x1, y1, x2, y2 float64) {
	gl.ClearColor(0, 0, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	DrawRectNdc(x1, y1, x2, y2)
}

func DrawRectNdc(x1, y1, x2, y2 float64) {
	gl.Begin(gl.QUADS)
	gl.Color3f(1, 0, 0)
	gl.Vertex2d(x1, y1)
	gl.Vertex2d(x1, y2)
	gl.Vertex2d(x2, y2)
	gl.Vertex2d(x2, y1)
	gl.End()
}