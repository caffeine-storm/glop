package render_test

import (
	"testing"

	"github.com/caffeine-storm/glop/render"
	"github.com/caffeine-storm/mathgl"
	"github.com/stretchr/testify/assert"
)

func TestShowMatFormat(t *testing.T) {
	exampleMat := mathgl.Mat4{
		0.0, 1.0, 2.0 / 3.0, 47,
		0.0, 1.40129846e-45, 1.17549435e-38, 0.2,
		1.0, 1.5, 1.75, 1.99999988e+9,
		2.0, 16777215, 3.40282347e+38, 42.13,
	}

	result := render.Showmat(exampleMat).String()
	expected := `0.0000000, 1.0000000, 0.6666669, 47.000000
0.0000000, 1.401e-45, 1.175e-38, 0.2000000
1.0000000, 1.5000000, 1.7500000, 1.999e+09
2.0000000, 16777215., 3.402e+38, 42.130001`

	assert.Equal(t, expected, result)
}
