package gui

type imageOption struct {
	ImageBox
	selectableOption
}

func (w *imageOption) DoRespond(ctx EventHandlingContext, event_group EventGroup) (consume, change_focus bool) {
	w.selectableOption.DoRespond(ctx, event_group)
	return
}

func (w *imageOption) SetSelected(selected bool) {
	if selected {
		w.SetShading(1, 1, 1, 1)
	} else {
		w.SetShading(0.5, 0.5, 0.5, 0.9)
	}
}

func makeImageOption(path string, data interface{}) SelectableWidget {
	var sio imageOption
	sio.ImageBox = *MakeImageBox()
	sio.ImageBox.SetImage(path)
	sio.data = data
	sio.EmbeddedWidget = &BasicWidget{CoreWidget: &sio}
	return &sio
}
