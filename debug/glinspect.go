package debug

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl-legacy/gl"
)

// Returns a string describing the current matrix mode.
func GetMatrixMode() string {
	buffer := []int32{0}
	gl.GetIntegerv(gl.MATRIX_MODE, buffer)
	mappedSymbols := map[int32]string{
		gl.MODELVIEW:  "model-view",
		gl.PROJECTION: "projection",
		gl.TEXTURE:    "texture",
		gl.COLOR:      "color",
	}

	result, ok := mappedSymbols[buffer[0]]
	if !ok {
		panic(fmt.Errorf("couldn't gl.GetInteger(gl.MATRIX_MODE): buffer[0]: %v", buffer[0]))
	}

	return result
}

// Returns (x-pos, y-pos, width, height) of the current viewport.
func GetViewport() (int32, int32, uint32, uint32) {
	buffer := []int32{0, 0, 0, 0}
	gl.GetIntegerv(gl.VIEWPORT, buffer)
	return buffer[0], buffer[1], uint32(buffer[2]), uint32(buffer[3])
}

// Returns the current (near, far) values as set from the last glDepthRange
// call.
func GetDepthRange() (float64, float64) {
	buffer := []float64{0, 0}
	gl.GetDoublev(gl.DEPTH_RANGE, buffer)
	return buffer[0], buffer[1]
}

func LogAndClearGlErrors(logger *log.Logger) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("couldn't call runtime.Caller(1)")
	}

	parts := strings.SplitAfter(file, "glop")
	file = parts[len(parts)-1][1:]

	for {
		glErr := gl.GetError()
		if glErr == gl.NO_ERROR {
			return
		}

		// TODO(tmckee): include the line number where we were called.
		logger.Printf("GlError(%s:%d): 0x%04x\n", file, line, glErr)
	}
}
