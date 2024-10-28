package gui

import (
	"encoding/gob"
	"image"
	"image/draw"
	"io"
	"log/slog"
	"math"
	"unsafe"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
)

type Justification int

const (
	Center Justification = iota
	Left
	Right
	Top
	Bottom
)

type Dictionary struct {
	Data dictData

	renderQueue render.RenderQueueInterface
	logger      *slog.Logger

	texture gl.Texture

	stringBlittingCache    map[string]blitBuffer
	paragraphBlittingCache map[string]blitBuffer
}

type dictData struct {
	// The Pix data of an image.RGBA of the packed 'grid of glyphs'.
	Pix []byte

	// Adjustments to apply to the rastering position while rendering strings of
	// text. For nice-to-read text, we need to support pushing some character
	// pairs closer together while pushing others apart. This adjustment is known
	// as kerning.  This is, effectively, a mapping of ordered character pairs to
	// an adjustment to apply between them.
	// TODO(tmckee): this is all zero right now, AFAICT, and should be a float
	// instead of an int... right!?
	Kerning map[rune]map[rune]int

	// Width and height in pixels of Pix's image.RGBA.
	Dx, Dy int

	// Map from rune to that rune's runeInfo.
	Info map[rune]runeInfo

	// runeInfo for all r < 256 will be stored here as well as in info so we can
	// avoid map lookups if possible.
	Ascii_info []runeInfo

	// At what vertical value is the line on which text is logically rendered.
	// This is determined by the positioning of the '.' rune.
	Baseline int

	// Amount glyphs were scaled down during packing.
	// TODO(tmckee): this doesn't seem to be present in old .gob files and nobody
	// seems to read it; we should remove it.
	Scale float64

	// The lowest and highest relative pixel position amongst the glyphs.
	Miny, Maxy int
}

// Describes the location and size of a glyph in a 'grid-of-glyphs' texture
// that has been 'packed'.
type runeInfo struct {
	// Texture's minimal sub-image of the glyph's texels.
	Pos image.Rectangle

	// Padded sub-image of the glyph's texels; like above but positioned relative
	// to the glyph. Used to include texels that the glyph does not own in the
	// texture but should be 'assumed' blank in the texture. That is, the texels
	// might not be blank in the texture because of tight packing but drawing the
	// character should operate as if they were.
	// TODO(tmckee): does that make sense? Is that what this acutally _is_?
	Bounds image.Rectangle

	// TODO(tmckee): Full_bounds seems to never get populated... presumably this
	// is the bounds to allocate and blit the character into. We will need to
	// 'MakeDictionary' with real .ttf fonts to remake the pre-packed files
	// because we need these dimensions.
	Full_bounds image.Rectangle

	// How far to move the rastering position to the right in natural pixels
	// after having rendered the corresponding rune. Does not account for
	// kerning.
	Advance float64
}

// Stores data for blitting from an underlying texture to the screen.
type blitBuffer struct {
	vertexBuffer gl.Buffer
	vertexData   []blitVertex

	indicesBuffer gl.Buffer
	indicesData   []uint16
}

// Stores indivdual vertex data for our blitting operations. (x, y) denotes a
// point on screen. (u, v) denotes a point in the texture.
type blitVertex struct {
	x, y float32
	u, v float32
}

func (d *Dictionary) Scale() float64 {
	return d.Data.Scale
}

func (d *Dictionary) getInfo(r rune) runeInfo {
	var info runeInfo
	if r >= 0 && r < 256 {
		info = d.Data.Ascii_info[r]
	} else {
		info, _ = d.Data.Info[r]
	}
	return info
}

func (d *Dictionary) MaxHeight() float64 {
	res := d.Data.Maxy - d.Data.Miny
	if res < 0 {
		res = 0
	}
	return float64(res)
}

func (d *Dictionary) split(s string, dx, height float64) []string {
	var lines []string
	var line []rune
	var word []rune
	pos := 0.0
	for _, r := range s {
		if r == ' ' {
			if len(line) > 0 {
				line = append(line, ' ')
			}
			for _, r := range word {
				line = append(line, r)
			}
			word = word[0:0]
		} else {
			word = append(word, r)
		}
		pos += d.getInfo(r).Advance
		if pos >= dx {
			pos = 0.0
			for _, r := range word {
				pos += d.getInfo(r).Advance
			}
			lines = append(lines, string(line))
			line = line[0:0]
		}
	}
	if pos < dx {
		if len(line) > 0 {
			line = append(line, ' ')
		}
		for _, r := range word {
			line = append(line, r)
		}
		lines = append(lines, string(line))
	} else {
		lines = append(lines, string(line))
		lines = append(lines, string(word))
	}
	return lines
}

