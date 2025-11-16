package strmanip_test

import (
	"fmt"
	"testing"

	"github.com/caffeine-storm/glop/strmanip"
	"github.com/stretchr/testify/assert"
)

func TestShow(t *testing.T) {
	testtable := []struct {
		Name     string
		Input    []string
		Expected string
	}{
		{
			"empty",
			[]string{},
			"[]",
		},
		{
			"singleton",
			[]string{"foo"},
			`["foo"]`,
		},
		{
			"triple",
			[]string{"foo", "bar", "baz"},
			`["foo", "bar", "baz"]`,
		},
		{
			"escaping",
			[]string{"has\"quotes\"", "new\nline", ""},
			`["has\"quotes\"", "new\nline", ""]`,
		},
	}

	for _, testcase := range testtable {
		t.Run(testcase.Name, func(t *testing.T) {
			joined := fmt.Sprintf("%s", strmanip.Show(testcase.Input))
			assert.Equal(t, testcase.Expected, joined)
		})
	}
}

func TestShowNonStrings(t *testing.T) {
	result := strmanip.Show([]float32{3.14, 2.7, -1})
	assert.Equal(t, result, "[\"3.14\", \"2.7\", \"-1\"]")
}
