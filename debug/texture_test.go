package debug_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path"
	"testing"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/system"
)

const maxuint16 = 0xffff

func foreachPixel(img image.Image, check func(x, y int, col color.Color)) {
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			check(x, y, img.At(x, y))
		}
	}
}

func mustLoadImage(imageFilePath string) image.Image {
	imageFilePath = path.Join("testdata", imageFilePath)
	file, err := os.Open(imageFilePath)
	if err != nil {
		panic(fmt.Errorf("couldn't open file %q: %w", imageFilePath, err))
	}

	img, _, err := image.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't image.Decode: %w", err))
	}

	return img
}

func givenATexture(imageFilePath string) gl.Texture {
	imageFilePath = path.Join("testdata", imageFilePath)
	file, err := os.Open(imageFilePath)
	if err != nil {
		panic(fmt.Errorf("couldn't open file %q: %w", imageFilePath, err))
	}

	img, _, err := image.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't image.Decode: %w", err))
	}

	rgbaImage := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImage, img.Bounds(), img, image.Point{}, draw.Src)

	return uploadTextureFromImage(rgbaImage)
}

func uploadTextureFromImage(img *image.RGBA) gl.Texture {
	bounds := img.Bounds()
	texture := gl.GenTexture()
	texture.Bind(gl.TEXTURE_2D)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	// Need to flip the input image vertically because image.RGBA stores the
	// top-left pixel first but gl.TexImage2D expects the bottom-left pixel
	// first.
	imgmanip.FlipVertically(img)

	// sanity check: what are the alpha values in the image? We're expecting
	// 4 row chunks of 4-black, 4-transparent, 4-red pixels
	// What's the first row?
	var reds, greens, blues []byte
	alphas := []byte{}
	for pix := 0; pix < img.Bounds().Dx(); pix++ {
		reds = append(reds, img.Pix[pix*4+0])
		greens = append(greens, img.Pix[pix*4+1])
		blues = append(blues, img.Pix[pix*4+2])
		alphas = append(alphas, img.Pix[pix*4+3])
	}
	fmt.Printf("reds for row 0: %v\n", reds)
	fmt.Printf("greens for row 0: %v\n", greens)
	fmt.Printf("blues for row 0: %v\n", blues)
	fmt.Printf("alphas for row 0: %v\n", alphas)

	gl.ActiveTexture(gl.TEXTURE0 + 0)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		bounds.Dx(),
		bounds.Dy(),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		img.Pix,
	)

	return texture
}

func isBlack(c color.Color) bool {
	r, g, b, a := c.RGBA()
	if r != 0 {
		return false
	}
	if g != 0 {
		return false
	}
	if b != 0 {
		return false
	}

	return a == maxuint16
}

func isTransparent(c color.Color) bool {
	_, _, _, a := c.RGBA()
	return a == 0
}

func isRed(c color.Color) bool {
	r, g, b, a := c.RGBA()
	if r != maxuint16 {
		return false
	}
	if g != 0 {
		return false
	}
	if b != 0 {
		return false
	}

	return a == maxuint16
}

func isBlue(c color.Color) bool {
	r, g, b, a := c.RGBA()
	if r != 0 {
		return false
	}
	if g != 0 {
		return false
	}
	if b != maxuint16 {
		return false
	}

	return a == maxuint16
}

func withShaderProgs(shaders *render.ShaderBank, vertShader string, fragShader string, fn func()) {
	err := shaders.RegisterShader("debugshaders", vertShader, fragShader)
	if err != nil {
		panic(fmt.Errorf("couldn't register debug shaders: %w", err))
	}

	err = shaders.EnableShader("debugshaders")
	if err != nil {
		panic(fmt.Errorf("couldn't enable debug shaders: %w", err))
	}

	defer func() {
		shaders.EnableShader("")
	}()
	fn()
}

