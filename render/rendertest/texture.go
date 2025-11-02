package rendertest

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path"

	"github.com/caffeine-storm/gl"
	"github.com/runningwild/glop/glog"
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
	imgmanip.FlipVertically[*image.RGBA](img, img.Pix)

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

func stringifyColorModel(m color.Model) string {
	switch m {
	case color.RGBAModel:
		return "color.RGBAModel"
	case color.RGBA64Model:
		return "color.RGBA64Model"
	case color.NRGBAModel:
		return "color.NRGBAModel"
	case color.NRGBA64Model:
		return "color.NRGBA64Model"
	case color.AlphaModel:
		return "color.AlphaModel"
	case color.Alpha16Model:
		return "color.Alpha16Model"
	case color.GrayModel:
		return "color.GrayModel"
	case color.Gray16Model:
		return "color.Gray16Model"
	default:
		return "dunno lol"
	}
}

func GivenATexture(imageFilePath string) (gl.Texture, func()) {
	imageFilePath = path.Join("testdata", imageFilePath)
	file, err := os.Open(imageFilePath)
	if err != nil {
		panic(fmt.Errorf("couldn't open file %q: %w", imageFilePath, err))
	}

	img, imgFmt, err := image.Decode(file)
	if err != nil {
		panic(fmt.Errorf("couldn't image.Decode: %w", err))
	}
	colourModel := stringifyColorModel(img.ColorModel())

	fmt.Printf("img type: %T, img fmt: %s, colourModel: %s\n", img, imgFmt, colourModel)

	// Note: we use an RGBA image here instead of an NRGBA because we want to be
	// uploading alpha-premultipled colours to OpenGL.
	rgbaImage := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImage, img.Bounds(), img, image.Point{}, draw.Src)

	logger := glog.WarningLogger()
	for idx := 0; idx < len(rgbaImage.Pix); idx += 4 {
		if max(rgbaImage.Pix[idx+0], rgbaImage.Pix[idx+1], rgbaImage.Pix[idx+2]) > rgbaImage.Pix[idx+3] {
			logger.Warn("found non-normalized colour", "idx", idx)
		}
	}
	return uploadTextureFromImage(rgbaImage), func() {
		gl.Texture(0).Bind(gl.TEXTURE_2D)
	}
}

func DrawTexturedQuad(pixelBounds image.Rectangle, tex gl.Texture, shaders *render.ShaderBank) {
	var left, bottom int32 = int32(pixelBounds.Min.X), int32(pixelBounds.Min.Y)
	var right, top int32 = int32(pixelBounds.Max.X), int32(pixelBounds.Max.Y)
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
		defer vertexBuffer.Delete()
		vertexBuffer.Bind(gl.ARRAY_BUFFER)
		// stride is how many bytes per vertex
		// 4 components * 4 bytes per component
		stride := 16
		gl.BufferData(gl.ARRAY_BUFFER, stride*len(verts), verts, gl.STATIC_DRAW)

		// - setup rendering parameters
		tex.Bind(gl.TEXTURE_2D)
		defer gl.Texture(0).Bind(gl.TEXTURE_2D)

		gl.EnableClientState(gl.VERTEX_ARRAY)
		defer gl.DisableClientState(gl.VERTEX_ARRAY)

		gl.VertexPointer(2, gl.INT, stride, nil)

		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
		// First texture co-ordinate is two components in; each component is 4
		// bytes so the first texture co-ordinate is 8 bytes in.
		gl.TexCoordPointer(2, gl.INT, stride, uintptr(8))

		gl.Buffer(0).Bind(gl.ELEMENT_ARRAY_BUFFER)

		// - render geometry
		gl.Enable(gl.BLEND)
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ZERO, gl.ONE)
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
