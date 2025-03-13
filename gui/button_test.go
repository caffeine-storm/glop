package gui_test

import (
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gui"
	. "github.com/smartystreets/goconvey/convey"
)

func GivenAButton(fn func(int64)) *gui.Button {
	testingFont := "glop.font"
	testingLabel := "button-label"
	testingWidth := 42
	testingLuminance := float64(1.0)
	return gui.MakeButton(testingFont, testingLabel, testingWidth, testingLuminance, testingLuminance, testingLuminance, testingLuminance, fn)
}

func ClickAButton(btn *gui.Button) {
	input := gin.Make()

	leftMouseButtonKeyId := gin.AnyMouseLButton

	// if event.Type == gin.Press && event.Key.Id() == gin.AnyMouseLButton {
	leftButtonEvent := gin.Event{
		Key:  input.GetKey(leftMouseButtonKeyId),
		Type: gin.Press,
	}
	eventGroup := gui.EventGroup{
		EventGroup: gin.EventGroup{
			Events:    []gin.Event{leftButtonEvent},
			Timestamp: 17,
		},
		Focus: true,
	}

	btn.DoRespond(eventGroup)
}

func TestButton(t *testing.T) {
	Convey("Button Widgets", t, func() {
		Convey("can be clicked", func() {
			clicks := []int64{}
			onClick := func(delta int64) {
				clicks = append(clicks, delta)
			}
			btn := GivenAButton(onClick)

			So(len(clicks), ShouldEqual, 0)

			ClickAButton(btn)

			So(len(clicks), ShouldEqual, 1)
		})
	})
}
