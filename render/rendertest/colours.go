package rendertest

import (
	"image/color"

	"github.com/go-gl-legacy/gl"
)

func WithClearColour(r, g, b, a gl.GLclampf, fn func()) {
	gl.PushAttrib(gl.COLOR_BUFFER_BIT)
	defer gl.PopAttrib()

	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	fn()
}

func getCurrentBackground() [4]float32 {
	oldClear := [4]float32{0, 0, 0, 0}
	gl.GetFloatv(gl.COLOR_CLEAR_VALUE, oldClear[:])
	return oldClear
}

func getCurrentForeground() [4]float32 {
	ret := [4]float32{}
	gl.GetFloatv(gl.CURRENT_COLOR, ret[:])
	return ret
}

func normColourToByte(f float32) uint8 {
	if f < 0 || f > 1.0 {
		panic("non-normalized float!")
	}
	return uint8(f * 255)
}

func GetCurrentBackgroundColor() color.RGBA {
	oldClear := getCurrentBackground()
	return color.RGBA{
		R: normColourToByte(oldClear[0]),
		G: normColourToByte(oldClear[1]),
		B: normColourToByte(oldClear[2]),
		A: normColourToByte(oldClear[3]),
	}
}

func GetCurrentForegroundColour() color.RGBA {
	oldFg := getCurrentForeground()
	return color.RGBA{
		R: normColourToByte(oldFg[0]),
		G: normColourToByte(oldFg[1]),
		B: normColourToByte(oldFg[2]),
		A: normColourToByte(oldFg[3]),
	}
}
