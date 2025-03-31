package gui

import (
	"fmt"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
)

type Point struct {
	X, Y int
}

func (p Point) Add(q Point) Point {
	return Point{
		X: p.X + q.X,
		Y: p.Y + q.Y,
	}
}
func (p Point) Inside(r Region) bool {
	if p.X < r.X {
		return false
	}
	if p.Y < r.Y {
		return false
	}
	if p.X >= r.X+r.Dx {
		return false
	}
	if p.Y >= r.Y+r.Dy {
		return false
	}
	return true
}

type Dims struct {
	Dx, Dy int
}

// TODO(tmckee): are we sure we want to embed Point and Dims? Seems more like
// they should just be named fields.
type Region struct {
	Point
	Dims
}

func (r Region) Add(p Point) Region {
	return Region{
		r.Point.Add(p),
		r.Dims,
	}
}

// Returns a region that is no larger than r that fits inside t.  The region
// will be located as closely as possible to r.
func (r Region) Fit(t Region) Region {
	if r.Dx > t.Dx {
		r.Dx = t.Dx
	}
	if r.Dy > t.Dy {
		r.Dy = t.Dy
	}
	if r.X < t.X {
		r.X = t.X
	}
	if r.Y < t.Y {
		r.Y = t.Y
	}
	if r.X+r.Dx > t.X+t.Dx {
		r.X -= (r.X + r.Dx) - (t.X + t.Dx)
	}
	if r.Y+r.Dy > t.Y+t.Dy {
		r.Y -= (r.Y + r.Dy) - (t.Y + t.Dy)
	}
	return r
}

func (r Region) Isect(s Region) Region {
	if r.X < s.X {
		r.Dx -= s.X - r.X
		r.X = s.X
	}
	if r.Y < s.Y {
		r.Dy -= s.Y - r.Y
		r.Y = s.Y
	}
	if r.X+r.Dx > s.X+s.Dx {
		r.Dx -= (r.X + r.Dx) - (s.X + s.Dx)
	}
	if r.Y+r.Dy > s.Y+s.Dy {
		r.Dy -= (r.Y + r.Dy) - (s.Y + s.Dy)
	}
	if r.Dx < 0 {
		r.Dx = 0
	}
	if r.Dy < 0 {
		r.Dy = 0
	}
	return r
}

func (r Region) Size() int {
	return r.Dx * r.Dy
}

// Need a global stack of regions because opengl only handles pushing/popping
// the state of the enable bits for each clip plane, not the planes themselves
var clippers []Region

// If we just declared this in setClipPlanes it would get allocated on the heap
// because we have to take the address of it to pass it to opengl.  By having
// it here we avoid that allocation - it amounts to a lot of someone is calling
// this every frame.
var eqs [4][4]float64

func (r Region) setClipPlanes() {
	eqs[0][0], eqs[0][1], eqs[0][2], eqs[0][3] = 1, 0, 0, -float64(r.X)
	eqs[1][0], eqs[1][1], eqs[1][2], eqs[1][3] = -1, 0, 0, float64(r.X+r.Dx)
	eqs[2][0], eqs[2][1], eqs[2][2], eqs[2][3] = 0, 1, 0, -float64(r.Y)
	eqs[3][0], eqs[3][1], eqs[3][2], eqs[3][3] = 0, -1, 0, float64(r.Y+r.Dy)
	gl.ClipPlane(gl.CLIP_PLANE0, eqs[0][:])
	gl.ClipPlane(gl.CLIP_PLANE1, eqs[1][:])
	gl.ClipPlane(gl.CLIP_PLANE2, eqs[2][:])
	gl.ClipPlane(gl.CLIP_PLANE3, eqs[3][:])
}

