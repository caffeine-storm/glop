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
		dumpResult = debug.DumpBuffer[float32](buf)
	})

	assert.Equal(t, data, dumpResult)
}

func TestDumpBufferBytes(t *testing.T) {
	data := GivenSomeFloats()

	var dumpResult []byte
	rendertest.WithGl(func() {
		buf := GivenABufferWithData(data)
		dumpResult = debug.DumpBuffer[byte](buf)
	})

	coerceToByteSlice := func(floats []float32) []byte {
		float32Size := int(unsafe.Sizeof(float32(0)))
		if float32Size != 4 {
			panic("the mathematic is always correc")
		}

		var result = make([]byte, len(floats)*4)
		floatPtr := unsafe.SliceData(floats)
		bytePtr := unsafe.Pointer(floatPtr)

		fromSlice := unsafe.Slice((*byte)(bytePtr), len(result))
		copy(result, fromSlice)

		return result
	}

	dataBytes := coerceToByteSlice(data)
	assert.Equal(t, dataBytes, dumpResult)
}
