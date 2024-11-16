package gui_test

import (
	"testing"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
)

func TestTextEditLine(t *testing.T) {
	t.Run("can construct", func(t *testing.T) {
		renderQueue := rendertest.MakeDiscardingRenderQueue()
		dict := gui.LoadDictionaryForTest(renderQueue, &gui.ConstDimser{}, glog.VoidLogger())
		gui.AddDictForTest("glop.font", dict, &render.ShaderBank{})
		result := gui.MakeTextEditLine("glop.font", "some text for editing", 42, 1, 1, 1, 1)
		if result == nil {
			t.Fatalf("gui.MakeTextEditLine returned nil!")
		}
	})
}
