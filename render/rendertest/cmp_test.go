package rendertest_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"testing"

	"github.com/runningwild/glop/debug/debugtest"
	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestFileNameHelpers(t *testing.T) {
	t.Run("TestExpectationFile", func(t *testing.T) {
		result := rendertest.ExpectationFile("text/lol", "pgm", 42)
		assert.Equal(t, "testdata/text/lol/42.pgm", result)
	})

	t.Run("TestMakeRejectName", func(t *testing.T) {
		reject0 := rendertest.MakeRejectName("testdata/text/lol/0.pgm", ".pgm")
		assert.Equal(t, "testdata/text/lol/0.rej.pgm", reject0)
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

		t.Run("supports image.Image interface", func(t *testing.T) {
			lhs := rendertest.MustLoadImage("checker/0.png")
			rhs := rendertest.MustLoadImage("checker/0.png")

			conveyResult := rendertest.ShouldLookLike(lhs, rhs)
			if conveyResult != "" {
				t.Fatalf("two loads of the same image should look alike")
			}
		})
	})
}

func TestComparingPngsAgainstPngs(t *testing.T) {
	t.Run("api is ergonomic", func(t *testing.T) {
		t.Run("can compare golang image to expected file", func(t *testing.T) {
			someImage := image.NewRGBA(image.Rect(0, 0, 50, 50))
			red := color.RGBA{255, 0, 0, 255}
			draw.Draw(someImage, someImage.Bounds(), image.NewUniform(red), image.Point{}, draw.Src)

			expectedFile := "red"
			mustBeEmpty := rendertest.ShouldLookLikeFile(someImage, expectedFile)
			if mustBeEmpty != "" {
				t.Fatalf("expected a 'match' but got failure %q", mustBeEmpty)
			}
		})
		t.Run("comparisons are correct w.r.t. symmetries", func(t *testing.T) {
			imageFilePath := "testdata/checker/0.png"
			testdata, err := os.Open(imageFilePath)
			if err != nil {
				panic(fmt.Errorf("couldn't open %q: %w", imageFilePath, err))
			}
			someImage, _, err := image.Decode(testdata)
			if err != nil {
				panic(fmt.Errorf("couldn't decode %q: %w", imageFilePath, err))
			}

			rgbaImage := image.NewRGBA(someImage.Bounds())
			draw.Draw(rgbaImage, someImage.Bounds(), someImage, image.Point{}, draw.Src)

			expectedFile := "checker"
			mustBeEmpty := rendertest.ShouldLookLikeFile(rgbaImage, expectedFile)
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

func TestCompareTransparentExpectations(t *testing.T) {
	t.Run("transparent result vs. transparent expecation", func(t *testing.T) {
		lhs := rendertest.MustLoadImage("checker/0.png")
		rhs := rendertest.MustLoadImage("checker/0.png")
		// Use a transparent background for the sake of this comparison.
		transparent := color.RGBA{}
		result := rendertest.ImagesAreWithinThreshold(lhs, rhs, rendertest.Threshold(0), transparent)
		assert.Equal(t, result, true)

		black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
		result = rendertest.ImagesAreWithinThreshold(lhs, rhs, rendertest.Threshold(5), black)
		assert.Equal(t, result, false)
	})

	t.Run("opaque result vs. transparent expectation", func(t *testing.T) {
		// rendertest.ShouldLookLikeFile should work out-of-the-box when the
		// expectation file has transparent pixels.
		rendertest.WithGlForTest(64, 64, func(_ system.System, queue render.RenderQueueInterface) {
			queue.Queue(func(st render.RenderQueueState) {
				// - Convert it to a texture
				tex := debugtest.GivenATexture("checker/0.png")

				rendertest.WithClearColour(0, 0, 1, 1, func() {
					// - Blit the texture accross the entire viewport
					debugtest.DrawTexturedQuad(image.Rect(0, 0, 64, 64), tex, st.Shaders())
				})
			})
			queue.Purge()

			conveyResult := rendertest.ShouldLookLikeFile(queue, "checker", rendertest.BackgroundColour(color.RGBA{R: 0, G: 0, B: 255, A: 255}))
			conveySuccess := ""
			if conveyResult != conveySuccess {
				t.Fatalf("ShouldLookLike returned a mismatch: %q", conveyResult)
			}
		})
	})
}

func TestCmpSpecs(t *testing.T) {
	blue := color.RGBA{
		R: 0,
		G: 0,
		B: 255,
		A: 255,
	}

	Convey("comparison helpers should be ergonomic", t, func() {
		Convey("for raw images", func() {
			checkers := rendertest.MustLoadRGBAImage("checker/0.png")
			So(checkers, rendertest.ShouldLookLikeFile, "checker")

			// When comparing raw images, the transparency must _match_.
			checkersOnBlue := imgmanip.DrawAsRgbaWithBackground(checkers, blue)
			So(checkersOnBlue, rendertest.ShouldNotLookLikeFile, "checker")
		})

		Convey("for rendered textures", func() {
			rendertest.WithGlForTest(64, 64, func(_ system.System, queue render.RenderQueueInterface) {
				queue.Queue(func(st render.RenderQueueState) {
					tex := debugtest.GivenATexture("checker/0.png")

					rendertest.WithClearColour(0, 0, 1, 1, func() {
						debugtest.DrawTexturedQuad(image.Rect(0, 0, 64, 64), tex, st.Shaders())
					})
				})
				queue.Purge()

				So(queue, rendertest.ShouldLookLikeFile, "checker", rendertest.BackgroundColour(blue))
			})
		})
	})
}
