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
	return gl.GLenum(gl.GetInteger(gl.MATRIX_MODE))
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

func GetActiveTextureUnit() gl.GLenum {
	return gl.GLenum(gl.TEXTURE0 - gl.GetInteger(gl.ACTIVE_TEXTURE))
}

type GlState struct {
	MatrixMode       string
	ModelViewMatrix  []float64
	ProjectionMatrix []float64
	ColorMatrix      []float64
	TextureMatrix    []float64

	Colours  map[string]string
	Bindings map[string]int
	FlagSet  map[string]gl.GLenum
}

func getColour(colourName gl.GLenum) (byte, byte, byte, byte) {
	r, g, b, a := gl.GetDouble4(colourName)
	return byte(r * 256), byte(g * 256), byte(b * 256), byte(a * 256)
}

func colourString(r, g, b, a byte) string {
	return fmt.Sprintf("(0x%x,0x%x,0x%x,0x%x)", r, g, b, a)
}

func GetFlagSet() map[string]gl.GLenum {
	intbools := map[bool]gl.GLenum{
		true:  gl.TRUE,
		false: gl.FALSE,
	}
	ret := map[string]gl.GLenum{}

	ret["ACTIVE_TEXTURE_UNIT"] = GetActiveTextureUnit()
	ret["GL_BLEND"] = intbools[gl.IsEnabled(gl.BLEND)]
	ret["GL_CLIP_PLANE0"] = intbools[gl.IsEnabled(gl.CLIP_PLANE0)]
	ret["GL_CLIP_PLANE1"] = intbools[gl.IsEnabled(gl.CLIP_PLANE1)]
	ret["GL_CLIP_PLANE2"] = intbools[gl.IsEnabled(gl.CLIP_PLANE2)]
	ret["GL_CLIP_PLANE3"] = intbools[gl.IsEnabled(gl.CLIP_PLANE3)]

	ret["GL_CULL_FACE"] = intbools[gl.IsEnabled(gl.CULL_FACE)]
	ret["GL_DEPTH_TEST"] = intbools[gl.IsEnabled(gl.DEPTH_TEST)]
	ret["GL_DITHER"] = intbools[gl.IsEnabled(gl.DITHER)]
	ret["GL_INDEX_ARRAY"] = intbools[gl.IsEnabled(gl.INDEX_ARRAY)]
	ret["GL_NORMAL_ARRAY"] = intbools[gl.IsEnabled(gl.NORMAL_ARRAY)]
	ret["GL_NORMALIZE"] = intbools[gl.IsEnabled(gl.NORMALIZE)]
	ret["GL_SCISSOR_TEST"] = intbools[gl.IsEnabled(gl.SCISSOR_TEST)]
	ret["GL_STENCIL_TEST"] = intbools[gl.IsEnabled(gl.STENCIL_TEST)]
	ret["GL_TEXTURE_2D"] = intbools[gl.IsEnabled(gl.TEXTURE_2D)]
	ret["GL_TEXTURE_3D"] = intbools[gl.IsEnabled(gl.TEXTURE_3D)]
	ret["GL_TEXTURE_COORD_ARRAY"] = intbools[gl.IsEnabled(gl.TEXTURE_COORD_ARRAY)]
	ret["GL_VERTEX_ARRAY"] = intbools[gl.IsEnabled(gl.VERTEX_ARRAY)]
	ret["GL_VERTEX_PROGRAM_POINT_SIZE"] = intbools[gl.IsEnabled(gl.VERTEX_PROGRAM_POINT_SIZE)]
	ret["GL_VERTEX_PROGRAM_TWO_SIDE"] = intbools[gl.IsEnabled(gl.VERTEX_PROGRAM_TWO_SIDE)]

	return ret
}

func GetBindingsSet() map[string]int {
	ret := map[string]int{}

	ret["ARRAY_BUFFER_BINDING"] = gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
	ret["ELEMENT_ARRAY_BUFFER_BINDING"] = gl.GetInteger(gl.ELEMENT_ARRAY_BUFFER_BINDING)
	ret["PIXEL_PACK_BUFFER_BINDING"] = gl.GetInteger(gl.PIXEL_PACK_BUFFER_BINDING)
	ret["PIXEL_UNPACK_BUFFER_BINDING"] = gl.GetInteger(gl.PIXEL_UNPACK_BUFFER_BINDING)
	ret["TEXTURE_BINDING_2D"] = gl.GetInteger(gl.TEXTURE_BINDING_2D)
	ret["TEXTURE_COORD_ARRAY_BUFFER_BINDING"] = gl.GetInteger(gl.TEXTURE_COORD_ARRAY_BUFFER_BINDING)
	ret["VERTEX_ARRAY_BUFFER_BINDING"] = gl.GetInteger(gl.VERTEX_ARRAY_BUFFER_BINDING)

	return ret
}

func GetColours() map[string]string {
	ret := map[string]string{}

	ret["gl.CURRENT_COLOR"] = colourString(getColour(gl.CURRENT_COLOR))
	ret["gl.COLOR_CLEAR_VALUE"] = colourString(getColour(gl.COLOR_CLEAR_VALUE))

	return ret
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

		Bindings: GetBindingsSet(),
		Colours:  GetColours(),
		FlagSet:  GetFlagSet(),
	}
}
