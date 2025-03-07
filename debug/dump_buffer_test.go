package debug_test

import (
	"testing"
	"unsafe"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func GivenSomeFloats() []float32 {
	return []float32{
		// Pretend we're drawing a quad with two tex-coords across the entire
		// screen.
		-1, -1, 0, 0, 0,
		-1, +1, 0, 0, 1,
		+1, +1, 0, 1, 1,
		+1, -1, 0, 1, 0,
	}
}

func GivenABufferWithData(data []float32) gl.Buffer {
	result := gl.GenBuffer()
	result.Bind(gl.ARRAY_BUFFER)

	floatSize := int(unsafe.Sizeof(float32(0)))
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(data), data, gl.STATIC_DRAW)

	return result
}

func TestDumpBuffer(t *testing.T) {
	data := GivenSomeFloats()

	var dumpResult []float32
	rendertest.WithGl(func() {
		buf := GivenABufferWithData(data)
		dumpResult = debug.DumpBuffer(buf)
	})

	assert.Equal(t, data, dumpResult)
}