func withClearColour(r, g, b, a gl.GLclampf, fn func()) {
	oldClear := [4]float32{0, 0, 0, 0}

	gl.GetFloatv(gl.COLOR_CLEAR_VALUE, oldClear[:])

	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	defer func() {
		gl.ClearColor(
			gl.GLclampf(oldClear[0]),
			gl.GLclampf(oldClear[1]),
			gl.GLclampf(oldClear[2]),
			gl.GLclampf(oldClear[3]))
	}()
	fn()
}

func blitOntoBlue(img image.Image) *image.RGBA {
	blue := image.NewUniform(color.RGBA{
		R: 0,
		G: 0,
		B: 255,
		A: 255,
	})
	result := image.NewRGBA(img.Bounds())

	draw.Draw(result, img.Bounds(), blue, image.Point{}, draw.Src)
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Over)

	return result
}

func drawTexturedQuad(pixelBounds image.Rectangle, tex gl.Texture, shaders *render.ShaderBank) {
	var left, right, top, bottom int32 = 0, int32(pixelBounds.Dx()), int32(pixelBounds.Dy()), 0
	var texleft, texright, textop, texbottom int32 = 0, 1, 1, 0

	// - enable texturing shaders
	withShaderProgs(shaders, debug_vertex_shader, debug_fragment_shader, func() {
		// - set shader variables/inputs; we want to use the 0'th texture unit.
		shaders.SetUniformI("debugshaders", "tex", 0)

		// - define geometry
		verts := []int32{
			// each vertex is an (x,y) in pixelspace and a (t,s) in texture space
			left, top, texleft, textop,
			left, bottom, texleft, texbottom,
			right, bottom, texright, texbottom,

			left, top, texleft, textop,
			right, bottom, texright, texbottom,
			right, top, texright, textop,
		}

		// - upload geometry
		vertexBuffer := gl.GenBuffer()
		vertexBuffer.Bind(gl.ARRAY_BUFFER)
		// stride is how many bytes per vertex
		// 4 components * 4 bytes per component
		stride := 16
		gl.BufferData(gl.ARRAY_BUFFER, stride*len(verts), verts, gl.STATIC_DRAW)

		// - setup rendering parameters
		tex.Bind(gl.TEXTURE_2D)

		gl.EnableClientState(gl.VERTEX_ARRAY)
		defer gl.DisableClientState(gl.VERTEX_ARRAY)

		gl.VertexPointer(2, gl.INT, stride, nil)

		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
		// First texture co-ordinate is two components in; each component is 4
		// bytes so the first texture co-ordinate is 8 bytes in.
		gl.TexCoordPointer(2, gl.INT, stride, uintptr(8))

		// - render geometry
		indices := []uint16{0, 1, 2, 3, 4, 5}
		gl.DrawElements(gl.TRIANGLES, len(indices), gl.UNSIGNED_SHORT, indices)
	})
}

func mustSaveImage(img image.Image, outputPath string) {
	f, err := os.Create(outputPath)
	if err != nil {
		panic(fmt.Errorf("couldn't os.Create %q: %w", outputPath, err))
	}

	err = png.Encode(f, img)
	if err != nil {
		panic(fmt.Errorf("coudln't png.Encode %q: %w", outputPath, err))
	}
}

