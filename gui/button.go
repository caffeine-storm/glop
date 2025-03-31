package gui

type Button struct {
	*TextLine
	Clickable
}

func MakeButton(fontId, text string, width int, r, g, b, a float64, f func(EventHandlingContext, int64)) *Button {
	var btn Button
	btn.TextLine = MakeTextLine(fontId, text, width, r, g, b, a)
	btn.TextLine.EmbeddedWidget = &BasicWidget{CoreWidget: &btn}
	btn.on_click = f
	return &btn
}
