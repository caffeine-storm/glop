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

func GetActiveTextureUnit() gl.GLenum {
	val := []int32{0}
	gl.GetIntegerv(gl.ACTIVE_TEXTURE, val[:])

	return gl.GLenum(gl.TEXTURE0 - val[0])
}

type GlState struct {
	MatrixMode       string
	ModelViewMatrix  []float64
	ProjectionMatrix []float64
	ColorMatrix      []float64
	TextureMatrix    []float64
	Bindings         map[string]int32
	FlagSet          map[string]gl.GLenum
}

func GetFlagSet() map[string]gl.GLenum {
	intbools := map[bool]gl.GLenum{
		true:  gl.TRUE,
		false: gl.FALSE,
	}
	ret := map[string]gl.GLenum{}

	ret["ACTIVE_TEXTURE"] = GetActiveTextureUnit()
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

func GetBindingsSet() map[string]int32 {
	ret := map[string]int32{}

	getbinding := func(name gl.GLenum) int32 {
		ret := [1]int32{}
		gl.GetIntegerv(name, ret[:])
		return ret[0]
	}

	ret["ARRAY_BUFFER_BINDING"] = getbinding(gl.ARRAY_BUFFER_BINDING)
	ret["ELEMENT_ARRAY_BUFFER_BINDING"] = getbinding(gl.ELEMENT_ARRAY_BUFFER_BINDING)
	ret["PIXEL_PACK_BUFFER_BINDING"] = getbinding(gl.PIXEL_PACK_BUFFER_BINDING)
	ret["PIXEL_UNPACK_BUFFER_BINDING"] = getbinding(gl.PIXEL_UNPACK_BUFFER_BINDING)
	ret["TEXTURE_BINDING_2D"] = getbinding(gl.TEXTURE_BINDING_2D)
	ret["TEXTURE_COORD_ARRAY_BUFFER_BINDING"] = getbinding(gl.TEXTURE_COORD_ARRAY_BUFFER_BINDING)
	ret["VERTEX_ARRAY_BUFFER_BINDING"] = getbinding(gl.VERTEX_ARRAY_BUFFER_BINDING)

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
		FlagSet:  GetFlagSet(),
	}
}
