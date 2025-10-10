package rendertest_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"
	"testing"

	"github.com/runningwild/glop/gloptest"
	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/runningwild/glop/strmanip"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var transparent = color.RGBA{}
var black = color.RGBA{A: 255}

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
			lhs := rendertest.MustLoadTestImage("checker")
			rhs := rendertest.MustLoadTestImage("checker")

			conveyResult := rendertest.ShouldLookLike(lhs, lhs, rendertest.BackgroundColour(transparent))
			if conveyResult != "" {
				t.Fatalf("two references to the same image object should look alike but got mismatch: %q", conveyResult)
			}
			conveyResult = rendertest.ShouldLookLike(lhs, rhs, rendertest.BackgroundColour(transparent))
			if conveyResult != "" {
				t.Fatalf("two loads of the same image should look alike but got mismatch: %q", conveyResult)
			}
		})
	})
}

func TestComparingPngsAgainstPngs(t *testing.T) {
	t.Run("api is ergonomic", func(t *testing.T) {
		t.Run("can compare golang image to expected file", func(t *testing.T) {
			someImage := image.NewNRGBA(image.Rect(0, 0, 50, 50))
			red := color.RGBA{255, 0, 0, 255}
			draw.Draw(someImage, someImage.Bounds(), image.NewUniform(red), image.Point{}, draw.Src)

			expectedFileAsString := "red"
			mustBeEmpty := rendertest.ShouldLookLikeFile(someImage, expectedFileAsString)
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

			rgbaImage := image.NewNRGBA(someImage.Bounds())
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
	t.Run("opaque result vs. transparent expectation", func(t *testing.T) {
		// rendertest.ShouldLookLikeFile should work out-of-the-box when the
		// expectation file has transparent pixels.
		testbuilder.WithSize(64, 64, func(queue render.RenderQueueInterface) {
			queue.Queue(func(st render.RenderQueueState) {
				// - Convert it to a texture
				tex, cleanup := rendertest.GivenATexture("checker/0.png")
				defer cleanup()

				render.WithBlankScreen(0, 0, 1, 1, func() {
					// - Blit the texture accross the entire viewport
					rendertest.DrawTexturedQuad(image.Rect(0, 0, 64, 64), tex, st.Shaders())
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

func TestStrangeComparisonBehaviour(t *testing.T) {
	testref := rendertest.NewTestdataReference("tut-regr")
	lhsImage := rendertest.MustLoadImageNRGBA(testref.Path(rendertest.TestNumber(0)))
	rhsImage := rendertest.MustLoadImageNRGBA(testref.Path(rendertest.TestNumber(1)))
	lhsbytes := lhsImage.Pix
	rhsbytes := rhsImage.Pix

	cmpresult := rendertest.CompareWithThreshold(lhsbytes, rhsbytes, rendertest.Threshold(0))
	if cmpresult == 0 {
		panic(fmt.Errorf("the input images should be different (even if just a bit) but they compared as the same!"))
	}

	cmpresult = rendertest.CompareWithThreshold(lhsbytes, rhsbytes, rendertest.Threshold(13))
	if cmpresult != 0 {
		panic(fmt.Errorf("the input images should within a threshold of each other"))
	}

	lhsBlitted := imgmanip.DrawAsNrgbaWithBackground(lhsImage, rendertest.BackgroundColour(black))
	lhsbytes = lhsBlitted.Pix
	rhsBlitted := imgmanip.DrawAsNrgbaWithBackground(rhsImage, rendertest.BackgroundColour(black))
	rhsbytes = rhsBlitted.Pix
	deltaBytes := rendertest.ComputeImageDifference(lhsbytes, rhsbytes)
	// We can assume input images are 1024x768
	for i, v := range deltaBytes {
		if v <= 13 {
			continue
		}

		t.Fail()
		x := (i / 4) % 1024
		y := (i / 4) / 1024
		channelidx := i % 4
		channel := []string{"r", "g", "b", "a"}[channelidx]
		t.Logf("mismatch at (%d, %d): %s=%v", x, y, channel, v)
	}
}

func shouldExistOnDisk(filepathAny interface{}, _ ...interface{}) string {
	filepath := filepathAny.(string)

	if val, e := os.Stat(filepath); e == nil {
		if val != nil {
			return ""
		}
	}

	return fmt.Sprintf("file %q should have existed on disk", filepath)
}

var _ Assertion = shouldExistOnDisk

// TODO(tmckee:#15): move this to gloptest; it's generally useful
func ShouldContainLog(actual interface{}, args ...interface{}) string {
	loglines, ok := actual.([]string)
	if !ok {
		panic(fmt.Errorf("ShouldContainLog needs a slice of strings for its 'actual' argument; got %T", actual))
	}

	filters := []string{}
	for _, arg := range args {
		filter, ok := arg.(string)
		if !ok {
			panic(fmt.Errorf("ShouldContainLog needs strings for its 'args' argument; got %T", actual))
		}

		filters = append(filters, filter)
	}

	// Make sure that 'loglines' contains at least one line that matches all the
	// 'filters'.
	for _, line := range loglines {
		mismatch := false
		for _, filter := range filters {
			if strings.Contains(line, filter) {
				continue
			}

			mismatch = true
		}

		if !mismatch {
			return ""
		}
	}

	return fmt.Sprintf("no log line matched all the filters\nlogs: %+v, filters: %+v", loglines, strmanip.Show(filters))
}

func TestCmpSpecs(t *testing.T) {
	blue := color.RGBA{
		R: 0,
		G: 0,
		B: 255,
		A: 255,
	}

	Convey("comparison helpers", t, func() {
		Convey("should be ergonomic", func() {
			Convey("for raw images", func() {
				checkers := rendertest.MustLoadTestImage("checker")
				So(checkers, rendertest.ShouldLookLikeFile, "checker")

				// When comparing raw images, the transparency must _match_.
				checkersOnBlue := imgmanip.DrawAsNrgbaWithBackground(checkers, blue)
				So(checkersOnBlue, rendertest.ShouldNotLookLikeFile, "checker")
			})

			Convey("for rendered textures", func(c C) {
				testbuilder.WithExpectation(c, "checker", rendertest.BackgroundColour(blue),
					func(st render.RenderQueueState) {
						tex, cleanup := rendertest.GivenATexture("checker/0.png")
						defer cleanup()

						render.WithBlankScreen(0, 0, 1, 1, func() {
							rendertest.DrawTexturedQuad(image.Rect(0, 0, 64, 64), tex, st.Shaders())
						})
					})
			})
		})

		Convey("should dump rejection files", func() {
			Convey("when sizes mismatch", func() {
				expectedFile := rendertest.NewTestdataReference("checker-fail")
				rejFileName := rendertest.MakeRejectName(expectedFile.Path(), ".png")

				img := rendertest.MustLoadTestImage(expectedFile)
				biggerImg := imgmanip.Scale(img, 2, 2)

				// Check that we're not accidentally running when there's already a
				// rejection file present.
				if shouldExistOnDisk(rejFileName) == "" {
					panic(fmt.Errorf("precondition violated: there's already a rejection file at %q", rejFileName))
				}

				defer func() {
					// Clean up the rejection file if it exists; whether we pass this
					// test or not, it must not be there afterwards.
					os.Remove(rejFileName)
				}()

				var compResult string
				logoutput := gloptest.CollectOutput(func() {
					compResult = rendertest.ShouldLookLikeFile(biggerImg, expectedFile, rendertest.MakeRejectFiles(true))
				})
				So(compResult, ShouldNotEqual, "") // b/c the images are different
				So(rejFileName, shouldExistOnDisk)

				// The log should mention that a comparison failed.
				So(logoutput, ShouldContainLog, "level=ERROR", `msg="size mismatch"`)
			})
		})
	})
}

func TestImageComparisonHelperArgTypes(t *testing.T) {
	stubRenderQueue := rendertest.MakeStubbedRenderQueue()
	t.Run("ShouldLookLikeText", func(t *testing.T) {
		t.Run("support string literal for testdata ref", func(t *testing.T) {
			res := rendertest.ShouldLookLikeText(stubRenderQueue, "emptyimg")
			if res != "" {
				t.Fatalf("unexpected failure: %q", res)
			}
		})
		t.Run("support TestDataReference instance for testdata ref", func(t *testing.T) {
			res := rendertest.ShouldLookLikeText(stubRenderQueue, rendertest.NewTestdataReference("emptyimg"))
			if res != "" {
				t.Fatalf("unexpected failure: %q", res)
			}
		})
	})
	t.Run("ShouldLookLikeFile", func(t *testing.T) {
		t.Run("support string literal for testdata ref", func(t *testing.T) {
			res := rendertest.ShouldLookLikeFile(stubRenderQueue, "emptyimg")
			if res != "" {
				t.Fatalf("unexpected failure: %q", res)
			}
		})
		t.Run("support TestDataReference instance for testdata ref", func(t *testing.T) {
			res := rendertest.ShouldLookLikeFile(stubRenderQueue, rendertest.NewTestdataReference("emptyimg"))
			if res != "" {
				t.Fatalf("unexpected failure: %q", res)
			}
		})
	})
}

func TestComparisonWithPanickyRenderQueue(t *testing.T) {
	t.Run("if the queue panics, it should be apparent", func(t *testing.T) {
		panickyQueue := rendertest.MakePanicingRenderQueue()
		expectedErrorMessage := (&rendertest.PanicQueueShouldNotBeCalledError{}).Error()
		testResult := rendertest.ShouldLookLikeFile(panickyQueue, "red")
		require.NotEqual(t, testResult, "", "the test should have failed")
		assert.Contains(t, testResult, expectedErrorMessage, "the failure message must include the correct reason")
	})
}
