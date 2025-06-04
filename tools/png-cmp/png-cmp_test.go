package main_test

import (
	"image"
	"testing"

	"github.com/runningwild/glop/tools/png-cmp"
)

func TestWeUseNrgba(t *testing.T) {
	fourByFour := image.Rect(0, 0, 4, 4)
	lhs := image.NewNRGBA(fourByFour)
	rhs := image.NewNRGBA(fourByFour)

	main.ImageCompare(lhs, rhs)
}
