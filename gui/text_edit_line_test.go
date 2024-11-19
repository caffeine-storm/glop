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

func TestRunTextEditLineSpecs(t *testing.T) {
	Convey("TextEditLine should work", t, TextEditLineSpecs)
}

func TextEditLineSpecs() {
	screenWidth, screenHeight := 200, 50

	Convey("Can make a 'lol' edit line", func() {
		renderQueue := rendertest.MakeDiscardingRenderQueue()
		dict := gui.LoadDictionaryForTest(renderQueue, glog.VoidLogger())
		g := MakeStubbedGui()
		gui.AddDictForTest(g, "glop.font", dict, &render.ShaderBank{})
		textLine := gui.MakeTextEditLine("glop.font", "lol", 42, 1, 1, 1, 1)
		So(textLine, ShouldNotBeNil)
	})

	Convey("TextEditLine draws its text", func() {
		rendertest.WithGlForTest(screenWidth, screenHeight, func(sys system.System, queue render.RenderQueueInterface) {
			// TODO(tmckee): XXX: having to remember to gui.Init is ... sad-making
			gui.Init(queue)
			dict := gui.LoadDictionaryForTest(queue, glog.DebugLogger())
			g := MakeStubbedGui()

			var shaderBank *render.ShaderBank
			queue.Queue(func(rqs render.RenderQueueState) {
				shaderBank = rqs.Shaders()
			})
			queue.Purge()

			gui.AddDictForTest(g, "glop.font", dict, shaderBank)

			textLine := gui.MakeTextEditLine("glop.font", "some text to edit", 32, 1, 1, 1, 1)

			queue.Queue(func(render.RenderQueueState) {
				textLine.Draw(gui.Region{
					Point: gui.Point{},
					Dims:  gui.Dims{screenWidth, screenHeight},
				}, g)
			})
			queue.Purge()

			So(queue, rendertest.ShouldLookLikeText, "some-edit-text")
		})
	})
}
