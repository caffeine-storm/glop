package gui

type TabFrame struct {
	StubDrawFocuseder
	BasicZone
	StandardParent
	active int
}

func MakeTabFrame(ws []Widget) *TabFrame {
	var frame TabFrame
	for _, w := range ws {
		frame.Children = append(frame.Children, w)
	}
	return &frame
}

func (w *TabFrame) String() string {
	return "tab frame"
}

func (w *TabFrame) SelectTab(n int) {
	if n >= 0 && n < len(w.Children) {
		w.active = n
	}
}

func (w *TabFrame) SelectedTab() int {
	return w.active
}

func (w *TabFrame) Respond(gui *Gui, group EventGroup) bool {
	if gui.IsMouseEvent(group) {
		var p Point
		p.X, p.Y = gui.GetMousePosition(group)
		if !p.Inside(w.Rendered()) {
			return false
		}
	}
	return w.Children[w.active].Respond(gui, group)
}

func (w *TabFrame) Think(gui *Gui, t int64) {
	w.Request_dims = Dims{}
	for i := range w.Children {
		w.Children[i].Think(gui, t)
		dims := w.Children[i].Requested()
		if dims.Dx > w.Request_dims.Dx {
			w.Request_dims.Dx = dims.Dx
		}
		if dims.Dy > w.Request_dims.Dy {
			w.Request_dims.Dy = dims.Dy
		}
	}
}

func (w *TabFrame) Draw(region Region, ctx DrawingContext) {
	tab := w.Children[w.active]
	tab.Draw(region, ctx)
	w.Render_region = tab.Rendered()
}
