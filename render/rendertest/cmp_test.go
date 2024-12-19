package rendertest_test

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/runningwild/glop/render/rendertest"
	"github.com/stretchr/testify/assert"
)

func TestFileNameHelpers(t *testing.T) {
	t.Run("TestExpectationFile", func(t *testing.T) {
		result := rendertest.ExpectationFile("text/lol", "pgm", 42)
		assert.Equal(t, "../../testdata/text/lol/42.pgm", result)
	})

	t.Run("TestMakeRejectName", func(t *testing.T) {
		reject0 := rendertest.MakeRejectName("../../testdata/text/lol/0.pgm", ".pgm")
		assert.Equal(t, "../../testdata/text/lol/0.rej.pgm", reject0)
	})
}

func TestPixelComparisonIsFuzzy(t *testing.T) {
	t.Run("TestShouldLookLike", func(t *testing.T) {
		t.Run("matching", func(t *testing.T) {
			someBytes := bytes.NewBuffer([]byte("some bytes"))
			someOtherBytes := bytes.NewBuffer([]byte("some bytes"))

			conveyResult := rendertest.ShouldLookLike(someBytes, someOtherBytes)
			if conveyResult != "" {
				// empty string represents 'success' in Convey
				t.Fatalf("two readers over the same data should look the same")
			}
		})

		t.Run("mismatch", func(t *testing.T) {
			someBytes := bytes.NewBuffer([]byte("some bytes"))
			someOtherBytes := bytes.NewBuffer([]byte("different bytes"))

			conveyResult := rendertest.ShouldLookLike(someBytes, someOtherBytes)
			if conveyResult == "" {
				// empty string represents 'success' in Convey
				t.Fatalf("two readers over different data should look dissimilar")
			}
		})

		t.Run("fuzzy matching", func(t *testing.T) {
			t.Run("within tolerance should pass", func(t *testing.T) {
				someBytes := bytes.NewBuffer([]byte("some bytes"))
				someOtherBytes := bytes.NewBuffer([]byte("some bytes"))

				someOtherBytes.Bytes()[0] += 1

				conveyResult := rendertest.ShouldLookLike(someBytes, someOtherBytes, rendertest.Threshold(5))
				if conveyResult != "" {
					t.Fatalf("a tolerance of 5 should not have been exceeded")
				}
			})
			t.Run("outside of tolerance should fail", func(t *testing.T) {
				someBytes := bytes.NewBuffer([]byte("some bytes"))
				someOtherBytes := bytes.NewBuffer([]byte("some bytes"))

				someOtherBytes.Bytes()[0] += 5

				conveyResult := rendertest.ShouldLookLike(someBytes, someOtherBytes, rendertest.Threshold(2))
				if conveyResult == "" {
					t.Fatalf("a tolerance of 2 should have been exceeded")
				}
			})
			t.Run("at tolerance should pass", func(t *testing.T) {
				someBytes := bytes.NewBuffer([]byte("some bytes"))
				someOtherBytes := bytes.NewBuffer([]byte("some bytes"))

				someOtherBytes.Bytes()[0] += 5

				conveyResult := rendertest.ShouldLookLike(someBytes, someOtherBytes, rendertest.Threshold(5))
				if conveyResult != "" {
					t.Fatalf("hitting a tolerance of 5 should not cause a failure")
				}
			})
		})
	})
}

func TestComparingPngsAgainstPngs(t *testing.T) {
	t.Run("api is ergonomic", func(t *testing.T) {
		t.Run("can compare golang image to expected file", func(t *testing.T) {
			someImage := image.NewRGBA(image.Rect(0, 0, 50, 50))
			red := color.RGBA{255, 0, 0, 255}
			draw.Draw(someImage, someImage.Bounds(), image.NewUniform(red), image.Point{}, draw.Src)

			expectedFile := "debug/red"
			mustBeEmpty := rendertest.ShouldLookLikeFile(someImage, expectedFile)
			if mustBeEmpty != "" {
				t.Fatalf("expected a 'match' but got failure %q", mustBeEmpty)
			}
		})
	})
}

func TestCompareWithThreshold(t *testing.T) {
	t.Run("same slices are equal", func(t *testing.T) {
		lhs := []byte("lol")
		rhs := []byte("lol")
		cmp := rendertest.CompareWithThreshold(lhs, rhs, rendertest.Threshold(2))
		assert.Equal(t, 0, cmp)
	})
	t.Run("slice less than", func(t *testing.T) {
		lhs := []byte("aol")
		rhs := []byte("lol")
		cmp := rendertest.CompareWithThreshold(lhs, rhs, rendertest.Threshold(2))
		assert.Equal(t, -1, cmp)
	})
	t.Run("slice greater than", func(t *testing.T) {
		lhs := []byte("lol")
		rhs := []byte("aol")
		cmp := rendertest.CompareWithThreshold(lhs, rhs, rendertest.Threshold(2))
		assert.Equal(t, 1, cmp)
	})
	t.Run("slice shorter than", func(t *testing.T) {
		lhs := []byte("lol")
		rhs := []byte("lolol")
		cmp := rendertest.CompareWithThreshold(lhs, rhs, rendertest.Threshold(2))
		assert.Equal(t, -1, cmp)
	})
	t.Run("slice longer than", func(t *testing.T) {
		lhs := []byte("lolol")
		rhs := []byte("lol")
		cmp := rendertest.CompareWithThreshold(lhs, rhs, rendertest.Threshold(2))
		assert.Equal(t, 1, cmp)
	})
	t.Run("slice at threshold", func(t *testing.T) {
		lhs := []byte("lolol")
		rhs := []byte("lolok")
		cmp := rendertest.CompareWithThreshold(lhs, rhs, rendertest.Threshold(2))
		assert.Equal(t, 0, cmp)
	})
}
