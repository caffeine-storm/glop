package rendertest

import (
	"image/color"

	"github.com/go-gl-legacy/gl"
)

func WithClearColour(r, g, b, a gl.GLclampf, fn func()) {
	oldClear := getCurrentBackground()

	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	defer func() {
		gl.ClearColor(
			gl.GLclampf(oldClear[0]),
			gl.GLclampf(oldClear[1]),
			gl.GLclampf(oldClear[2]),
			gl.GLclampf(oldClear[3]))
	}()
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

func normColorToByte(f float32) uint8 {
	if f < 0 || f > 1.0 {
		panic("non-normalized float!")
	}
	return uint8(f * 255)
}

func GetCurrentBackgroundColor() color.RGBA {
	oldClear := getCurrentBackground()
	return color.RGBA{
		R: normColorToByte(oldClear[0]),
		G: normColorToByte(oldClear[1]),
		B: normColorToByte(oldClear[2]),
		A: normColorToByte(oldClear[3]),
	}
}

func GetCurrentForegroundColour() color.RGBA {
	oldFg := getCurrentForeground()
	return color.RGBA{
		R: normColorToByte(oldFg[0]),
		G: normColorToByte(oldFg[1]),
		B: normColorToByte(oldFg[2]),
		A: normColorToByte(oldFg[3]),
	}
}
