package debug

import (
	"image"

	"github.com/go-gl-legacy/gl"
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

	return img, nil
}
