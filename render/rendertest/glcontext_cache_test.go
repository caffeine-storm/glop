package rendertest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runningwild/glop/gloptest"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func thisFunctionDereferencesNil() {
	var nilPointer *string = nil

	_ = len(*nilPointer)

	panic(fmt.Errorf("should not get here"))
}

func TestFailureMessages(t *testing.T) {
	t.Run("for nil pointer dereferences, it includes the func that dereferenced the nil", func(t *testing.T) {
		testoutput := make(chan []string)

		go func() {
			testoutput <- gloptest.CollectOutput(func() {
				canaryTestInstance := &testing.T{}
				Convey("over bogus testing.T to trick Convey", canaryTestInstance, func() {
					rendertest.RunTestWithCachedContext(64, 64, func(system.System, system.NativeWindowHandle, render.RenderQueueInterface) {
						thisFunctionDereferencesNil()
						So("wait", ShouldContainSubstring, "wut?")
					})
				})
			})
		}()

		result := strings.Join(<-testoutput, "\n")
		assert.Contains(t, result, "thisFunctionDereferencesNil")
	})
}
