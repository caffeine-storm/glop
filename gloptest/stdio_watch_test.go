package gloptest_test

import (
	"testing"

	"github.com/runningwild/glop/gloptest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStdioWatchWithConvey(t *testing.T) {
	Convey("gloptest.StdioWatch can take a func that calls Convey.So", t, func() {
		So(1, ShouldEqual, 1)

		gloptest.CollectOutput(func() {
			So(2, ShouldEqual, 2)
		})
	})
}
