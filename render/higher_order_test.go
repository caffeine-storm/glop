package render_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func pickADifferentMode(someMatrixMode render.MatrixMode) render.MatrixMode {
	switch someMatrixMode {
	case render.MatrixModeModelView:
		return render.MatrixModeProjection
	case render.MatrixModeProjection:
		fallthrough
	case render.MatrixModeTexture:
		fallthrough
	case render.MatrixModeColour:
		return render.MatrixModeModelView
	default:
		panic(fmt.Errorf("bad matrixmode: %d", someMatrixMode))
	}
}

func TestWithMatrixMode(t *testing.T) {
	var beforeMode, duringMode, targetMode, afterMode render.MatrixMode
	rendertest.WithGl(func() {
		beforeMode = render.GetCurrentMatrixMode()
		targetMode = pickADifferentMode(beforeMode)

		if beforeMode == targetMode {
			panic(fmt.Errorf("bad test; need to find a _different_ mode"))
		}

		render.WithMatrixMode(targetMode, func() {
			duringMode = render.GetCurrentMatrixMode()
		})

		afterMode = render.GetCurrentMatrixMode()
	})

	assert.Equal(t, duringMode, targetMode)
	assert.Equal(t, afterMode, beforeMode)
}
