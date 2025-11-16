package render_test

import (
	"testing"

	"github.com/caffeine-storm/glop/render"
	"github.com/stretchr/testify/assert"
)

func TestNormalizationRoundTripping(t *testing.T) {
	assert := assert.New(t)
	for i := 0; i < 256; i++ {
		asByte := uint8(i)
		asFloat := render.ByteToNormalizedColour(asByte)
		backAgain := render.NormalizedColourToByte(asFloat)

		assert.Equal(asByte, backAgain)
	}
}
