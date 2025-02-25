package strmanip_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/strmanip"
	"github.com/stretchr/testify/assert"
)

func TestShow(t *testing.T) {
	testdata := []string{
		"foo",
		"bar",
		"baz",
	}

	joined := fmt.Sprintf("%s", strmanip.Show(testdata))

	assert.Equal(t, `["foo", "bar", "baz"]`, joined)
}
