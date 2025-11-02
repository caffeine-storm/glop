package rendertest

import (
	"unsafe"

	"github.com/caffeine-storm/gl"
)

func GivenABufferWithData(data []float32) gl.Buffer {
	oldbuf := gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
	defer gl.Buffer(oldbuf).Bind(gl.ARRAY_BUFFER)

	result := gl.GenBuffer()
	result.Bind(gl.ARRAY_BUFFER)

	floatSize := int(unsafe.Sizeof(float32(0)))
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(data), data, gl.STATIC_DRAW)

	return result
}
