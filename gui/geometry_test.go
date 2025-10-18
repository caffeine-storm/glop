package gui_test

import (
	"testing"

	"github.com/runningwild/glop/gui"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestRegion(t *testing.T) {
	t.Run("clipping", func(t *testing.T) {
		Convey("region clipping", t, func(c C) {
			testbuilder.New().WithExpectation(c, "red-with-border").Run(func() {
				// Set a clipping region to block any drawing outside of a square in the
				// middle.
				r := gui.MakeRegion(4, 4, 56, 56)

				r.PushClipPlanes()
				defer r.PopClipPlanes()

				// Draw a red square across the 'whole' viewport.
				rendertest.DrawRectNdc(-1, -1, 1, 1)
			})
		})
	})

	t.Run("MakeRegion", func(t *testing.T) {
		assert := assert.New(t)
		x, y, dx, dy := 4, 8, 6, 2
		region := gui.MakeRegion(x, y, dx, dy)

		assert.Equal(region.X, x)
		assert.Equal(region.Y, y)
		assert.Equal(region.Dx, dx)
		assert.Equal(region.Dy, dy)
	})
}
