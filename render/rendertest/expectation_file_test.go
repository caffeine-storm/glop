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

	t.Run("looks like its key", func(t *testing.T) {
		assert.Equal(t, string(checker), "checker", "a testdata reference should look like its key")
	})

	t.Run(".Path() refers to a file in testdata/", func(t *testing.T) {
		assert.Equal(t, checker.Path(), "testdata/checker/0.png", "default path should look for 0.png")
	})

	t.Run(".PathNumber(n) fills in the right number", func(t *testing.T) {
		assert.Equal(t, checker.PathNumber(0), "testdata/checker/0.png", "path number 0 should look for 0.png")

		assert.Equal(t, checker.PathNumber(7), "testdata/checker/7.png", "path number 7 should look for 7.png")
	})

	t.Run(".PathExtension(foo) should look for something.foo", func(t *testing.T) {
		assert.Equal(t, checker.PathExtension("txt"), "testdata/checker/0.txt", "path extension 'txt' should look for 0.txt")
	})

	t.Run(".Path supports multiple option parameters", func(t *testing.T) {
		args := []interface{}{
			rendertest.TestNumber(42),
			rendertest.FileExtension("tar.gz"),
		}
		assert.Equal(t, checker.Path(args...), "testdata/checker/42.tar.gz", "path should be fully customizable")
	})

	t.Run("rejects paths already starting with 'testdata'", func(t *testing.T) {
		assert.Panics(t, func() {
			rendertest.NewTestdataReference("testdata/but/fail")
		})
	})
}
