package gui

type SelectBox struct {
	Table
	selected int
}

// TODO(tmckee#42): 'vertical' is sucky; use MakeVerticalSelectBox,
// MakeHorizontalSelectBox and delegate to makeSelectBox(options, table)
func MakeSelectBox(options []SelectableWidget, vertical bool) *SelectBox {
	var sb SelectBox
	if vertical {
		sb.Table = MakeVerticalTable()
	} else {
		sb.Table = MakeHorizontalTable()
	}
	for i := range options {
		option := options[i]
		option.SetSelectFunc(func(EventHandlingContext, int64) {
			sb.SetSelectedOption(option.GetData())
		})
		sb.AddChild(option)
		option.SetSelected(false)
	}
	sb.SetSelectedIndex(-1)
	return &sb
}

func MakeSelectTextBox(text_options []string, width int) *SelectBox {
	options := make([]SelectableWidget, len(text_options))
	for i := range options {
		options[i] = makeTextOption(text_options[i], width)
	}
	return MakeSelectBox(options, true)
}

func MakeSelectImageBox(paths []string, names []string) *SelectBox {
	options := make([]SelectableWidget, len(paths))
	for i := range options {
		options[i] = makeImageOption(paths[i], names[i])
	}
	return MakeSelectBox(options, false)
}

func (w *SelectBox) String() string {
	return "select box"
}

func (w *SelectBox) GetSelectedIndex() int {
	return w.selected
}

func (w *SelectBox) SetSelectedIndex(index int) {
	w.selectIndex(index)
}

func (w *SelectBox) GetSelectedOption() interface{} {
	if w.selected == -1 {
		return ""
	}
	return w.GetChildren()[w.selected].(SelectableWidget).GetData()
}

func (w *SelectBox) SetSelectedOption(option interface{}) {
	for i := range w.GetChildren() {
		if w.GetChildren()[i].(SelectableWidget).GetData() == option {
			w.selectIndex(i)
			return
		}
	}
	w.selectIndex(-1)
}

func (w *SelectBox) selectIndex(index int) {
	children := w.GetChildren()
	if w.selected >= 0 && w.selected < len(children) {
		children[w.selected].(SelectableWidget).SetSelected(false)
	}
	if index < 0 || index >= len(children) {
		index = -1
	} else {
		children[index].(SelectableWidget).SetSelected(true)
	}
	w.selected = index
}
