package rendertest

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"path"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/imgmanip"
	"github.com/runningwild/glop/render"
)

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

func GivenATexture(imageFilePath string) gl.Texture {
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

func DrawTexturedQuad(pixelBounds image.Rectangle, tex gl.Texture, shaders *render.ShaderBank) {
	var left, right, top, bottom int32 = 0, int32(pixelBounds.Dx()), int32(pixelBounds.Dy()), 0
	var texleft, texright, textop, texbottom int32 = 0, 1, 1, 0

	// - enable texturing shaders
	WithShaderProgs(shaders, debug_vertex_shader, debug_fragment_shader, func() {
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
