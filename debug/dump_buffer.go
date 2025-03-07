package debug

import (
	"unsafe"

	"github.com/go-gl-legacy/gl"
)

func DumpBuffer(buf gl.Buffer) []float32 {
	// TODO(tmckee): assert on render thread? somehow?

	// save old ARRAY_BUFFER mapping; revert on return
	var oldBinding [1]int32
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, oldBinding[:])
	// If there was no buffer bound to ARRAY_BUFFER, the get returns 0. We'll
	// reset to this 'unbound' state so as not to leak the 'buf' binding.
	defer gl.Buffer(oldBinding[0]).Bind(gl.ARRAY_BUFFER)

	// bind the buffer to ARRAY_BUFFER
	buf.Bind(gl.ARRAY_BUFFER)

	// map what's bound to ARRAY_BUFFER; revert on return
	data := gl.MapBuffer(gl.ARRAY_BUFFER, gl.READ_WRITE)
	defer gl.UnmapBuffer(gl.ARRAY_BUFFER)

	// find how much data is in the buffer
	dataByteSize := gl.GetBufferParameteriv(gl.ARRAY_BUFFER, gl.BUFFER_SIZE)
	numFloats := dataByteSize / int32(unsafe.Sizeof(float32(0)))

	// memcpy and return the data
	var asSlice []float32 = unsafe.Slice((*float32)(data), numFloats)
	result := make([]float32, len(asSlice))
	copy(result, asSlice)

	return result
}
