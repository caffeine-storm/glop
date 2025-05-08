package rendertest_test

import (
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	t.Run("rendertest.GivenABufferWithData", func(t *testing.T) {
		t.Run("shouldn't clobber gl.ARRAY_BUFFER", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.DeprecatedWithGl(func() {
				oldbuf := gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
				buf := rendertest.GivenABufferWithData([]float32{
					0, 1, 2, 3, 4, 5,
				})
				assert.NotEqual(oldbuf, buf)

				afterbuf := gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
				assert.Equal(oldbuf, afterbuf)
			})
		})
		t.Run("shouldn't clobber gl.ELEMENT_ARRAY_BUFFER", func(t *testing.T) {
			assert := assert.New(t)
			rendertest.DeprecatedWithGl(func() {
				oldbuf := gl.GetInteger(gl.ELEMENT_ARRAY_BUFFER_BINDING)
				buf := rendertest.GivenABufferWithData([]float32{
					0, 1, 2, 3, 4, 5,
				})
				assert.NotEqual(oldbuf, buf)

				afterbuf := gl.GetInteger(gl.ELEMENT_ARRAY_BUFFER_BINDING)
				assert.Equal(oldbuf, afterbuf)
			})
		})
	})
}
