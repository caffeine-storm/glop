package debug

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/imgmanip"
)

func getBoundTextureSize() (width int, height int) {
	buffer := []int32{0}

	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_WIDTH, buffer)
	width = int(buffer[0])

	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_HEIGHT, buffer)
	height = int(buffer[0])

	return
}

func getBoundTextureFormat() gl.GLenum {
	buffer := []int32{0}
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_INTERNAL_FORMAT, buffer)
	return gl.GLenum(buffer[0])
}

func getBytesPerPixel(textureFormat gl.GLenum) int {
	ret, ok := map[gl.GLenum]int{
		gl.RGBA:            4,
		gl.LUMINANCE_ALPHA: 2,
	}[textureFormat]

	if !ok {
		panic(fmt.Errorf("unknown texture format: %d", textureFormat))
	}

	return ret
}

type TexFormat int

const (
	TexFormatRGBA           = gl.RGBA
	TexFormatLuminanceAlpha = gl.LUMINANCE_ALPHA
)

func (tf TexFormat) String() string {
	switch tf {
	case TexFormatRGBA:
		return "gl.RGBA"
	case TexFormatLuminanceAlpha:
		return "gl.LUMINANCE_ALPHA"
	default:
		panic(fmt.Errorf("unknown textureformat %d", int(tf)))
	}
}

func summarize(data []byte) string {
	datalen := len(data)
	if datalen > 64 {
		data = data[:64]
	}

	hex := hex.Dump(data)

	return fmt.Sprintf("data[0:64 of %d]: %q", datalen, hex)
}

func DumpTexture(textureId gl.Texture) (*image.RGBA, error) {
	textureId.Bind(gl.TEXTURE_2D)

	textureWidth, textureHeight := getBoundTextureSize()
	texformat := getBoundTextureFormat()
	bytesPerPixel := getBytesPerPixel(texformat)
	data := make([]byte, textureWidth*textureHeight*bytesPerPixel)

	gl.GetTexImage(gl.TEXTURE_2D, 0, texformat, gl.UNSIGNED_BYTE, data)

	glog.TraceLogger().Trace("DumpTexture", "data-from-gl", summarize(data), "texformat", TexFormat(texformat))

	var img image.Image
	switch texformat {
	case TexFormatRGBA:
		rgba := image.NewRGBA(image.Rect(0, 0, textureWidth, textureHeight))
		rgba.Pix = data
		img = rgba
	case TexFormatLuminanceAlpha:
		ga := imgmanip.NewGrayAlpha(image.Rect(0, 0, textureWidth, textureHeight))
		ga.Pix = data
		img = ga
	default:
		panic(fmt.Errorf("unknown texformat: %d", int(texformat)))
	}

	// We need to flip the image about the horizontal midline because OpenGL
	// dumps from the bottom-to-top.
	return imgmanip.ToRGBA(imgmanip.VertFlipped{Image: img}), nil
}

func DumpTextureAsPngFile(textureId gl.Texture, path string) error {
	outfile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("couldn't os.Create %q: %w", path, err)
	}
	defer outfile.Close()

	return DumpTextureAsPng(textureId, outfile)
}

func DumpTextureAsPng(textureId gl.Texture, outfile io.Writer) error {
	img, err := DumpTexture(textureId)
	if err != nil {
		return fmt.Errorf("couldn't DumpTexture: %w", err)
	}

	err = png.Encode(outfile, img)
	if err != nil {
		return fmt.Errorf("couldn't png.Encode: %w", err)
	}
	return nil
}
