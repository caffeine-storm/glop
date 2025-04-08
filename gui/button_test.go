package gui_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gin/gintesting"
	"github.com/runningwild/glop/gui"
	. "github.com/smartystreets/goconvey/convey"
)

func GivenAButton(fn func(gui.EventHandlingContext, int64)) *gui.Button {
	testingFont := "glop.font"
	testingLabel := "button-label"
	testingWidth := 42
	testingLuminance := float64(1.0)
	return gui.MakeButton(testingFont, testingLabel, testingWidth, testingLuminance, testingLuminance, testingLuminance, testingLuminance, fn)
}

func ClickAButton(btn *gui.Button) {
	input := gin.Make()

	eventGroup := gui.EventGroup{
		EventGroup:                 gintesting.ClickEventGroup(input),
		DispatchedToFocussedWidget: true,
	}

	btn.DoRespond(nil, eventGroup)
}

func TestButton(t *testing.T) {
	Convey("Button Widgets", t, func() {
		Convey("can be clicked", func() {
			clicks := []int64{}

			onClick := func(ctx gui.EventHandlingContext, delta int64) {
				clicks = append(clicks, delta)
			}
			btn := GivenAButton(onClick)

			So(len(clicks), ShouldEqual, 0)

			ClickAButton(btn)

			So(len(clicks), ShouldEqual, 1)
		})
	})
}
