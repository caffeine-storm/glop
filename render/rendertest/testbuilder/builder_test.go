package testbuilder_test

import (
	"image/color"
	"log"
	"strings"
	"testing"

	"github.com/go-gl-legacy/gl"
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
			Convey("but detached from a 'real' testing.T so that we don't fail the whole test", &testing.T{}, func() {
				testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
					So(1, ShouldNotEqual, 1)
				})
				shouldBeFalse = true
			})
		})

		require.False(t, shouldBeFalse)
		assert.NotContains(t, strings.Join(output, "\n"), "___FAILURE_HALT___")
	})
}

func TestBuilderFluentApi(t *testing.T) {
	t.Run("default builder runs on render thread", func(t *testing.T) {
		testbuilder.New().Run(func() {
			rendertest.AssertOnRenderThread(t)
		})
	})

	t.Run("can run off of render thread", func(t *testing.T) {
		testbuilder.New().WithQueue().Run(func(queue render.RenderQueueInterface) {
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
			testbuilder.New().WithSize(64, 128).Run(func() {
				rendertest.AssertOnRenderThread(t)
				_, _, dx, dy := debug.GetViewport()
				assert.Equal(dx, uint32(64))
				assert.Equal(dy, uint32(128))
			})
		})

		t.Run("and get the queue after", func(t *testing.T) {
			assert := assert.New(t)
			testbuilder.New().WithSize(64, 128).WithQueue().Run(func(queue render.RenderQueueInterface) {
				rendertest.AssertOffRenderThread(t)
				queue.Queue(func(render.RenderQueueState) {
					_, _, dx, dy := debug.GetViewport()
					assert.Equal(dx, uint32(64))
					assert.Equal(dy, uint32(128))
				})
				queue.Purge()
			})
		})

		t.Run("and get the queue first", func(t *testing.T) {
			assert := assert.New(t)
			testbuilder.New().WithQueue().WithSize(64, 128).Run(func(queue render.RenderQueueInterface) {
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
			testbuilder.New().WithExpectation(c, "red").Run(func() {
				rendertest.DrawRectNdc(-1, -1, 1, 1)
			})

			purpleBackground := color.RGBA{
				R: 225, G: 0, B: 255, A: 255,
			}
			testbuilder.New().WithExpectation(c, "red-on-purple", purpleBackground).Run(func() {
				rendertest.DrawRectNdc(-0.5, -0.5, 0.5, 0.5)
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
