package gui_test

import (
	"testing"

	"github.com/runningwild/glop/glog/glogtest"
	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/gui/guitest"
	"github.com/stretchr/testify/assert"
)

var dims = gui.Dims{200, 400}

type loggingWidget struct{}

func (*loggingWidget) Draw(dims gui.Dims, ctx gui.DrawingContext) {
	ctx.GetLogger().Info("Draw got called")
}

func givenAWidgetThatLogs() *loggingWidget {
	return &loggingWidget{}
}

type testcase struct {
	name   string
	expect [2]float32
	input  [2]int
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
		// The floating point values returned need to be within epsilon of the
		// expected value. Since 'dims' is so small, epsilon is actually kinda big
		// in this case but we'll compute an epsilon here that enforces
		// not-out-by-more-than-half-a-pixel.
		epsilon_x := 1 / float64(dims.Dx)
		epsilon_y := 1 / float64(dims.Dy)

		g := guitest.MakeStubbedGui(dims)

		testcases := []testcase{
			{
				name:   "centre",
				expect: [2]float32{0, 0},
				input:  [2]int{dims.Dx / 2, dims.Dy / 2},
			},
			{
				name:   "bottom-left quadrant",
				expect: [2]float32{-0.8, -0.9},
				input:  [2]int{dims.Dx / 10, dims.Dy / 20},
			},
		}
		for _, testcase := range testcases {
			t.Run(testcase.name, func(t *testing.T) {
				assert := assert.New(t)
				ndc_x, ndc_y := g.ScreenToNDC(testcase.input[0], testcase.input[1])
				assert.InDelta(float64(testcase.expect[0]), float64(ndc_x), epsilon_x)
				assert.InDelta(float64(testcase.expect[1]), float64(ndc_y), epsilon_y)
			})
		}
	})
}
