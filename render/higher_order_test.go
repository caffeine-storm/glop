package render_test

import (
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
	"github.com/stretchr/testify/assert"
)

func TestWithMatrixMode(t *testing.T) {
	foo := 42

	render.WithMatrixMode(gl.MODELVIEW, func() {
		foo = 17
	})

	assert.Equal(t, foo, 17)
}
