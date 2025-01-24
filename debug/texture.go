package debug

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"github.com/go-gl-legacy/gl"
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

func DumpTexture(textureId gl.Texture) (*image.RGBA, error) {
	gl.Enable(gl.TEXTURE_2D)
	textureId.Bind(gl.TEXTURE_2D)

	textureWidth, textureHeight := getBoundTextureSize()
	img := image.NewRGBA(image.Rect(0, 0, textureWidth, textureHeight))

	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, img.Pix)

	// We need to flip the image about the horizontal midline because OpenGL
	// dumps from the bottom-to-top.
	imgmanip.FlipVertically(img)

	return img, nil
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
