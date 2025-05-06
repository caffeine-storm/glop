package rendertest

import (
	"unsafe"

	"github.com/go-gl-legacy/gl"
)

func GivenABufferWithData(data []float32) gl.Buffer {
	result := gl.GenBuffer()
	result.Bind(gl.ARRAY_BUFFER)

	floatSize := int(unsafe.Sizeof(float32(0)))
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(data), data, gl.STATIC_DRAW)

	return result
}
