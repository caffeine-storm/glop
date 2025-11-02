package testbuilder_test

import (
	"fmt"
	"image/color"
	"log"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/caffeine-storm/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/gloptest"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunningOnRenderThread(t *testing.T) {
	testbuilder.New().Run(func() {
		versionString := gl.GetString(gl.VERSION)
		log.Printf("versionString: %q\n", versionString)

		if versionString == "" {
			t.Error("gl.GetString(gl.VERSION) must not return the empty string once OpenGL is initialized")
		}
	})
}

func TestConveyHalting(t *testing.T) {
	t.Run("if the Convey framework wants to halt, we don't confuse things", func(t *testing.T) {
		shouldBeFalse := false
		output := gloptest.CollectOutput(func() {
			Convey("but detached from a 'real' testing.T so that we don't fail the whole test", &testing.T{}, func(c C) {
				testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
					c.So(1, ShouldNotEqual, 1)
				})
				shouldBeFalse = true
			})
		})

		require.False(t, shouldBeFalse)
		assert.NotContains(t, strings.Join(output, "\n"), "___FAILURE_HALT___")
		assert.NotContains(t, strings.Join(output, "\n"), "Convey operation made without context on goroutine stack")
	})
}

func TestBuilderFluentApi(t *testing.T) {
	t.Run("default builder runs on render thread", func(t *testing.T) {
		testbuilder.Run(func() {
			rendertest.AssertOnRenderThread(t)
		})
	})

	t.Run("can run off of render thread", func(t *testing.T) {
		testbuilder.Run(func(queue render.RenderQueueInterface) {
			rendertest.AssertOffRenderThread(t)

			queue.Queue(func(render.RenderQueueState) {
				rendertest.AssertOnRenderThread(t)
			})
			queue.Purge()
		})
	})

	t.Run("can specify dimensions", func(t *testing.T) {
		t.Run("with literal sizes", func(t *testing.T) {
			assert := assert.New(t)
			testbuilder.WithSize(64, 128, func() {
				rendertest.AssertOnRenderThread(t)
				_, _, dx, dy := debug.GetViewport()
				assert.Equal(dx, uint32(64))
				assert.Equal(dy, uint32(128))
			})
		})

		t.Run("and get the queue", func(t *testing.T) {
			assert := assert.New(t)
			testbuilder.WithSize(64, 128, func(queue render.RenderQueueInterface) {
				rendertest.AssertOffRenderThread(t)
				queue.Queue(func(render.RenderQueueState) {
					_, _, dx, dy := debug.GetViewport()
					assert.Equal(dx, uint32(64))
					assert.Equal(dy, uint32(128))
				})
				queue.Purge()
			})
		})
	})

	t.Run("easy expectations checking", func(t *testing.T) {
		Convey("expectationy things need convey", t, func(c C) {
			testbuilder.WithExpectation(c, "red", func() {
				rendertest.DrawRectNdc(-1, -1, 1, 1)
			})

			testbuilder.WithExpectation(c, "red", func(st render.RenderQueueState) {
				rendertest.DrawRectNdc(-1, -1, 1, 1)
			})

			pinkBackground := color.RGBA{
				R: 255, G: 0, B: 255, A: 255,
			}
			testbuilder.WithExpectation(c, "red-on-pink", pinkBackground, func() {
				render.WithBlankScreen(1.0, 0.0, 1.0, 1.0, func() {
					rendertest.DrawRectNdc(-0.5, -0.5, 0.5, 0.5)
				})
			})

			testbuilder.WithSizeAndExpectation(64, 64, c, "red", func(queue render.RenderQueueInterface) {
				queue.Queue(func(render.RenderQueueState) {
					rendertest.DrawRectNdc(-1, -1, 1, 1)
				})
				queue.Purge()
			})
		})
	})
}

func TestGlStateLeakage(t *testing.T) {
	t.Run("testbuilder should complain upon leakage", func(t *testing.T) {
		assert.Panics(t, func() {
			testbuilder.New().Run(func() {
				// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
				buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
				buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
			})
		})
	})
}

// Tests ought to be able to fail without being annoying.
// Tests should fail if they call panic() or testing.T.Fatalf().
// These calls could happen on or off of a render thread.
func TestFailureHandling(t *testing.T) {
	t.Run("calling testing.T.Fatalf off of the render thread", func(t *testing.T) {
		timeout, operr := gloptest.RunWithDeadline(20*time.Millisecond, func() {
			fakeTesting := &testing.T{}
			testbuilder.Run(func(queue render.RenderQueueInterface) {
				// Note: we _don't_ use 'queue' because we want to call Fatalf off of the
				// render thread for this test.
				fakeTesting.Fatalf("simulated test-failure")
			})
		})

		if timeout != nil {
			t.Fatalf("the test-under-test should have failed before the timeout")
		}
		if operr == nil {
			panic(fmt.Errorf("the test-under-test should have failed"))
		}
	})
}

func clearUnsetValues(mp map[string]int) {
	todelete := map[string]bool{}
	for key, val := range mp {
		if val == 0 {
			todelete[key] = true
		}
	}

	for key := range todelete {
		delete(mp, key)
	}
}

func TestCrossTalkPrevention(t *testing.T) {
	// Before and after a test, there should be certain invariants otherwise
	// tests are susceptible to cross-talk.
	var initialState map[string]int
	testbuilder.New().Run(func() {
		initialState = debug.GetBindingsSet()
	})

	t.Run("initial state", func(t *testing.T) {
		for bindingName, boundVal := range initialState {
			if boundVal != 0 {
				t.Logf("found initially bound state: name: %q, val: %d\n", bindingName, boundVal)
				t.Fail()
			}
		}
	})
	clearUnsetValues(initialState)
	assert.Empty(t, initialState, "tests should be allowed to expect a clean initial state")

	var taintedState map[string]int
	t.Run("testbuilder helpers send errors if state-change is leaked", func(t *testing.T) {
		var expectedError error = nil
		assert.Panics(t, func() {
			testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
				queue.AddErrorCallback(func(_ render.RenderQueueInterface, e error) {
					expectedError = e
				})
				queue.Queue(func(render.RenderQueueState) {
					// An example of tainted state is leaving ELEMENT_ARRAY_BUFFER bound
					buf := rendertest.GivenABufferWithData([]float32{0, 1, 2, 0, 2, 3})
					buf.Bind(gl.ELEMENT_ARRAY_BUFFER)
					taintedState = debug.GetBindingsSet()
				})
				queue.Purge()

				// We haven't run the "cleanup" phase of this testcase so state leakage
				// should not be checked yet.
				assert.Nil(t, expectedError)
			})
		})
		assert.NotNil(t, expectedError)

		clearUnsetValues(taintedState)
		assert.NotEqual(t, slices.Sorted(maps.Keys(initialState)), slices.Sorted(maps.Keys(taintedState)))
	})
}