func (r Region) PushClipPlanes() {
	if len(clippers) == 0 {
		gl.Enable(gl.CLIP_PLANE0)
		gl.Enable(gl.CLIP_PLANE1)
		gl.Enable(gl.CLIP_PLANE2)
		gl.Enable(gl.CLIP_PLANE3)
		r.setClipPlanes()
		clippers = append(clippers, r)
	} else {
		cur := clippers[len(clippers)-1]
		clippers = append(clippers, r.Isect(cur))
		clippers[len(clippers)-1].setClipPlanes()
	}
}
func (r Region) PopClipPlanes() {
	clippers = clippers[0 : len(clippers)-1]
	if len(clippers) == 0 {
		gl.Disable(gl.CLIP_PLANE0)
		gl.Disable(gl.CLIP_PLANE1)
		gl.Disable(gl.CLIP_PLANE2)
		gl.Disable(gl.CLIP_PLANE3)
	} else {
		clippers[len(clippers)-1].setClipPlanes()
	}
}

//func (r Region) setViewport() {
//  gl.Viewport(r.Point.X, r.Point.Y, r.Dims.Dx, r.Dims.Dy)
//}

type Zone interface {
	// Returns the dimensions that this Widget would like available to
	// render itself.  A Widget should only update the value it returns from
	// this method when its Think() method is called.
	Requested() Dims

	// Returns ex,ey, where ex and ey indicate whether this Widget is
	// capable of expanding along the X and Y axes, respectively.
	Expandable() (bool, bool)

	// Returns the region that this Widget used to render itself the last
	// time it was rendered.  Should be completely contained within the
	// region that was passed to it on its last call to Draw.
	Rendered() Region
}

type EventGroup struct {
	gin.EventGroup
	DispatchedToFocussedWidget bool
}

type WidgetParent interface {
	AddChild(w Widget)
	RemoveChild(w Widget)
	GetChildren() []Widget
}

type DrawingContext interface {
	GetDictionary(fontname string) *Dictionary
	GetShaders(fontname string) *render.ShaderBank
	GetLogger() glog.Logger
}

type UpdateableDrawingContext interface {
	DrawingContext
	SetDictionary(fontname string, d *Dictionary)
	SetShaders(fontname string, b *render.ShaderBank)
}

type EventHandlingContext interface {
	IsMouseEvent(grp EventGroup) bool
	GetMousePosition() (int, int)
	LeftButton() bool
	MiddleButton() bool
	RightButton() bool
}

