package rendertest

import (
	"unsafe"

	"github.com/go-gl-legacy/gl"
)

func GivenABufferWithData(data []float32) gl.Buffer {
	oldbuf := [1]int32{}
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, oldbuf[:])
	defer gl.Buffer(oldbuf[0]).Bind(gl.ARRAY_BUFFER)

	result := gl.GenBuffer()
	result.Bind(gl.ARRAY_BUFFER)

	floatSize := int(unsafe.Sizeof(float32(0)))
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(data), data, gl.STATIC_DRAW)

	return result
}
