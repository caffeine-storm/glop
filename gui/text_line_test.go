package gui_test

import (
	"testing"

	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render/rendertest"
)

func TestTextLine(t *testing.T) {
	t.Run("Can make a 'lol' line", func(t *testing.T) {
		renderQueue := rendertest.MakeDiscardingRenderQueue()
		dict := gui.LoadDictionaryForTest(renderQueue, &gui.ConstDimser{}, glog.VoidLogger())
		gui.AddDictForTest("glop.font", dict)
		textLine := gui.MakeTextLine("glop.font", "lol", 42, 1, 1, 1, 1)
		if textLine == nil {
			t.Fatalf("got a nil TextLine back :(")
		}
	})
}