func TestTextureDebugging(t *testing.T) {
	t.Run("can dump texture to image object", func(t *testing.T) {
		var dumpedImage *image.RGBA
		var err error

		rendertest.WithGl(func() {
			tex := givenATexture("red/0.png")

			dumpedImage, err = debug.DumpTexture(tex)
			if err != nil {
				t.Fatalf("dumping failed: %v", err)
			}
		})

		// The dump should have produced a 50x50 pixel image of all red.
		foreachPixel(dumpedImage, func(x, y int, col color.Color) {
			if !isRed(col) {
				t.Log("non-red pixel", "x", x, "y", y, "colour", col)
				t.Fail()
			}
		})
	})

	t.Run("can dump to writer", func(t *testing.T) {
		pngBuffer := &bytes.Buffer{}

		rendertest.WithGl(func() {
			tex := givenATexture("red/0.png")

			err := debug.DumpTextureAsPng(tex, pngBuffer)
			if err != nil {
				t.Fatalf("dumping failed: %v", err)
			}
		})

		dumpedImage, err := png.Decode(pngBuffer)
		if err != nil {
			panic(fmt.Errorf("couldn't decode pngBuffer: %w", err))
		}

		// The dump should have produced a 50x50 pixel image of all red.
		foreachPixel(dumpedImage, func(x, y int, col color.Color) {
			if !isRed(col) {
				t.Log("non-red pixel", "x", x, "y", y, "colour", col)
				t.Fail()
			}
		})
	})

	t.Run("can dump non-uniform texture", func(t *testing.T) {
		pngBuffer := &bytes.Buffer{}

		rendertest.WithGl(func() {
			tex := givenATexture("checker/0.png")

			err := debug.DumpTextureAsPng(tex, pngBuffer)
			if err != nil {
				t.Fatalf("dumping failed: %v", err)
			}
		})

		dumpedImage, err := png.Decode(pngBuffer)
		if err != nil {
			panic(fmt.Errorf("couldn't decode pngBuffer: %w", err))
		}

		// When things are quite broken, we'll just ellide noisy log messages.
		logCount := 0
		logProblem := func(args ...interface{}) {
			logCount++
			if logCount > 10 {
				return
			}
			if logCount == 10 {
				t.Log("(supressing duplicate messages")
			}

			t.Log(args...)
		}

		// The dump should have produced a 64x64 pixel image of a cycle of squares
		// (each 4x4 pixels) that are black, transparent then red.
		foreachPixel(dumpedImage, func(x, y int, col color.Color) {
			idx := (y/4)*16 + (x/4)*1
			switch idx % 3 {
			case 0:
				if !isBlack(col) {
					logProblem("non-black pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			case 1:
				if !isTransparent(col) {
					logProblem("non-transparent pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			case 2:
				if !isRed(col) {
					logProblem("non-red pixel", "x", x, "y", y, "colour", col)
					t.Fail()
				}
			}
		})
	})

	t.Run("textures should match the underlying image", func(t *testing.T) {
		// - Load an image
		expectedImage := mustLoadImage("checker/0.png")

		expectedImage = blitOntoBlue(expectedImage)

		// - Screen-shotting reads every pixel without prior knowledge; any
		// 'transparent' pieces will be set to the clear-colour. Set a blue
		// clear-colour (blue does not exist in the checker image), then draw the
		// checker image over top a blue background. Comparing the screenshot to
		// the drawn image is what we need to do.

		bounds := expectedImage.Bounds()
		width, height := bounds.Dx(), bounds.Dy()
		rendertest.WithGlForTest(width, height, func(_ system.System, queue render.RenderQueueInterface) {
			var result bool
			var resultImage image.Image
			queue.Queue(func(st render.RenderQueueState) {
				// - Convert it to a texture
				tex := givenATexture("checker/0.png")

				withClearColour(0, 0, 1, 1, func() {
					// - Blit the texture accross the entire viewport
					drawTexturedQuad(bounds, tex, st.Shaders())

					// - Verify a screenshot matches the image.
					resultImage = debug.ScreenShotRgba(width, height)

					result = rendertest.ImagesAreWithinThreshold(expectedImage, resultImage, rendertest.Threshold(3))
				})
			})
			queue.Purge()

			if result != true {
				// stash a copy of the broken image we just made
				mustSaveImage(resultImage, rendertest.MakeRejectName("testdata/checker/0.png", ".png"))
				t.Fatalf("image mismatch")
			}
		})
	})
}

const debug_vertex_shader string = `
  #version 120
  varying vec3 pos;

  void main() {
    gl_Position = ftransform();
    gl_ClipVertex = gl_ModelViewMatrix * gl_Vertex;
    gl_FrontColor = gl_Color;
    gl_TexCoord[0] = gl_MultiTexCoord0;
    gl_TexCoord[1] = gl_MultiTexCoord1;
    pos = gl_Vertex.xyz;
  }
`

const debug_fragment_shader string = `
  #version 120
  uniform sampler2D tex;

  void main() {
    vec2 tpos = gl_TexCoord[0].xy;
    gl_FragColor = texture2D(tex, tpos);
  }
`