// TODO: This isn't working - not even a little
func (d *Dictionary) RenderParagraph(s string, x, y, z, dx, height float64, halign, valign Justification) {
	lines := d.split(s, dx, height)
	total_height := height * float64(len(lines)-1)
	switch valign {
	case Bottom:
		y += total_height
	case Center:
		y += total_height / 2
	}
	for _, line := range lines {
		d.RenderString(line, x, y, z, height, halign)
		y -= height
	}
}

// Figures out how wide a string will be if rendered at its natural size.
func (d *Dictionary) StringPixelWidth(s string) float64 {
	width := 0.0
	var prev rune
	for _, r := range s {
		info := d.getInfo(r)
		width += info.Advance

		// Need to account for kerning adjustments
		if kernData, ok := d.Data.Kerning[prev]; ok {
			width += float64(kernData[r])
		}
		prev = r
	}
	return width
}

// TODO(tmckee): refactor uses of Dictionary to not require calling
// RenderString/RenderParagraph from a render queue but dispatch the op
// internally.
// Renders the string 's' at the given (x, y, z) in normalized device co-ordinates and at a height of 'height' in pixels.
func (d *Dictionary) RenderString(s string, x, y, z, height float64, just Justification) {
	d.logger.Debug("RenderString called", "s", s, "x", x, "y", y, "z", z, "height", height, "just", just)
	debug.LogAndClearGlErrors(d.logger)

	if len(s) == 0 {
		return
	}

	stride := unsafe.Sizeof(blitVertex{})
	// TODO(tmckee): d.data.Maxy-d.data.Miny is d.MaxHeight() ... need to DRY
	// this out.
	scale := height / float64(d.Data.Maxy-d.Data.Miny)
	width_texunits := float32(d.StringPixelWidth(s) * scale)
	d.logger.Debug("sizes", "scale", scale, "stride", stride, "width", width_texunits)
	d.logger.Debug("dict-dims", "Dx", d.Data.Dx, "Dy", d.Data.Dy)
	x_pos_geounits := float32(x)

	height_geounits := 1.0
	// TODO(tmckee): gaaah! Dy should not be glyph height!!!
	height_texunits := float64(d.Data.Dy)
	texunits_to_geounits := float32(height_geounits / height_texunits)

	switch just {
	case Center:
		// TODO(tmckee): we shouldn't add/substract things that have different units
		x_pos_geounits -= width_texunits / 2
	case Right:
		// TODO(tmckee): we shouldn't add/substract things that have different units
		x_pos_geounits -= width_texunits
	}

	blittingData, ok := d.stringBlittingCache[s]
	if !ok {
		// We have to actually render a string!
		x_pos_geounits = 0
		var prev rune
		for _, r := range s {
			// TODO(tmckee): why toss out the mapped value, then look it up again?!
			if _, ok := d.Data.Kerning[prev]; ok {
				x_pos_geounits += float32(d.Data.Kerning[prev][r])
			}
			prev = r
			info := d.getInfo(r)
			d.logger.Debug("render-char", "x_pos", x_pos_geounits, "rune", r, "runeInfo", info, "dict-maxy", d.Data.Maxy)
			xleft_geounits := x_pos_geounits
			xright_geounits := x_pos_geounits + float32(info.Pos.Max.X-info.Pos.Min.X)*texunits_to_geounits
			// TODO(tmckee): uh... what? shouldn't it just be ytop_geounits, ybot := scale, 0 ?
			ytop_geounits := float32(1.0)
			ybot_geounits := float32(0.0)
			start := uint16(len(blittingData.vertexData))
			blittingData.indicesData = append(blittingData.indicesData, start+0)
			blittingData.indicesData = append(blittingData.indicesData, start+1)
			blittingData.indicesData = append(blittingData.indicesData, start+2)
			blittingData.indicesData = append(blittingData.indicesData, start+0)
			blittingData.indicesData = append(blittingData.indicesData, start+2)
			blittingData.indicesData = append(blittingData.indicesData, start+3)

			// Note: the texture is loaded 'upside down' so we flip our y-coordinates
			// in texture-space.
			blittingData.vertexData = append(blittingData.vertexData, blitVertex{
				x: xleft_geounits,
				y: ytop_geounits,
				u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
			})
			blittingData.vertexData = append(blittingData.vertexData, blitVertex{
				x: xleft_geounits,
				y: ybot_geounits,
				u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
			})
			blittingData.vertexData = append(blittingData.vertexData, blitVertex{
				x: xright_geounits,
				y: ybot_geounits,
				u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
			})
			blittingData.vertexData = append(blittingData.vertexData, blitVertex{
				x: xright_geounits,
				y: ytop_geounits,
				u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
			})
			x_pos_geounits += float32(info.Advance) * texunits_to_geounits
		}

		d.logger.Debug("geometry", "verts", blittingData.vertexData, "idxs", blittingData.indicesData)
		blittingData.vertexBuffer = gl.GenBuffer()
		blittingData.vertexBuffer.Bind(gl.ARRAY_BUFFER)
		gl.BufferData(gl.ARRAY_BUFFER, int(stride)*len(blittingData.vertexData), blittingData.vertexData, gl.STATIC_DRAW)

		blittingData.indicesBuffer = gl.GenBuffer()
		blittingData.indicesBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(unsafe.Sizeof(blittingData.indicesData[0]))*len(blittingData.indicesData), blittingData.indicesData, gl.STATIC_DRAW)
		d.stringBlittingCache[s] = blittingData
	}

	// Reset x-pos
	x_pos_geounits = float32(x)
	switch just {
	case Center:
		x_pos_geounits -= width_texunits / 2
	case Right:
		x_pos_geounits -= width_texunits
	}

	debug.LogAndClearGlErrors(d.logger)

	err := render.EnableShader("glop.font")
	if err != nil {
		panic(err)
	}
	defer render.EnableShader("")

	debug.LogAndClearGlErrors(d.logger)

	diff := 20/math.Pow(height, 1.0) + 5*math.Pow(d.Data.Scale, 1.0)/math.Pow(height, 1.0)
	if diff > 0.45 {
		diff = 0.45
	}
	d.logger.Debug("RenderStringDiff", "diff", diff)
	render.SetUniformF("glop.font", "dist_min", float32(0.5-diff))
	render.SetUniformF("glop.font", "dist_max", float32(0.5+diff))

	debug.LogAndClearGlErrors(d.logger)

	// We want to use the 0'th texture unit.
	// TODO(tmckee): this seems to be getting an 'INVALID_VALUE' glerror back.
	// Look into whether we've correctly looked up the uniform location.
	render.SetUniformI("glop.font", "tex", gl.TEXTURE0+0)

	debug.LogAndClearGlErrors(d.logger)

	{
		d.logger.Debug("matrixmode", "mode", debug.GetMatrixMode())

		x, y, w, h := debug.GetViewport()
		d.logger.Debug("viewport", "x", x, "y", y, "w", w, "h", h)

		near, far := debug.GetDepthRange()
		d.logger.Debug("depth", "near", near, "far", far)
	}

	gl.PushAttrib(gl.COLOR_BUFFER_BIT)
	defer gl.PopAttrib()
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// TODO(tmckee): we should do error checking with glGetError:
	// https://docs.gl/gl2/glGetError
	// TODO(tmckee): This seems specific to OpenGL2/2.1: https://docs.gl/gl2/glEnable
	gl.Enable(gl.TEXTURE_2D)
	d.texture.Bind(gl.TEXTURE_2D)

	gl.EnableClientState(gl.VERTEX_ARRAY)
	defer gl.DisableClientState(gl.VERTEX_ARRAY)
	blittingData.vertexBuffer.Bind(gl.ARRAY_BUFFER)
	gl.VertexPointer(2, gl.FLOAT, int(stride), nil)

	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	blittingData.indicesBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.TexCoordPointer(2, gl.FLOAT, int(stride), unsafe.Offsetof(blittingData.vertexData[0].u))

	gl.DrawElements(gl.TRIANGLES, len(blittingData.indicesData), gl.UNSIGNED_SHORT, nil)

	debug.LogAndClearGlErrors(d.logger)
}

