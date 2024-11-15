package rendertest_test

import (
	"testing"

	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func TestExpectationFile(t *testing.T) {
	result := rendertest.ExpectationFile("lol", "pgm", 42)
	assert.Equal(t, "../testdata/text/lol/42.pgm", result)
}

func TestMakeRejectName(t *testing.T) {
	reject0 := rendertest.MakeRejectName("../testdata/text/lol/0.pgm", ".pgm")
	assert.Equal(t, "../testdata/text/lol/0.rej.pgm", reject0)
}
