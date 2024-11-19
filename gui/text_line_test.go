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
	Convey("TextLine should work", t, TextLineSpecs)
}

func TextLineSpecs() {
	screenWidth, screenHeight := 200, 50

	Convey("Can make a 'lol' line", func() {
		renderQueue := rendertest.MakeDiscardingRenderQueue()
		dict := gui.LoadDictionaryForTest(renderQueue, &gui.ConstDimser{}, glog.VoidLogger())
		gui.AddDictForTest("glop.font", dict, &render.ShaderBank{})
		textLine := gui.MakeTextLine("glop.font", "lol", 42, 1, 1, 1, 1)
		So(textLine, ShouldNotBeNil)
	})

	Convey("TextLine draws its text", func() {
		rendertest.WithGlForTest(screenWidth, screenHeight, func(sys system.System, queue render.RenderQueueInterface) {
			// TODO(tmckee): XXX: having to remember to gui.Init is ... sad-making
			gui.Init(queue)
			dimser := &gui.ConstDimser{Value: gui.Dims{screenWidth, screenHeight}}
			dict := gui.LoadDictionaryForTest(queue, dimser, glog.DebugLogger())
			g := MakeStubbedGui()

			var shaderBank *render.ShaderBank
			queue.Queue(func(rqs render.RenderQueueState) {
				shaderBank = rqs.Shaders()
			})
			queue.Purge()

			gui.AddDictForTest("glop.font", dict, shaderBank)

			textLine := gui.MakeTextLine("glop.font", "some text", 32, 1, 1, 1, 1)

			queue.Queue(func(render.RenderQueueState) {
				textLine.Draw(gui.Region{
					Point: gui.Point{},
					Dims:  gui.Dims{screenWidth, screenHeight},
				}, g)
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeText, "some-text")
		})
	})
}