func fix24_8_to_float64(n raster.Fix32) float64 {
	// 'n' is a fractional value packed into an int32 with the 24
	// most-significant bits representing the 'whole' portion and the 8
	// least-significant bits representing the fractional part.
	return float64(n) / (2 ^ 8)
}

func MakeDictionary(font *truetype.Font, size int, renderQueue render.RenderQueueInterface) *Dictionary {
	alphabet := " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*([]{};:'\",.<>/?\\|`~-_=+"
	context := freetype.NewContext()
	context.SetFont(font)
	width := 300
	height := 300
	context.SetSrc(image.White)
	dpi := 150.0
	context.SetFontSize(float64(size))
	context.SetDPI(dpi)
	var letters []image.Image
	rune_mapping := make(map[rune]image.Image)
	rune_info := make(map[rune]runeInfo)
	for _, r := range alphabet {
		canvas := image.NewRGBA(image.Rect(-width/2, -height/2, width/2, height/2))
		context.SetDst(canvas)
		context.SetClip(canvas.Bounds())

		advance, _ := context.DrawString(string([]rune{r}), raster.Point{})
		sub := minimalSubImage(canvas)
		letters = append(letters, sub)
		rune_mapping[r] = sub
		adv := fix24_8_to_float64(advance.X)
		rune_info[r] = runeInfo{Bounds: sub.bounds, Advance: adv}
	}
	packed := packImages(letters)

	for _, r := range alphabet {
		ri := rune_info[r]
		ri.Pos = packed.GetPackedLocation(rune_mapping[r])
		rune_info[r] = ri
	}

	dx := 1
	for dx < packed.Bounds().Dx() {
		dx = dx << 1
	}
	dy := 1
	for dy < packed.Bounds().Dy() {
		dy = dy << 1
	}

	pim := image.NewRGBA(image.Rect(0, 0, dx, dy))
	draw.Draw(pim, pim.Bounds(), packed, image.Point{}, draw.Src)
	var dict Dictionary
	dict.Data.Pix = pim.Pix
	// TODO(tmckee): Dy will be two glyphs tall in some cases; Dy should _not_
	// get used as the height of a glyph!!!
	dict.Data.Dx = pim.Bounds().Dx()
	dict.Data.Dy = pim.Bounds().Dy()
	dict.Data.Info = rune_info

	dict.Data.Ascii_info = make([]runeInfo, 256)
	for r := rune(0); r < 256; r++ {
		if info, ok := dict.Data.Info[r]; ok {
			dict.Data.Ascii_info[r] = info
		}
	}
	dict.Data.Baseline = dict.Data.Info['.'].Bounds.Min.Y

	dict.Data.Miny = int(1e9)
	dict.Data.Maxy = int(-1e9)
	for _, info := range dict.Data.Info {
		if info.Bounds.Min.Y < dict.Data.Miny {
			dict.Data.Miny = info.Bounds.Min.Y
		}
		if info.Bounds.Max.Y > dict.Data.Maxy {
			dict.Data.Maxy = info.Bounds.Max.Y
		}
	}

	dict.renderQueue = renderQueue
	dict.setupGlStuff()

	return &dict
}

