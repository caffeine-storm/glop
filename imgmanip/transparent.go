package imgmanip

import "image"

func IsTransparent(img *image.NRGBA) bool {
	// An NRGBA image is fully transparent iff every alpha vaule is 0.
	data := img.Pix
	for i := 3; i < len(data); i += 4 {
		if data[i] != 0 {
			return false
		}
	}
	return true
}
