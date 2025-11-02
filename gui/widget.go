package gui

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime"

	"github.com/caffeine-storm/gl"
	"github.com/caffeine-storm/glu"
)

type Widget interface {
	Zone
	Responder

	// Called regularly with a timestamp and a reference to the root widget.
	Think(*Gui, int64)

	Draw(Region, DrawingContext)
	DrawFocused(Region, DrawingContext)
	String() string
}

type CoreWidget interface {
	Zone

	DoThink(dt int64, isFocused bool)

	// If change_focus is true, then the EventGroup will be consumed,
	// regardless of the value of consume
	DoRespond(EventHandlingContext, EventGroup) (consume, change_focus bool)

	Draw(Region, DrawingContext)
	DrawFocused(Region, DrawingContext)

	GetChildren() []Widget
	String() string
}

type EmbeddedWidget interface {
	Responder
	Think(*Gui, int64)
}

type WidgetParent interface {
	AddChild(w Widget)
	RemoveChild(w Widget)
	GetChildren() []Widget
}

// TODO(tmckee): is a BasicWidget just a way to implement 'Widget' for a
// CoreWidget?
type BasicWidget struct {
	CoreWidget
}

func (w *BasicWidget) Think(gui *Gui, t int64) {
	kids := w.GetChildren()
	for i := range kids {
		kids[i].Think(gui, t)
	}
	w.DoThink(t, w == gui.FocusWidget())
}

func (w *BasicWidget) Respond(gui *Gui, event_group EventGroup) bool {
	if mpos, ok := gui.UseMousePosition(event_group); ok {
		if !mpos.Inside(w.Rendered()) {
			return false
		}
	}
	consume, change_focus := w.DoRespond(gui, event_group)

	if change_focus {
		if event_group.DispatchedToFocussedWidget {
			gui.DropFocus()
		} else {
			gui.TakeFocus(w)
		}
		return true
	}
	if consume {
		return true
	}

	kids := w.GetChildren()
	for i := len(kids) - 1; i >= 0; i-- {
		if kids[i].Respond(gui, event_group) {
			return true
		}
	}
	return false
}

type ImageBox struct {
	EmbeddedWidget
	StubDoResponder
	StubDoThinker
	StubDrawFocuseder
	BasicZone
	Childless

	active     bool
	texture    gl.Texture
	r, g, b, a float64
}

func MakeImageBox() *ImageBox {
	var ib ImageBox
	ib.EmbeddedWidget = &BasicWidget{CoreWidget: &ib}
	runtime.SetFinalizer(&ib, freeTexture)
	ib.r, ib.g, ib.b, ib.a = 1, 1, 1, 1
	return &ib
}

func (w *ImageBox) String() string {
	return "image box"
}

func (w *ImageBox) SetShading(r, g, b, a float64) {
	w.r, w.g, w.b, w.a = r, g, b, a
}

func freeTexture(w *ImageBox) {
	if w.active {
		w.texture.Delete()
		w.active = false
	}
	w.texture = 0
}

// Does not take ownserhip of the texture, you must still free the texture
// when you are done with it.
func (w *ImageBox) SetImageByTexture(texture gl.Texture, dx, dy int) {
	w.UnsetImage()
	w.texture = texture
	w.Request_dims.Dx = dx
	w.Request_dims.Dy = dy
	w.active = false
}

func (w *ImageBox) UnsetImage() {
	freeTexture(w)
}

func (w *ImageBox) SetImage(path string) {
	w.UnsetImage()
	data, err := os.Open(path)
	if err != nil {
		// TODO: Log error
		return
	}

	var img image.Image
	img, _, err = image.Decode(data)
	if err != nil {
		// TODO: Log error
		return
	}

	w.Request_dims.Dx = img.Bounds().Dx()
	w.Request_dims.Dy = img.Bounds().Dy()
	canvas := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	for y := 0; y < canvas.Bounds().Dy(); y++ {
		for x := 0; x < canvas.Bounds().Dx(); x++ {
			r, g, b, a := img.At(x, y).RGBA()
			base := 4*x + canvas.Stride*y
			canvas.Pix[base] = uint8(r)
			canvas.Pix[base+1] = uint8(g)
			canvas.Pix[base+2] = uint8(b)
			canvas.Pix[base+3] = uint8(a)
		}
	}

	// TODO(tmckee:clean): reuse texture manager things here instead of
	// re-rolling our own.
	w.texture = gl.GenTexture()
	w.texture.Bind(gl.TEXTURE_2D)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	glu.Build2DMipmaps(gl.TEXTURE_2D, gl.RGBA, img.Bounds().Dx(), img.Bounds().Dy(), gl.RGBA, gl.UNSIGNED_BYTE, canvas.Pix)

	w.active = true
}

func (w *ImageBox) Draw(region Region, ctx DrawingContext) {
	w.Render_region = region

	// We check texture == 0 and not active because active only indicates if we
	// have a texture that we need to free later.  It's possible for us to have
	// a texture that someone else owns.
	if w.texture == 0 {
		return
	}

	w.texture.Bind(gl.TEXTURE_2D)
	gl.Enable(gl.BLEND)
	gl.Color4d(w.r, w.g, w.b, w.a)
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 0)
	gl.Vertex2i(region.X, region.Y)
	gl.TexCoord2f(0, -1)
	gl.Vertex2i(region.X, region.Y+region.Dy)
	gl.TexCoord2f(1, -1)
	gl.Vertex2i(region.X+region.Dx, region.Y+region.Dy)
	gl.TexCoord2f(1, 0)
	gl.Vertex2i(region.X+region.Dx, region.Y)
	gl.End()
}

type CollapseWrapper struct {
	EmbeddedWidget
	Wrapper
	CollapsableZone
	StubDoResponder
	StubDrawFocuseder
}

func MakeCollapseWrapper(w Widget) *CollapseWrapper {
	var cw CollapseWrapper
	cw.EmbeddedWidget = &BasicWidget{CoreWidget: &cw}
	cw.Child = w
	return &cw
}

func (w *CollapseWrapper) String() string {
	return "collapse wrapper"
}

func (w *CollapseWrapper) DoThink(int64, bool) {
	w.Request_dims = w.Child.Requested()
	w.Render_region = w.Child.Rendered()
}

func (w *CollapseWrapper) Draw(region Region, ctx DrawingContext) {
	if w.Collapsed {
		w.Render_region = Region{}
		return
	}
	w.Child.Draw(region, ctx)
	w.Render_region = region
}

type OptionContainer interface {
	SetSelectedOption(Widget)
}

type SelectableWidget interface {
	Widget

	// The selectable widget will call this function when clicked
	SetSelectFunc(func(EventHandlingContext, int64))

	SetSelected(bool)
	GetData() interface{}
}

type selectableOption struct {
	Clickable
	data   interface{}
	parent OptionContainer
}

func (so *selectableOption) GetData() interface{} {
	return so.data
}

func (so *selectableOption) SetSelectFunc(f func(EventHandlingContext, int64)) {
	so.on_click = f
}

type rootWidget struct {
	EmbeddedWidget
	StandardParent
	BasicZone
	StubDoResponder
	StubDoThinker
	StubDrawFocuseder
}

func (r *rootWidget) String() string {
	return "root"
}

func (r *rootWidget) Draw(region Region, ctx DrawingContext) {
	r.Render_region = region
	for i := range r.Children {
		r.Children[i].Draw(region, ctx)
	}
}
