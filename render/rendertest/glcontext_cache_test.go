package rendertest_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/runningwild/glop/gloptest"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		assert.Contains(t, result, "thisFunctionDereferencesNil", "should include function attribution")

		// Check for pattern '<filename>:<linenumber>\n'
		filename := "glcontext_cache_canary_test.go"
		_, result, ok := strings.Cut(result, filename)
		require.True(t, ok, "should include file attribution")

		_, result, ok = strings.Cut(result, ":")
		require.True(t, ok, "needs a ':' between file name and line")

		lineNumber, _, ok := strings.Cut(result, "\n")
		require.True(t, ok, "need a newline after the line number")

		val, err := strconv.Atoi(lineNumber)
		assert.Positive(t, val)
		assert.NoError(t, err, "should include line number attribution")
	})
}