func LoadDictionary(r io.Reader, renderQueue render.RenderQueueInterface, logger *slog.Logger) (*Dictionary, error) {
	var d Dictionary
	err := gob.NewDecoder(r).Decode(&d.Data)
	if err != nil {
		return nil, err
	}
	d.renderQueue = renderQueue
	d.logger = logger
	d.setupGlStuff()
	return &d, nil
}

func (d *Dictionary) Store(outputStream io.Writer) error {
	return gob.NewEncoder(outputStream).Encode(d.Data)
}

// Sets up anything that wouldn't have been loaded from disk, including
// all opengl data.
func (d *Dictionary) setupGlStuff() {
	d.stringBlittingCache = make(map[string]blitBuffer)
	d.paragraphBlittingCache = make(map[string]blitBuffer)

	d.renderQueue.Queue(func() {
		gl.Enable(gl.TEXTURE_2D)
		d.texture = gl.GenTexture()
		d.texture.Bind(gl.TEXTURE_2D)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

		gl.ActiveTexture(gl.TEXTURE0 + 0)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.ALPHA,
			d.Data.Dx,
			d.Data.Dy,
			0,
			gl.ALPHA,
			gl.UNSIGNED_INT,
			d.Data.Pix)

		gl.Disable(gl.TEXTURE_2D)
	})
}
