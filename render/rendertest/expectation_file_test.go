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

	if checker.PathNumber(0) != "testdata/checker/0.png" {
		t.Fatalf("path number 0 should look for 0.png")
	}

	if checker.PathExtension("txt") != "testdata/checker/0.txt" {
		t.Fatalf("path extension 'txt' should look for 0.txt")
	}

	if checker.Path(rendertest.TestNumber(42), rendertest.FileExtension("tar.gz")) != "testdata/checker/42.tar.gz" {
		t.Fatalf("path should be fully customizable")
	}
}
