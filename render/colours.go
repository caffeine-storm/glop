package render

import (
	"fmt"
	"image/color"
	"slices"

	"github.com/caffeine-storm/gl"
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

func assertNormalized[T ~float32](values ...T) {
	if slices.ContainsFunc(values, func(v T) bool {
		return v < 0.0 || v > 1.0
	}) {
		panic(fmt.Errorf("each float needs to be in the range [0, 1] but got %v", values))
	}
}

func NormalizedColourToByte(f float32) uint8 {
	assertNormalized(f)

	ret := int(f * 256)
	if ret >= 256 {
		ret = 255
	}
	return uint8(ret)
}

func ByteToNormalizedColour(b uint8) float32 {
	return float32(b) / 255
}

func GetCurrentBackgroundColor() color.NRGBA {
	oldClear := getCurrentBackground()
	return color.NRGBA{
		R: NormalizedColourToByte(oldClear[0]),
		G: NormalizedColourToByte(oldClear[1]),
		B: NormalizedColourToByte(oldClear[2]),
		A: NormalizedColourToByte(oldClear[3]),
	}
}

func GetCurrentForegroundColour() color.NRGBA {
	oldFg := getCurrentForeground()
	return color.NRGBA{
		R: NormalizedColourToByte(oldFg[0]),
		G: NormalizedColourToByte(oldFg[1]),
		B: NormalizedColourToByte(oldFg[2]),
		A: NormalizedColourToByte(oldFg[3]),
	}
}
