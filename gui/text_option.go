package gui

type textOption struct {
	TextLine
	selectableOption
}

func (w *textOption) DoRespond(ctx EventHandlingContext, event_group EventGroup) (consume, change_focus bool) {
	w.selectableOption.DoRespond(ctx, event_group)
	return
}

func (w *textOption) SetSelected(selected bool) {
	if selected {
		w.SetColor(0.9, 1, 0.9, 1)
	} else {
		w.SetColor(0.6, 0.4, 0.4, 1)
	}
}

func makeTextOption(text string, width int) SelectableWidget {
	var so textOption
	so.TextLine = *MakeTextLine("standard_18", text, width, 1, 1, 1, 1)
	so.data = text
	so.EmbeddedWidget = &BasicWidget{CoreWidget: &so}
	return &so
}
