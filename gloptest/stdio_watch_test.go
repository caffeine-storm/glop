package gloptest_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/runningwild/glop/gloptest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDoubleClosePipeWrite(t *testing.T) {
	_, write, err := os.Pipe()
	if err != nil {
		panic(fmt.Errorf("couldn't os.Pipe(): %w", err))
	}

	write.Close()
	write.Close()
}

func TestCollectOutputWithConvey(t *testing.T) {
	Convey("gloptest.CollectOutput", t, func() {
		Convey("can take a func that calls Convey.So", func() {
			So(1, ShouldEqual, 1)

			gloptest.CollectOutput(func() {
				So(2, ShouldEqual, 2)
			})
		})
		Convey("can take a func that panics", func() {
			testpass := false
			didrun := false
			func() {
				defer func() {
					if v := recover(); v != nil {
						testpass = true
					}
				}()
				gloptest.CollectOutput(func() {
					So(3, ShouldEqual, 3)
					didrun = true
					panic("for testing")
				})
			}()
			So(testpass, ShouldBeTrue)
			So(didrun, ShouldBeTrue)
		})
	})
}
