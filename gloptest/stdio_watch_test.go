package gloptest_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/caffeine-storm/glop/gloptest"
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
			didrun := false
			var collectedOutput []string
			func() {
				collectedOutput = gloptest.CollectOutput(func() {
					So(func() {
						didrun = true
						fmt.Printf("1: some output")
						panic("for testing")
					}, ShouldPanic)
					fmt.Printf("2: some more output")
				})
			}()
			So(didrun, ShouldBeTrue)

			singleString := strings.Join(collectedOutput, "\n")
			So(singleString, ShouldContainSubstring, "1: some output")
			So(singleString, ShouldContainSubstring, "2: some more output")
		})
	})
}
