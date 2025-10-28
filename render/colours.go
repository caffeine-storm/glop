package render

import (
	"fmt"
	"image/color"

	"github.com/go-gl-legacy/gl"
)

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
		panic(fmt.Errorf("non-normalized float: %v is not in range [0.0, 1.0]", f))
	}
	ret := int(f * 256)
	if ret >= 256 {
		ret = 255
	}
	return uint8(ret)
}

func GetCurrentBackgroundColor() color.NRGBA {
	oldClear := getCurrentBackground()
	return color.NRGBA{
		R: normColourToByte(oldClear[0]),
		G: normColourToByte(oldClear[1]),
		B: normColourToByte(oldClear[2]),
		A: normColourToByte(oldClear[3]),
	}
}

func GetCurrentForegroundColour() color.NRGBA {
	oldFg := getCurrentForeground()
	return color.NRGBA{
		R: normColourToByte(oldFg[0]),
		G: normColourToByte(oldFg[1]),
		B: normColourToByte(oldFg[2]),
		A: normColourToByte(oldFg[3]),
	}
}
