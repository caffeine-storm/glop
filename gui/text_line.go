package gui

import (
	"image/color"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
)

type TextLine struct {
	EmbeddedWidget
	Childless
	NonResponder
	NonFocuser
	BasicZone
	text      string
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

func nextPowerOf2(n uint32) uint32 {
	if n == 0 {
		return 1
	}
	for i := uint(0); i < 32; i++ {
		p := uint32(1) << i
		if n <= p {
			return p
		}
	}
	return 0
}

type Button struct {
	*TextLine
	Clickable
}

func MakeButton(fontId, text string, width int, r, g, b, a float64, f func(int64)) *Button {
	var btn Button
	btn.TextLine = MakeTextLine(fontId, text, width, r, g, b, a)
	btn.TextLine.EmbeddedWidget = &BasicWidget{CoreWidget: &btn}
	btn.on_click = f
	return &btn
}

// TODO(tmckee): we should take a font by reference instead of by
// stringified-name. That way, the compiler can check for us that the font is
// loaded.
func MakeTextLine(fontId, text string, width int, r, g, b, a float64) *TextLine {
	var w TextLine

	w.font_id = fontId
	w.text = text
	w.EmbeddedWidget = &BasicWidget{CoreWidget: &w}
	// w.SetFontSize(12) // TODO(tmckee) ... waat?
	w.SetColor(r, g, b, a)
	w.Request_dims = Dims{width, 35}
	return &w
}

func (w *TextLine) SetColor(r, g, b, a float64) {
	w.color = color.RGBA{
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

func (w *TextLine) preDraw(region Region, ctx DrawingContext) {
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
	req := w.Request_dims
	if req.Dx > region.Dx {
		req.Dx = region.Dx
	}
	if req.Dy > region.Dy {
		req.Dy = region.Dy
	}
	if req.Dx*region.Dy < req.Dy*region.Dx {
		req.Dy = (region.Dy * req.Dx) / region.Dx
	} else {
		req.Dx = (region.Dx * req.Dy) / region.Dy
	}
	w.Render_region.Dims = req
	w.Render_region.Point = region.Point

	glog.TraceLogger().Trace("coreDraw", "w.Request_dims", w.Request_dims, "w.Render_region", w.Render_region)
	{
		r, g, b, a := w.color.RGBA()
		gl.Color4d(float64(r)/65535, float64(g)/65535, float64(b)/65535, float64(a)/65535)
	}

	// TODO(tmckee): arbitrary!
	height := 12
	target := w.Render_region.Point
	target.Y = w.Render_region.Dims.Dy - target.Y
	target.Y += height
	glog.TraceLogger().Trace("target", "target", target)
	d := ctx.GetDictionary(w.font_id)
	shaders := ctx.GetShaders("glop.font")
	d.RenderString(w.text, target, height, Left, shaders)
}
