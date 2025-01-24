package imgmanip

import "image"

// Rewrites the given input image flipping it vertically.
func FlipVertically(img *image.RGBA) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	tmp := make([]byte, width*4)
	for rowIdx := 0; rowIdx < height/2; rowIdx++ {
		a, b := rowIdx, height-rowIdx-1
		if a >= b {
			break
		}
		arow := img.Pix[a*width*4 : (a+1)*width*4]
		brow := img.Pix[b*width*4 : (b+1)*width*4]
		copy(tmp, arow)
		copy(arow, brow)
		copy(brow, tmp)
	}
}
