package debug

import (
	"fmt"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glew"
)

func mappedSymbols() map[gl.GLenum]string {
	return map[gl.GLenum]string{
		gl.MODELVIEW:  "model-view",
		gl.PROJECTION: "projection",
		gl.TEXTURE:    "texture",
		gl.COLOR:      "color",
	}
}

// Returns an enum denoting the current matrix mode.
func GetMatrixMode() gl.GLenum {
	buffer := []int32{0}
	gl.GetIntegerv(gl.MATRIX_MODE, buffer)
	return gl.GLenum(buffer[0])
}

func getMatrix(paramName gl.GLenum) []float64 {
	ret := [16]float64{}
	gl.GetDoublev(paramName, ret[:])
	return ret[:]
}

func GetColorMatrix() []float64 {
	// The 'Color Matrix' is part of an optional piece of the core OpenGL API.
	// Nobody is required to implement it and they'll return an error if we try
	// to query for it blindly. If the imaging extensions aren't supported, we'll
	// return nil instead.
	if !glew.GL_ARB_imaging {
		return nil
	}
	return getMatrix(gl.COLOR_MATRIX)
}

func GetModelViewMatrix() []float64 {
	return getMatrix(gl.MODELVIEW_MATRIX)
}

func GetProjectionMatrix() []float64 {
	return getMatrix(gl.PROJECTION_MATRIX)
}

func GetTextureMatrix() []float64 {
	return getMatrix(gl.TEXTURE_MATRIX)
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

type GlState struct {
	MatrixMode       string
	ModelViewMatrix  []float64
	ProjectionMatrix []float64
	ColorMatrix      []float64
	TextureMatrix    []float64
}

func (st *GlState) String() string {
	return fmt.Sprintf("%+v", *st)
}

// Returns a high-level description of what the current GL state is.
func GetGlState() *GlState {
	return &GlState{
		MatrixMode:       mappedSymbols()[GetMatrixMode()],
		ModelViewMatrix:  GetModelViewMatrix(),
		ProjectionMatrix: GetProjectionMatrix(),
		ColorMatrix:      GetColorMatrix(),
		TextureMatrix:    GetTextureMatrix(),
	}
}
