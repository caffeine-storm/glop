package rendertest_test

import (
	"testing"

	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func TestExpectationFilePaths(t *testing.T) {
	t.Run("from non-local package", func(t *testing.T) {
		result := rendertest.ExpectationFile("text/lol", "pgm", 42)
		assert.Equal(t, "testdata/text/lol/42.pgm", result)
	})
}

func TestTestdataReference(t *testing.T) {
	checker := rendertest.NewTestdataReference("checker")

	if checker != "checker" {
		t.Fatalf("a testdata reference should look like its key")
	}

	if checker.Path() != "testdata/checker/0.png" {
		t.Fatalf("default path should look for 0.png")
	}
}
