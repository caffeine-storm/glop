package gui_test

import (
	"testing"

	"github.com/runningwild/glop/glog/glogtest"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/gui/guitest"
	"github.com/stretchr/testify/assert"
)

var dims = gui.Dims{13, 43}

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

	t.Run("can transform screen space to 'gui space' (a.k.a. normalized device coordinates", func(t *testing.T) {
		assert := assert.New(t)
		g := guitest.MakeStubbedGui(dims)
		ndc_x, ndc_y := g.ScreenToNDC(6, 21)

		assert.Equal(float32(0), ndc_x)
		assert.Equal(float32(0), ndc_y)
	})
}
