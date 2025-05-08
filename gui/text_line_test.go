package gui_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/gui/guitest"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
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
		dict := gui.LoadAndInitializeDictionaryForTest(renderQueue, glog.VoidLogger())
		g := guitest.MakeStubbedGui(gui.Dims{screenWidth, screenHeight})
		g.SetDictionary("dict_10", dict)
		g.SetShaders("glop.font", &render.ShaderBank{})

		textLine := widgetBuilder("lol")
		So(textLine, ShouldNotBeNil)
	})

	Convey("TextLine draws the given text", func() {
		rendertest.GlTest().WithSize(screenWidth, screenHeight).WithQueue().Run(func(queue render.RenderQueueInterface) {
			dict := gui.LoadAndInitializeDictionaryForTest(queue, glog.DebugLogger())
			g := guitest.MakeStubbedGui(gui.Dims{screenWidth, screenHeight})

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

func MultipleTextLineTest(widgetBuilder func(text string) GenericLine) {
	Convey("drawing more than one line", func() {
		line1 := widgetBuilder("first line")
		line2 := widgetBuilder("second line")
		line3 := widgetBuilder("third line")

		rendertest.GlTest().WithSize(screenWidth, screenHeight).WithQueue().Run(func(queue render.RenderQueueInterface) {
			dict := gui.LoadAndInitializeDictionaryForTest(queue, glog.DebugLogger())
			g := guitest.MakeStubbedGui(gui.Dims{screenWidth, screenHeight})

			var shaderBank *render.ShaderBank
			queue.Queue(func(rqs render.RenderQueueState) {
				shaderBank = rqs.Shaders()
			})
			queue.Purge()

			g.SetDictionary("dict_10", dict)
			g.SetShaders("glop.font", shaderBank)

			lineheight := screenHeight / 5
			queue.Queue(func(render.RenderQueueState) {
				line1.Draw(gui.Region{
					Point: gui.Point{0, 0},
					Dims:  gui.Dims{screenWidth, lineheight},
				}, g)
				line2.Draw(gui.Region{
					Point: gui.Point{0, lineheight * 2},
					Dims:  gui.Dims{screenWidth, lineheight},
				}, g)
				line3.Draw(gui.Region{
					Point: gui.Point{0, lineheight * 4},
					Dims:  gui.Dims{screenWidth, lineheight},
				}, g)
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeText, "multi-line")
		})
	})
}

func TextLineSpecs() {
	Convey("TextLine can draw text", func() {
		GenericTextLineTest("some-text", func(text string) GenericLine {
			return gui.MakeTextLine("dict_10", text, 32, 1, 1, 1, 1)
		})

		MultipleTextLineTest(func(text string) GenericLine {
			return gui.MakeTextLine("dict_10", text, 32, 1, 1, 1, 1)
		})
	})

	Convey("TextEditLine can draw text", func() {
		GenericTextLineTest("some-edit-text", func(text string) GenericLine {
			return gui.MakeTextEditLine("dict_10", text, 42, 1, 1, 1, 1)
		})
	})
}