type Widget interface {
	Zone

	// Called regularly with a timestamp and a reference to the root widget.
	Think(*Gui, int64)

	// Returns true if this widget or any of its children consumed the
	// event group
	Respond(*Gui, EventGroup) bool

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
	Think(*Gui, int64)
	Respond(*Gui, EventGroup) (consume bool)
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
	if gui.IsMouseEvent(event_group) {
		var p Point
		p.X, p.Y = gui.GetMousePosition()
		if !p.Inside(w.Rendered()) {
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

type BasicZone struct {
	Request_dims  Dims
	Render_region Region
	Ex, Ey        bool
}

func (bz BasicZone) Requested() Dims {
	return bz.Request_dims
}
func (bz BasicZone) Rendered() Region {
	return bz.Render_region
}
func (bz BasicZone) Expandable() (bool, bool) {
	return bz.Ex, bz.Ey
}

type CollapsableZone struct {
	Collapsed     bool
	Request_dims  Dims
	Render_region Region
	Ex, Ey        bool
}

func (cz CollapsableZone) Requested() Dims {
	if cz.Collapsed {
		return Dims{}
	}
	return cz.Request_dims
}
func (cz CollapsableZone) Rendered() Region {
	if cz.Collapsed {
		return Region{Point: cz.Render_region.Point}
	}
	return cz.Render_region
}
func (cz *CollapsableZone) Expandable() (bool, bool) {
	if cz.Collapsed {
		return false, false
	}
	return cz.Ex, cz.Ey
}

// Embed a Clickable object to run a specified function when the widget
// is clicked and run a specified function.
type Clickable struct {
	on_click func(EventHandlingContext, int64)
}

func (c Clickable) DoRespond(ctx EventHandlingContext, event_group EventGroup) (bool, bool) {
	event := event_group.Events[0]
	if event.Type == gin.Press && event.Key.Id() == gin.AnyMouseLButton {
		c.on_click(ctx, event_group.Timestamp)
		return true, false
	}
	return false, false
}

type StubDrawFocuseder struct{}

func (n StubDrawFocuseder) DrawFocused(Region, DrawingContext) {}

type StubDoThinker struct{}

func (n StubDoThinker) DoThink(int64, bool) {}

type StubDoResponder struct{}

func (n StubDoResponder) DoRespond(EventHandlingContext, EventGroup) (bool, bool) {
	return false, false
}

type Childless struct{}

func (c Childless) GetChildren() []Widget { return nil }

// Wrappers are used to wrap existing widgets inside another widget to add some
// specific behavior (like making it hideable).  This can also be done by creating
// a new widget and embedding the appropriate structs, but sometimes this is more
// convenient.
type Wrapper struct {
	Child Widget
}

func (w Wrapper) GetChildren() []Widget { return []Widget{w.Child} }
func (w Wrapper) Draw(region Region, ctx DrawingContext) {
	w.Child.Draw(region, ctx)
}

type StandardParent struct {
	Children []Widget
}

func (s *StandardParent) GetChildren() []Widget {
	return s.Children
}
func (s *StandardParent) AddChild(w Widget) {
	s.Children = append(s.Children, w)
}
func (s *StandardParent) RemoveChild(w Widget) {
	for i := range s.Children {
		if s.Children[i] == w {
			s.Children[i] = s.Children[len(s.Children)-1]
			s.Children = s.Children[0 : len(s.Children)-1]
			return
		}
	}
}
func (s *StandardParent) ReplaceChild(old, new Widget) {
	for i := range s.Children {
		if s.Children[i] == old {
			s.Children[i] = new
			return
		}
	}
}
func (s *StandardParent) RemoveAllChildren() {
	s.Children = s.Children[0:0]
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

type ButtonPressType int

const (
	ButtonPressTypeUp = iota
	ButtonPressTypeDown
	ButtonPressTypeStart
	ButtonPressTypeEnd
)

func nextPressType(current ButtonPressType, now gin.MouseEventType) ButtonPressType {
	switch now {
	case gin.MouseEventTypeClick:
		switch current {
		case ButtonPressTypeUp:
			return ButtonPressTypeStart
		case ButtonPressTypeDown:
			return ButtonPressTypeDown
		case ButtonPressTypeStart:
			return ButtonPressTypeDown
		case ButtonPressTypeEnd:
			return ButtonPressTypeStart
		}
	case gin.MouseEventTypeMove:
		fallthrough
	case gin.MouseEventTypeWheel:
		return current
	}

	panic(fmt.Errorf("unknown MouseEventType: %v", now))
}

type mouseState struct {
	X, Y                int
	Left, Middle, Right ButtonPressType
}

func (ms *mouseState) GetPosition() (int, int) {
	return ms.X, ms.Y
}

func (ms *mouseState) Update(mouseEvent gin.MouseEvent) {
	ms.X = mouseEvent.X
	ms.Y = mouseEvent.Y

	ms.Left = nextPressType(ms.Left, mouseEvent.Type)
}

type Gui struct {
	root rootWidget

	dictionaries map[string]*Dictionary
	shaders      map[string]*render.ShaderBank

	// Stack of widgets that have focus
	focus []Widget

	mouseState mouseState
	logger     glog.Logger
}

var _ DrawingContext = (*Gui)(nil)
var _ UpdateableDrawingContext = (*Gui)(nil)
var _ EventHandlingContext = (*Gui)(nil)

type MissingFontError struct {
	error
}

func (g *Gui) GetDictionary(fontId string) *Dictionary {
	ret, ok := g.dictionaries[fontId]
	if !ok {
		panic(MissingFontError{fmt.Errorf("no registered font with id %q", fontId)})
	}
	return ret
}

func (g *Gui) GetShaders(fontname string) *render.ShaderBank {
	ret, ok := g.shaders[fontname]
	if !ok {
		panic(MissingFontError{fmt.Errorf("no registered shaders for id %q", fontname)})
	}
	return ret
}

func (g *Gui) GetLogger() glog.Logger {
	return g.logger
}

func (g *Gui) SetDictionary(fontname string, d *Dictionary) {
	g.dictionaries[fontname] = d
}

func (g *Gui) SetShaders(fontname string, b *render.ShaderBank) {
	g.shaders[fontname] = b
}

func Make(dims Dims, dispatcher gin.EventDispatcher) (*Gui, error) {
	return MakeLogged(dims, dispatcher, glog.VoidLogger())
}

func MakeLogged(dims Dims, dispatcher gin.EventDispatcher, logger glog.Logger) (*Gui, error) {
	// Note that, since each Gui should only be used in once RenderQueue, we
	// don't have to worry about font name collisions here.
	g := Gui{
		dictionaries: map[string]*Dictionary{},
		shaders:      map[string]*render.ShaderBank{},
		logger:       logger,
	}
	g.root.EmbeddedWidget = &BasicWidget{CoreWidget: &g.root}
	g.root.Request_dims = dims
	g.root.Render_region.Dims = dims
	dispatcher.RegisterEventListener(&g)
	dispatcher.AddMouseListener(g.UpdateMouseState)
	return &g, nil
}

func (g *Gui) GetWindowDimensions() Dims {
	return g.root.Request_dims
}

func (g *Gui) Draw() {
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	region := g.root.Render_region
	gl.Ortho(float64(region.X), float64(region.X+region.Dx), float64(region.Y), float64(region.Y+region.Dy), 1000, -1000)
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	g.root.Draw(region, g)
	if g.FocusWidget() != nil {
		g.FocusWidget().DrawFocused(region, g)
	}
}

func (g *Gui) Think(t int64) {
	g.root.Think(g, t)
}

func (g *Gui) UpdateMouseState(mouseEvent gin.MouseEvent) {
	g.mouseState.Update(mouseEvent)
}

func (g *Gui) HandleEventGroup(gin_group gin.EventGroup) {
	event_group := EventGroup{gin_group, false}

	// If there is one or more focused widgets, tell the top-of the focus-stack
	// to 'Respond' first.
	if len(g.focus) > 0 {
		event_group.DispatchedToFocussedWidget = true
		consume := g.focus[len(g.focus)-1].Respond(g, event_group)
		if consume {
			// If the focused widget consumed the event, we're done.
			return
		}
		event_group.DispatchedToFocussedWidget = false
	}

	// Without having consumed the event above, give the tree of widgets under
	// 'root' a shot at handling the event.
	g.root.Respond(g, event_group)
}

func (g *Gui) AddChild(w Widget) {
	g.root.AddChild(w)
}

func (g *Gui) RemoveChild(w Widget) {
	g.root.RemoveChild(w)
}

func (g *Gui) TakeFocus(w Widget) {
	if len(g.focus) == 0 {
		g.focus = append(g.focus, nil)
	}
	g.focus[len(g.focus)-1] = w
}

func (g *Gui) DropFocus() {
	g.focus = g.focus[0 : len(g.focus)-1]
}

func (g *Gui) FocusWidget() Widget {
	if len(g.focus) == 0 {
		return nil
	}
	return g.focus[len(g.focus)-1]
}

func (g *Gui) IsMouseEvent(grp EventGroup) bool {
	if len(grp.Events) < 0 {
		return false
	}

	evt := grp.Events[0]
	return evt.Key.Id().Device.Type == gin.DeviceTypeMouse
}

func (g *Gui) GetMousePosition() (int, int) {
	return g.mouseState.GetPosition()
}

func stateToButtonFlag(tp ButtonPressType) bool {
	return tp == ButtonPressTypeDown || tp == ButtonPressTypeStart
}

func (g *Gui) LeftButton() bool {
	return stateToButtonFlag(g.mouseState.Left)
}

func (g *Gui) MiddleButton() bool {
	return stateToButtonFlag(g.mouseState.Middle)
}

func (g *Gui) RightButton() bool {
	return stateToButtonFlag(g.mouseState.Right)
}
