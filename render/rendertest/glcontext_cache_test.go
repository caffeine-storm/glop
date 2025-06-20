package rendertest_test

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
)

func thisFunctionDereferencesNil() {
	var nilPointer *string = nil

	_ = len(*nilPointer)

	panic(fmt.Errorf("should not get here"))
}

func TestFailureMessages(t *testing.T) {
	Convey("for null pointer dereferences", t, func() {
		Convey("include stacktrace with line and file info", func() {
			var testoutput []byte
			stackContents := make(chan []byte)
			go func() {
				canaryTestInstance := &testing.T{}
				Convey("over bogus testing.T to trick Convey", canaryTestInstance, func() {
					defer func() {
						e := recover()
						if e == nil {
							panic("umm... wut?")
						}
						stackContents <- debug.Stack()
					}()

					rendertest.RunTestWithCachedContext(64, 64, func(system.System, system.NativeWindowHandle, render.RenderQueueInterface) {
						thisFunctionDereferencesNil()
					})
				})
			}()
			testoutput = <-stackContents

			So(string(testoutput), ShouldContainSubstring, "thisFunctionDereferencesNil")
		})
	})
}
