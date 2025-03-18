package gui_test

import (
	"testing"

	"github.com/runningwild/glop/glog/glogtest"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/gui/guitest"
)

var dims = gui.Dims{13, 42}

type loggingWidget struct{}

func (*loggingWidget) Draw(dims gui.Dims, ctx gui.DrawingContext) {
	ctx.GetLogger().Info("Draw got called")
}

func givenAWidgetThatLogs() *loggingWidget {
	return &loggingWidget{}
}

func TestGui(t *testing.T) {
	t.Run("Make", func(t *testing.T) {
		_ = guitest.MakeStubbedGui(dims)
	})

	t.Run("make with logger", func(t *testing.T) {
		bufferedLogger := glogtest.NewBufferedLogger()

		guiInstance, err := gui.MakeLogged(dims, guitest.MakeStubbedEventDispatcher(), bufferedLogger)
		if err != nil {
			t.Fatalf("got unexpected error while gui.Make'ing: %v", err)
		}

		if guiInstance == nil {
			t.Fatalf("got no error from Make but got no gui either!")
		}
		t.Run("can log from a widget's Draw()", func(t *testing.T) {
			testWidget := givenAWidgetThatLogs()

			testWidget.Draw(dims, guiInstance)

			if !bufferedLogger.Contains("Draw got called") {
				t.Fatalf("the testWidget should have logged 'Draw got called' buffer: %v", bufferedLogger.String())
			}
		})
	})
}
