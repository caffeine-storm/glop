package gui

import (
	"image/color"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
)

type TextLine struct {
	EmbeddedWidget
	Childless
	StubDoResponder
	StubDrawFocuseder
	BasicZone
	text string
	// TODO(tmckee:#24): this isn't written to; it'll always be the empty string
	// so TextEditLine is broken and can't see when the text changes.
	next_text string
	font_id   string
	initted   bool
	rdims     Dims
	color     color.Color
	scale     float64
}

func (w *TextLine) String() string {
	return "text line"
}

// TODO(tmckee): we should take a font by reference instead of by
// stringified-name. That way, the compiler can check for us that the font is
// loaded.
// TODO(tmckee): we shouldn't have to pass a width at construction, just need
// one during draw.
func MakeTextLine(fontId, text string, width int, r, g, b, a float64) *TextLine {
	var w TextLine

	w.font_id = fontId
	w.text = text
	w.EmbeddedWidget = &BasicWidget{CoreWidget: &w}
	w.SetColor(r, g, b, a)
	// TODO(tmckee): Request_dims isn't used; should it be? It's supposed to let
	// us pick a size at construction time but do we need/use that?
	// It's used as 'natural dimensions' in other widgets.
	w.Request_dims = Dims{Dx: width, Dy: 25}
	return &w
}

func (w *TextLine) SetColor(r, g, b, a float64) {
	w.color = color.NRGBA{
		R: uint8(255 * r),
		G: uint8(255 * g),
		B: uint8(255 * b),
		A: uint8(255 * a),
	}
}

func (w *TextLine) GetText() string {
	return w.text
}

func (w *TextLine) SetText(str string) {
	w.text = str
}

func (w *TextLine) DoThink(int64, bool) {
}

func (w *TextLine) preDraw(region Region, _ DrawingContext) {
	// Draw a black rectangle over the region to erase what might be there
	// already.
	gl.Color3d(0, 0, 0)
	gl.Begin(gl.QUADS)
	gl.Vertex2i(region.X, region.Y)
	gl.Vertex2i(region.X, region.Y+region.Dy)
	gl.Vertex2i(region.X+region.Dx, region.Y+region.Dy)
	gl.Vertex2i(region.X+region.Dx, region.Y)
	gl.End()
}

func (w *TextLine) postDraw(region Region, ctx DrawingContext) {
}

func (w *TextLine) Draw(region Region, ctx DrawingContext) {
	region.PushClipPlanes()
	defer region.PopClipPlanes()
	w.preDraw(region, ctx)
	w.coreDraw(region, ctx)
	w.postDraw(region, ctx)
}

func (w *TextLine) coreDraw(region Region, ctx DrawingContext) {
	if region.Size() == 0 {
		glog.WarningLogger().Warn("TextLine.coreDraw given empty region; no-oping", "w.text", w.text)
		return
	}
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	gl.Color4d(1.0, 1.0, 1.0, 1.0)
	w.Render_region = region

	glog.TraceLogger().Trace("coreDraw", "w.Render_region", w.Render_region, "text", w.GetText())
	{
		r, g, b, a := w.color.RGBA()
		gl.Color4d(float64(r)/65535, float64(g)/65535, float64(b)/65535, float64(a)/65535)
	}

	height := w.Render_region.Dy
	target := w.Render_region.Point
	glog.TraceLogger().Trace("target", "target", target)
	d := ctx.GetDictionary(w.font_id)
	shaders := ctx.GetShaders("glop.font")
	d.RenderString(w.text, target, height, Left, shaders)
}
