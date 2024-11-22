package gui_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

func TestRunTextLineSpecs(t *testing.T) {
	Convey("Drawing Lines of Text", t, TextLineSpecs)
}

const screenWidth, screenHeight = 200, 50

type GenericLine interface {
	Draw(gui.Region, gui.DrawingContext)
}

func GenericTextLineTest(text string, widgetBuilder func(text string) GenericLine) {
	Convey("Can make a 'lol' line", func() {
		renderQueue := rendertest.MakeDiscardingRenderQueue()
		dict := gui.LoadDictionaryForTest(renderQueue, glog.VoidLogger())
		g := MakeStubbedGui()
		g.SetDictionary("dict_10", dict)
		g.SetShaders("glop.font", &render.ShaderBank{})

		textLine := widgetBuilder("lol")
		So(textLine, ShouldNotBeNil)
	})

	Convey("TextLine draws the given text", func() {
		rendertest.WithGlForTest(screenWidth, screenHeight, func(sys system.System, queue render.RenderQueueInterface) {
			dict := gui.LoadDictionaryForTest(queue, glog.DebugLogger())
			g := MakeStubbedGui()

			var shaderBank *render.ShaderBank
			queue.Queue(func(rqs render.RenderQueueState) {
				shaderBank = rqs.Shaders()
			})
			queue.Purge()

			g.SetDictionary("dict_10", dict)
			g.SetShaders("glop.font", shaderBank)

			textLine := widgetBuilder(text)

			queue.Queue(func(render.RenderQueueState) {
				textLine.Draw(gui.Region{
					Point: gui.Point{},
					Dims:  gui.Dims{screenWidth, screenHeight},
				}, g)
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeText, text)
		})
	})
}

func TextLineSpecs() {
	Convey("TextLine can draw text", func() {
		GenericTextLineTest("some-text", func(text string) GenericLine {
			return gui.MakeTextLine("dict_10", text, 32, 1, 1, 1, 1)
		})
	})

	Convey("TextEditLine can draw text", func() {
		GenericTextLineTest("some-edit-text", func(text string) GenericLine {
			return gui.MakeTextEditLine("dict_10", text, 42, 1, 1, 1, 1)
		})
	})
}
