package gui

import (
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/render"
)

type cursor struct {
	index  int     // Index before which the cursor is placed
	pos    float64 // position of the cursor in pixels from the left hand side
	moved  bool    // whether or not the cursor has been moved recently
	on     bool    // whether or not the curosr is showing
	period int64   // how fast the cursor should blink
	start  int64   // last time cursor.on was set to true
}
type TextEditLine struct {
	TextLine
	cursor cursor
}

var shift_mapping map[gin.KeyId]byte

func init() {
	lower := "abcdefghijklmnopqrstuvwxyz0123456789-=[];,./\\'"
	upper := "ABCDEFGHIJKLMNOPQRSTUVWXYZ)!@#$%^&*(_+{}:<>?|\""
	shift_mapping = make(map[gin.KeyId]byte)
	for i := range lower {
		lowerKeyId := gin.KeyId{
			Device: gin.DeviceId{
				Type:  gin.DeviceTypeKeyboard,
				Index: gin.DeviceIndexAny,
			},
			Index: gin.KeyIndex(i),
		}
		shift_mapping[gin.KeyId(lowerKeyId)] = upper[i]
	}
	shift_mapping[gin.AnySpace] = 0
}

func (w *TextEditLine) String() string {
	return "text edit line"
}

func MakeTextEditLine(fontId, text string, width int, r, g, b, a float64) *TextEditLine {
	var w TextEditLine
	w.TextLine = *MakeTextLine(fontId, text, width, r, g, b, a)
	w.EmbeddedWidget = &BasicWidget{CoreWidget: &w}

	w.scale = 1.0
	w.cursor.index = len(w.text)
	w.cursor.pos = w.findOffsetAtIndex(w.cursor.index)
	w.cursor.period = 500 // half a second
	return &w
}

func (w *TextEditLine) findIndexAtOffset(offset int) int {
	low := 0
	high := 1
	var low_off, high_off float64
	for high < len(w.text) && high_off < float64(offset) {
		low = high
		low_off = high_off
		high++
		high_off = w.findOffsetAtIndex(high)
	}
	if float64(offset)-low_off < high_off-float64(offset) {
		return low
	}
	return high
}

func (w *TextEditLine) findOffsetAtIndex(index int) float64 {
	// pt := freetype.Pt(0, 0)
	if index > len(w.text) {
		index = len(w.text)
	}
	if index < 0 {
		index = 0
	}
	// TODO(tmckee): XXX: STUBBED!
	// adv, _ := w.context.DrawString(w.text[:index], pt)
	// return float64(adv.X>>8) * w.scale
	return 42
}

func (w *TextEditLine) DoThink(t int64, focus bool) {
	changed := w.text != w.next_text
	w.TextLine.DoThink(t, false)
	if focus && w.cursor.start == 0 {
		w.cursor.start = t
		w.cursor.on = true
	}
	if !focus {
		w.cursor.start = 0
		w.cursor.on = false
	}
	if w.cursor.start > 0 {
		w.cursor.on = ((t-w.cursor.start)/w.cursor.period)%2 == 0
	}
	if w.cursor.moved || changed {
		w.cursor.pos = w.findOffsetAtIndex(w.cursor.index)
		w.cursor.moved = false
	}
}

func (w *TextEditLine) IsBeingEdited() bool {
	return w.cursor.start > 0
}

func (w *TextEditLine) SetText(text string) {
	max := (w.cursor.index >= len(w.text))
	w.TextLine.SetText(text)
	if max || w.cursor.index > len(w.text) {
		w.cursor.index = len(w.text)
	}
}

func characterFromEventGroup(event_group EventGroup) byte {
	for _, event := range event_group.Events {
		if v, ok := shift_mapping[event.Key.Id()]; ok {
			// if gin.In().GetKeyById(gin.EitherShift).IsDown() {
			if gin.In().GetKeyById(gin.AnyLeftShift).IsDown() ||
				gin.In().GetKeyById(gin.AnyRightShift).IsDown() {
				return v
			} else {
				return byte(event.Key.Id().Index)
			}
		}
	}
	return 0
}

func (w *TextEditLine) DoRespond(ctx EventHandlingContext, event_group EventGroup) (consume, change_focus bool) {
	if w.cursor.index > len(w.text) {
		w.cursor.index = len(w.text)
	}
	event := event_group.PrimaryEvent()
	if !event.IsPress() {
		return
	}
	key_id := event.Key.Id()
	if event_group.DispatchedToFocussedWidget {
		if key_id == gin.AnyEscape || key_id == gin.AnyReturn {
			change_focus = true
			return
		}
		if event_group.IsPressed(gin.AnyBackspace) {
			if len(w.text) > 0 && w.cursor.index > 0 {
				var pre, post string
				if w.cursor.index > 0 {
					pre = w.text[0 : w.cursor.index-1]
				}
				if w.cursor.index < len(w.text) {
					post = w.text[w.cursor.index:]
				}
				w.SetText(pre + post)
				w.cursor.index--
				w.cursor.moved = true
			}
		} else if v := characterFromEventGroup(event_group); v != 0 {
			w.SetText(w.text[0:w.cursor.index] + string([]byte{v}) + w.text[w.cursor.index:])
			w.cursor.index++
			w.cursor.moved = true
		} else if key_id == gin.AnyMouseLButton {
			// TODO(#28): probably want to look at the Y co-ordinate too, right?
			if pt, ok := ctx.UseMousePosition(event_group); ok {
				cx := w.TextLine.Render_region.X
				w.cursor.index = w.findIndexAtOffset(pt.X - cx)
				w.cursor.moved = true
			}
		} else if event_group.IsPressed(gin.AnyLeft) {
			if w.cursor.index > 0 {
				w.cursor.index--
				w.cursor.moved = true
			}
		} else if event_group.IsPressed(gin.AnyRight) {
			if w.cursor.index < len(w.text) {
				w.cursor.index++
				w.cursor.moved = true
			}
		}
		consume = true
	} else {
		change_focus = event.Key.Id() == gin.AnyMouseLButton
	}
	return
}

func (w *TextEditLine) Draw(region Region, ctx DrawingContext) {
	region.PushClipPlanes()
	defer region.PopClipPlanes()
	gl.Disable(gl.TEXTURE_2D)
	render.WithColour(0.3, 0.3, 0.3, 0.9, func() {
		gl.Begin(gl.QUADS)
		gl.Vertex2i(region.X+1, region.Y+1)
		gl.Vertex2i(region.X+1, region.Y-1+region.Dy)
		gl.Vertex2i(region.X-1+region.Dx, region.Y-1+region.Dy)
		gl.Vertex2i(region.X-1+region.Dx, region.Y+1)
		gl.End()
		w.TextLine.preDraw(region, ctx)
		w.TextLine.coreDraw(region, ctx)
		gl.Disable(gl.TEXTURE_2D)
		if w.cursor.on {
			gl.Color3d(1, 0.3, 0)
		} else {
			gl.Color3d(0.5, 0.3, 0)
		}
		gl.Begin(gl.LINES)
		gl.Vertex2i(region.X+int(w.cursor.pos), region.Y)
		gl.Vertex2i(region.X+int(w.cursor.pos), region.Y+region.Dy)
		gl.End()
		w.TextLine.postDraw(region, ctx)
	})
}
