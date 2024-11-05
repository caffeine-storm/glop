package gui

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/draw"
	"io"
	"log/slog"
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

type Dimser interface {
	Dims() Dims
}

type ConstDimser struct {
	Value Dims
}

func (c *ConstDimser) Dims() Dims {
	return c.Value
}

type Dictionary struct {
	Data dictData

	logger *slog.Logger

	// TODO(tmckee): we need to call gl.DeleteTexture on this to clean up
	// properly.
	texture gl.Texture

	dims Dimser

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
	// TODO(tmckee): don't export this field; we don't need to include it in the
	// serialization if we remember to rebuild it on load. I don't want to try to
	// hide it now becuase that might break decoding old .gob files.
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

// Holds the metadata of a glyph in a 'grid-of-glyphs' texture. This metadata
// is used to blit one of the glyphs while rendering a string of text.
// E.g. A runeInfo for 'j' could look like
//
//	runeInfo{
//	  Pos: (391,15)-(398,34),
//	  Bounds: (-3,-15)-(4,4),
//	  Advance: 5.5390625,
//	}
//
// - 'Pos' defines the sub-image at lower-left corner of (391,15) and an
// upper-right corner of (398,34). These co-ordinates are relative to the
// entire 'grid-of-glyphs' texture.
// - 'Bounds' encodes that, when blitting, take the texels at 'Pos' and write
// them to the rectangle (-3,-15)-(4,4). The written rectangle is defined with
// respect to a "current raster position". Practically, this means that the 'j'
// can set pixels further left or further down than the 'current raster
// position'.
// - 'Advance' sets a distance to advance the 'current raster position', after
// blitting this glyph. This Advance does not account for kerning.
//
// In our example, the bottom-most and left-most texel of a 'j' will be drawn
// below and to the left ofthe current raster position.
//
// Note: Each sub-image includes a 1-texel, transparent border around 'real
// texels'. This ensures texture sampling won't mistakenly blend texels from
// adjancent glyphs. It means, however, that a 'no-adjustment' 'Bounds' value
// is (unintuitively) (-1,-1).
//
// TODO(tmckee): we don't need a rectangle to encode the adjustment; just a
// Point.
type runeInfo struct {
	Pos     image.Rectangle
	Bounds  image.Rectangle
	Advance float64
}

func (r *runeInfo) String() string {
	return fmt.Sprintf("Pos: %+v, Bounds: %+v, Advance: %f", r.Pos, r.Bounds, r.Advance)
}

// Stores data for blitting from an underlying texture to the screen.
type blitBuffer struct {
	// TODO(tmckee): we need to call gl.DeleteBuffer on this to clean up
	// properly.
	vertexBuffer gl.Buffer
	vertexData   []blitVertex

	// TODO(tmckee): we need to call gl.DeleteBuffer on this to clean up
	// properly.
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

func (d *Dictionary) MaxHeight() int {
	res := d.Data.Maxy - d.Data.Miny
	if res < 0 {
		res = 0
	}
	return res
}

func (d *Dictionary) split(s string, lineWidth int) []string {
	var lines []string
	var line []rune
	var word []rune
	pos := 0.0 // Sub-pixel precision
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
		if pos >= float64(lineWidth) {
			pos = 0.0
			for _, r := range word {
				pos += d.getInfo(r).Advance
			}
			lines = append(lines, string(line))
			line = line[0:0]
		}
	}
	if pos < float64(lineWidth) {
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

// TODO: This isn't working - not being tested yet
func (d *Dictionary) RenderParagraph(s string, x, y, boundingWidth int, lineHeight int, halign, valign Justification) {
	lines := d.split(s, boundingWidth)
	total_height := lineHeight * len(lines)
	switch valign {
	case Bottom:
		y += total_height
	case Center:
		y += total_height / 2
	}
	for _, line := range lines {
		d.RenderString(line, Point{X: x, Y: y}, lineHeight, halign)
		y -= lineHeight
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

func (d *Dictionary) getScreenDimensions() Dims {
	return d.dims.Dims()
}

// TODO(tmckee): refactor uses of Dictionary to not require calling
// RenderString/RenderParagraph from a render queue but dispatch the op
// internally.
// Renders the string 's' with its lower-left corner at the given position with
// the given height. Values are in units of pixels w.r.t. an origin at the
// top-left of the screen and a screen size as given.
func (d *Dictionary) RenderString(s string, target Point, height int, just Justification) {
	d.logger.Debug("RenderString called", "s", s, "target", target, "height", height, "just", just)

	if len(s) == 0 {
		return
	}

	stride := unsafe.Sizeof(blitVertex{})
	width_texunits := d.StringPixelWidth(s)

	d.logger.Debug("sizes", "stride", stride, "width", width_texunits)
	d.logger.Debug("dict-dims", "Dx", d.Data.Dx, "Dy", d.Data.Dy)

	screenDims := d.getScreenDimensions()
	screenPixelWidth, screenPixelHeight := screenDims.Dx, screenDims.Dy

	width_pixels_to_geounits := 2.0 / float64(screenPixelWidth)
	height_pixels_to_geounits := 2.0 / float64(screenPixelHeight)
	d.logger.Debug("screenpix", "width", screenPixelWidth, "height", screenPixelHeight)

	// To convert from screen space to NDC, move the origin half a screen in the
	// positive direction, then scale distances by a factor of
	// width-of-screen-in-pixels == 2.0 NDC
	x_pos_geounits := float64(target.X-screenPixelWidth/2) * width_pixels_to_geounits
	y_pos_geounits := -float64(target.Y-(screenPixelHeight/2)) * height_pixels_to_geounits

	height_geounits := float64(height) * height_pixels_to_geounits
	// TODO(tmckee): hardcoded to dict_10.gob for now :(
	height_texunits := float64(32)

	width_texunits_to_pixels := float64(1.0)
	width_texunits_to_geounits := width_texunits_to_pixels * width_pixels_to_geounits

	string_width_geounits := width_texunits * width_texunits_to_geounits
	padding_geounits := (1.0 - string_width_geounits)

	d.logger.Debug("widths", "padding_geounits", padding_geounits, "string_width_geounits", string_width_geounits, "width_texunits", width_texunits, "height_texunits", height_texunits)
	switch just {
	case Center:
		// TODO(tmckee): we shouldn't add/substract things that have different units
		x_pos_geounits += padding_geounits / 2
	case Right:
		// TODO(tmckee): we shouldn't add/substract things that have different units
		x_pos_geounits += padding_geounits
	}

	blittingData, ok := d.stringBlittingCache[s]
	if !ok {
		// We have to actually render a string!
		var prev rune
		for _, r := range s {
			// TODO(tmckee): why toss out the mapped value, then look it up again?!
			if _, ok := d.Data.Kerning[prev]; ok {
				// TODO(tmckee): XXX: !!!: no, this has to scale; Kerning adjustments
				// are in 'natural' widths... right?
				x_pos_geounits += float64(d.Data.Kerning[prev][r])
			}
			prev = r
			info := d.getInfo(r)
			xleft_geounits := x_pos_geounits
			xright_geounits := x_pos_geounits + float64(info.Bounds.Dx()-2)*width_texunits_to_geounits
			ytop_geounits := float32(y_pos_geounits + height_geounits)
			ybot_geounits := float32(y_pos_geounits)
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
				x: float32(xleft_geounits),
				y: ytop_geounits,
				u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
			})
			blittingData.vertexData = append(blittingData.vertexData, blitVertex{
				x: float32(xleft_geounits),
				y: ybot_geounits,
				u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
			})
			blittingData.vertexData = append(blittingData.vertexData, blitVertex{
				x: float32(xright_geounits),
				y: ybot_geounits,
				u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
			})
			blittingData.vertexData = append(blittingData.vertexData, blitVertex{
				x: float32(xright_geounits),
				y: ytop_geounits,
				u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
			})
			d.logger.Debug("render-char", "x_pos", x_pos_geounits, "rune", string(r), "runeInfo", info, "geometry", blittingData.vertexData[start:])
			x_pos_geounits += info.Advance * width_texunits_to_geounits

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

	d.logger.Debug("renderstring blittingData", "todraw", s, "data", blittingData)

	err := render.EnableShader("glop.font")
	if err != nil {
		panic(err)
	}
	defer render.EnableShader("")

	// TODO(tmckee): 'diff' was used for configuring a clamping function
	// (smoothstep) in the shader. The math is broken, though, and alyways comes
	// out to something that then gets clamped to 0.45
	// diff := 20/math.Pow(height, 1.0) + 5*math.Pow(d.Data.Scale, 1.0)/math.Pow(height, 1.0)
	// if diff > 0.45 {
	// diff = 0.45
	// }
	diff := 0.45
	d.logger.Debug("RenderStringDiff", "diff", diff)
	render.SetUniformF("glop.font", "dist_min", float32(0.5-diff))
	render.SetUniformF("glop.font", "dist_max", float32(0.5+diff))

	// We want to use the 0'th texture unit.
	render.SetUniformI("glop.font", "tex", 0)

	debug.LogAndClearGlErrors(d.logger)

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
	return float64(n/256) + float64(n%256)/256.0
}

func MakeDictionary(font *truetype.Font, size int, renderQueue render.RenderQueueInterface, dimser Dimser, logger *slog.Logger) *Dictionary {
	alphabet := " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*([]{};:'\",.<>/?\\|`~-_=+"
	context := freetype.NewContext()
	context.SetFont(font)
	width := 300
	height := 300
	context.SetSrc(image.White)
	dpi := 150.0
	context.SetFontSize(float64(size))
	context.SetDPI(dpi)

	// Use a glyph packing scheme. Each glyph gets a cell in a grid.
	// - Each cell is only large enough to contain all the texels of that glyph
	// plus a one-texel wide border. The 1-texel border can be shared by separate
	// glyphs.
	// - Each cell has an associated 'offset' so that the contained glyph can be
	// adjusted w.r.t. the baseline and the current raster position. That way,
	// something like a 'j' can get moved down or a '^' can get moved up and we
	// don't need a bunch of spacer texels loaded into the GPU.
	// - We also need to store the 'advance' to know how far to move the raster
	// position. The advance does not account for kerning.

	var letters []image.Image
	rune_mapping := make(map[rune]image.Image)
	rune_info := make(map[rune]runeInfo)
	for _, r := range alphabet {
		canvas := image.NewRGBA(image.Rect(-width/2, -height/2, width/2, height/2))
		context.SetDst(canvas)
		context.SetClip(canvas.Bounds())

		advance, _ := context.DrawString(string([]rune{r}), raster.Point{})
		sub := MinimalSubImage(canvas)
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

	dict.dims = dimser
	dict.logger = logger

	dict.uploadGlyphTexture(renderQueue)

	return &dict
}

func LoadDictionary(r io.Reader, renderQueue render.RenderQueueInterface, dimser Dimser, logger *slog.Logger) (*Dictionary, error) {
	var d Dictionary
	err := d.Load(r)
	if err != nil {
		return nil, err
	}
	d.dims = dimser
	d.logger = logger
	d.uploadGlyphTexture(renderQueue)
	return &d, nil
}

func (d *Dictionary) Load(inputStream io.Reader) error {
	return gob.NewDecoder(inputStream).Decode(&d.Data)
}

func (d *Dictionary) Store(outputStream io.Writer) error {
	return gob.NewEncoder(outputStream).Encode(d.Data)
}

func (d *Dictionary) uploadGlyphTexture(renderQueue render.RenderQueueInterface) {
	d.stringBlittingCache = make(map[string]blitBuffer)
	d.paragraphBlittingCache = make(map[string]blitBuffer)

	renderQueue.Queue(func() {
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
