package gui

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/gloptest"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func LoadDictionaryForTest(renderQueue render.RenderQueueInterface, logger glog.Logger) *Dictionary {
	dictReader, err := os.Open("testdata/fonts/dict_10.gob")
	if err != nil {
		panic(fmt.Errorf("couldn't os.Open: %w", err))
	}

	d, err := LoadDictionary(dictReader, renderQueue, logger)
	if err != nil {
		panic(fmt.Errorf("couldn't LoadDictionary: %w", err))
	}

	renderQueue.Queue(render.RenderJob(func(st render.RenderQueueState) {
		fontShaderName := "glop.font"
		if st.Shaders().HasShader(fontShaderName) {
			return
		}

		err = st.Shaders().RegisterShader(fontShaderName, font_vertex_shader, font_fragment_shader)
	}))
	renderQueue.Purge()
	if err != nil {
		panic(fmt.Errorf("couldn't register glop.font!: %w", err))
	}

	return d
}

// Renders the given string with pixel units and an origin at the bottom-left.
func renderStringForTest(toDraw string, x, y, height int, sys system.System, queue render.RenderQueueInterface, just Justification, logger glog.Logger) {
	d := LoadDictionaryForTest(queue, logger)

	queue.Queue(func(st render.RenderQueueState) {
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		d.RenderString(toDraw, Point{x, y}, height, just, st.Shaders())
	})

	queue.Purge()
}

func TestDictionaryMaxHeight(t *testing.T) {
	t.Run("default-height-is-zero", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{}

		require.Equal(0, d.MaxHeight())
	})
	t.Run("zero-height-at-non-zero-offset", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 42,
				Maxy: 42,
			},
		}

		require.Equal(0, d.MaxHeight())
	})
	t.Run("height-clamped-non-negative", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 42,
				Maxy: 0,
			},
		}

		require.Equal(0, d.MaxHeight())
	})
	t.Run("height-is-delta-min-max", func(t *testing.T) {
		require := require.New(t)

		d := Dictionary{
			Data: dictData{
				Miny: 0,
				Maxy: 42,
			},
		}

		require.Equal(42, d.MaxHeight())
	})
}

func TestDictionaryGetInfo(t *testing.T) {
	t.Run("AsciiInfoSucceeds", func(t *testing.T) {
		assert := assert.New(t)

		queue := rendertest.MakeDiscardingRenderQueue()
		d := LoadDictionaryForTest(queue, glog.TraceLogger())

		emptyRuneInfo := runeInfo{}
		// In ascii, all the characters we care about are between 0x20 (space) and
		// 0x7E (tilde).
		for idx := ' '; idx <= '~'; idx++ {
			info := d.getInfo(rune(idx))
			assert.NotEqual(emptyRuneInfo, info)
		}
	})

	// TODO(tmckee): verify slices of texture by runeInfo correspond to correct
	// letters
	// TODO(tmckee): verify texture image in GL matches expectations
}

func DictionaryRenderStringSpec() {
	screenSizeCases := []struct {
		label            string
		screenDimensions Dims
	}{
		{
			label: "natural match to dict dimensions",
			screenDimensions: Dims{
				Dx: 512,
				Dy: 64,
			},
		},
		{
			label: "unnatural dimensions",
			screenDimensions: Dims{
				Dx: 1024,
				Dy: 512,
			},
		},
		{
			label: "wait, wut?",
			screenDimensions: Dims{
				Dx: 512,
				Dy: 64,
			},
		},
		{
			label: "same aspect ratio but bigger",
			screenDimensions: Dims{
				Dx: 1024,
				Dy: 128,
			},
		},
		{
			label: "other dimensions",
			screenDimensions: Dims{
				Dx: 800,
				Dy: 640,
			},
		},
		{
			label: "small dimensions",
			screenDimensions: Dims{
				Dx: 64,
				Dy: 64,
			},
		},
	}
	for testIndex, testcase := range screenSizeCases {
		testnumber := rendertest.TestNumber(testIndex)
		Convey(fmt.Sprintf("[%s]", testcase.label), func() {
			leftPixel := testcase.screenDimensions.Dx / 2
			topPixel := testcase.screenDimensions.Dy / 2
			height := 22
			just := Left
			var logger glog.Logger = glog.TraceLogger()

			rendertest.WithGlForTest(testcase.screenDimensions.Dx, testcase.screenDimensions.Dy, func(sys system.System, render render.RenderQueueInterface) {
				doRenderString := func(toDraw string) {
					renderStringForTest(toDraw, leftPixel, topPixel, height, sys, render, just, logger)
				}

				Convey("Can render 'lol'", func() {
					doRenderString("lol")

					So(render, rendertest.ShouldLookLikeText, "lol", testnumber)
				})

				Convey("Can render 'credits' centred", func() {
					just = Center

					doRenderString("Credits")

					So(render, rendertest.ShouldLookLikeText, "credits", testnumber)
				})

				Convey("Can render somewhere other than the origin", func() {
					Convey("can render at the bottom left", func() {
						leftPixel = 10
						topPixel = 32
						doRenderString("offset")

						So(render, rendertest.ShouldLookLikeText, "offset", testnumber)
					})
				})

				Convey("Can render to a given height", func() {
					height = 5
					doRenderString("tall-or-small")

					So(render, rendertest.ShouldLookLikeText, "tall-or-small", testnumber)
				})

				Convey("stdout isn't spammed by RenderString", func() {
					logger = glog.VoidLogger()

					stdoutLines := gloptest.CollectOutput(func() {
						doRenderString("spam check")
					})

					So(stdoutLines, ShouldEqual, []string{})
				})
			})
		})
	}
}

func TestRunTextSpecs(t *testing.T) {
	Convey("Dictionaries should render strings", t, DictionaryRenderStringSpec)
}
