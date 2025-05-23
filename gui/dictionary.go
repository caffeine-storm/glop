package gui

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/draw"
	"io"
	"unsafe"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
)

// Shader stuff - The font stuff requires that we use some simple shaders
// TODO(tmckee): if we enable depth testing, do we also need to define a 'z'
// coordinate for the geometry? ought to be able to just assign 0 to the
// z-component of ... something?
const font_vertex_shader string = `
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

const font_fragment_shader string = `
  #version 120
  uniform sampler2D tex;

  void main() {
    vec2 tpos = gl_TexCoord[0].xy;
    float dist = texture2D(tex, tpos).a;
    float alpha = smoothstep(0.05, 0.95, dist);
    gl_FragColor = gl_Color * vec4(1.0, 1.0, 1.0, alpha);
  }
`

type Justification int

const (
	Center Justification = iota
	Left
	Right
	Top
	Bottom
)

type Dictionary struct {
	Data RasteredFont

	logger glog.Logger

	// TODO(tmckee): we need to call gl.DeleteTexture on this to clean up
	// properly.
	texture gl.Texture

	stringBlittingCache    map[string]blitBuffer
	paragraphBlittingCache map[string]blitBuffer
}

type RasteredFont struct {
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

	// runeInfo for all r < 256 will be stored here as well as in Info so we can
	// avoid map lookups if possible. Not stored on disk but reconstructed when
	// deserializing.
	asciiInfo []runeInfo

	// At what vertical value is the line on which text is logically rendered.
	// This is determined by the positioning of the '.' rune.
	Baseline int

	// The lowest and highest relative pixel position amongst the glyphs.
	Miny, Maxy int
}

func (d *RasteredFont) rebuildAsciiInfo() {
	d.asciiInfo = make([]runeInfo, 256)
	for r := rune(0); r < 256; r++ {
		if info, ok := d.Info[r]; ok {
			d.asciiInfo[r] = info
		}
	}
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

const stride = int(unsafe.Sizeof(blitVertex{}))

func (d *Dictionary) getInfo(r rune) runeInfo {
	return d.Data.getInfo(r)
}

func (font *RasteredFont) getInfo(r rune) runeInfo {
	var info runeInfo
	if r >= 0 && r < 256 {
		info = font.asciiInfo[r]
	} else {
		info, _ = font.Info[r]
	}
	return info
}

func (d *Dictionary) MaxHeight() int {
	return d.Data.MaxHeight()
}

func (font *RasteredFont) MaxHeight() int {
	res := font.Maxy - font.Miny
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
func (d *Dictionary) RenderParagraph(s string, x, y, boundingWidth int, lineHeight int, halign, valign Justification, shaders *render.ShaderBank) {
	lines := d.split(s, boundingWidth)
	total_height := lineHeight * len(lines)
	switch valign {
	case Bottom:
		y += total_height
	case Center:
		y += total_height / 2
	}
	for _, line := range lines {
		d.RenderString(line, Point{X: x, Y: y}, lineHeight, halign, shaders)
		y -= lineHeight
	}
}

// Figures out how wide a string will be if rendered at its natural size.
func (d *Dictionary) StringPixelWidth(s string) float64 {
	return d.Data.StringPixelWidth(s)
}

func (font *RasteredFont) StringPixelWidth(s string) float64 {
	width := 0.0
	var prev rune
	for _, r := range s {
		info := font.getInfo(r)
		width += info.Advance

		// Need to account for kerning adjustments
		if kernData, ok := font.Kerning[prev]; ok {
			width += float64(kernData[r])
		}
		prev = r
	}
	return width
}

func buildBlittingData(s string, d *Dictionary, x_pos_px, y_pos_px, height_px float64) blitBuffer {
	blittingData := blitBuffer{}
	var prev rune
	verticalScale := height_px / float64(d.MaxHeight())
	horizontalScale := verticalScale
	for _, r := range s {
		if kernAdjustment, ok := d.Data.Kerning[prev]; ok {
			x_pos_px += float64(kernAdjustment[r])
		}
		prev = r
		info := d.getInfo(r)
		xleft_px := x_pos_px
		xright_px := x_pos_px + float64(info.Bounds.Dx())*horizontalScale
		ybot_px := float32(y_pos_px)
		ytop_px := float32(y_pos_px + height_px)
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
			x: float32(xleft_px),
			y: ytop_px,
			u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
			v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
		})
		blittingData.vertexData = append(blittingData.vertexData, blitVertex{
			x: float32(xleft_px),
			y: ybot_px,
			u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
			v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
		})
		blittingData.vertexData = append(blittingData.vertexData, blitVertex{
			x: float32(xright_px),
			y: ybot_px,
			u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
			v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
		})
		blittingData.vertexData = append(blittingData.vertexData, blitVertex{
			x: float32(xright_px),
			y: ytop_px,
			u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
			v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
		})
		d.logger.Trace("render-char", "x_pos", x_pos_px, "rune", string(r), "runeInfo", info, "geometry", blittingData.vertexData[start:])
		x_pos_px += info.Advance * horizontalScale
	}

	d.logger.Trace("geometry", "verts", blittingData.vertexData, "idxs", blittingData.indicesData)
	blittingData.vertexBuffer = gl.GenBuffer()
	blittingData.vertexBuffer.Bind(gl.ARRAY_BUFFER)
	defer gl.Buffer(0).Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, stride*len(blittingData.vertexData), blittingData.vertexData, gl.STATIC_DRAW)

	blittingData.indicesBuffer = gl.GenBuffer()
	blittingData.indicesBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	defer gl.Buffer(0).Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(unsafe.Sizeof(blittingData.indicesData[0]))*len(blittingData.indicesData), blittingData.indicesData, gl.STATIC_DRAW)

	return blittingData
}

// Renders the string 's' at the given position with the given height. Values
// are in units of pixels w.r.t. an origin at the bottom-left of the screen.
// The text is positioned based on the given justification:
//
//	Left: use 'target.X' for the left-hand extent of what's drawn
//	Centre: use 'target.X' for the middle of what's drawn
//	Right: use 'target.X' for the right-hand extent of what's drawn
//
// The bottom of the text's bounding box is aligned to the Y co-ordinate of the
// target.
func (d *Dictionary) RenderString(s string, target Point, height int, just Justification, shaders *render.ShaderBank) {
	d.logger.Trace("RenderString called", "s", s, "target", target, "height", height, "just", just)

	if d.texture == 0 {
		panic(fmt.Errorf("can't RenderString for uninitialized Dictionary"))
	}

	if len(s) == 0 {
		return
	}

	string_width_px := d.StringPixelWidth(s)

	d.logger.Trace("sizes", "width", string_width_px, "d.Data.Dx", d.Data.Dx, "d.Data.Dy", d.Data.Dy)

	x_pos_px := float64(target.X)
	y_pos_px := float64(target.Y)

	height_px := float64(height)

	switch just {
	case Center:
		x_pos_px -= string_width_px / 2
	case Right:
		x_pos_px -= string_width_px
	}

	blittingData, ok := d.stringBlittingCache[s]
	if !ok {
		blittingData = buildBlittingData(s, d, x_pos_px, y_pos_px, height_px)
		d.stringBlittingCache[s] = blittingData
	}

	d.logger.Trace("renderstring blittingData", "todraw", s, "data", blittingData)

	err := shaders.EnableShader("glop.font")
	if err != nil {
		panic(err)
	}
	defer shaders.EnableShader("")

	// We want to use the 0'th texture unit.
	shaders.SetUniformI("glop.font", "tex", 0)

	debug.LogAndClearGlErrors(d.logger)

	gl.PushAttrib(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	defer gl.PopAttrib()
	gl.Enable(gl.BLEND)
	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ZERO, gl.ONE)
	gl.Disable(gl.DEPTH_TEST)

	d.texture.Bind(gl.TEXTURE_2D)
	defer gl.Texture(0).Bind(gl.TEXTURE_2D)

	gl.EnableClientState(gl.VERTEX_ARRAY)
	defer gl.DisableClientState(gl.VERTEX_ARRAY)
	blittingData.vertexBuffer.Bind(gl.ARRAY_BUFFER)
	defer gl.Buffer(0).Bind(gl.ARRAY_BUFFER)
	gl.VertexPointer(2, gl.FLOAT, stride, nil)

	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	blittingData.indicesBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	defer gl.Buffer(0).Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.TexCoordPointer(2, gl.FLOAT, stride, unsafe.Offsetof(blittingData.vertexData[0].u))

	gl.DrawElements(gl.TRIANGLES, len(blittingData.indicesData), gl.UNSIGNED_SHORT, nil)

	debug.LogAndClearGlErrors(d.logger)
}

func fix24_8_to_float64(n raster.Fix32) float64 {
	// 'n' is a fractional value packed into an int32 with the 24
	// most-significant bits representing the 'whole' portion and the 8
	// least-significant bits representing the fractional part.
	return float64(n/256) + float64(n%256)/256.0
}

func RasterizeFont(font *truetype.Font, pointSize int) RasteredFont {
	alphabet := " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*([]{};:'\",.<>/?\\|`~-_=+"
	context := freetype.NewContext()
	context.SetFont(font)
	width := 300
	height := 300
	context.SetSrc(image.White)
	dpi := 150.0
	context.SetFontSize(float64(pointSize))
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

	var result RasteredFont
	result.Pix = pim.Pix
	result.Dx = pim.Bounds().Dx()
	result.Dy = pim.Bounds().Dy()
	result.Info = rune_info

	result.rebuildAsciiInfo()
	result.Baseline = result.Info['.'].Bounds.Min.Y

	result.Miny = int(1e9)
	result.Maxy = int(-1e9)
	for _, info := range result.Info {
		if info.Bounds.Min.Y < result.Miny {
			result.Miny = info.Bounds.Min.Y
		}
		if info.Bounds.Max.Y > result.Maxy {
			result.Maxy = info.Bounds.Max.Y
		}
	}

	return result
}

func MakeAndInitializeDictionary(font *truetype.Font, size int, renderQueue render.RenderQueueInterface, logger glog.Logger) *Dictionary {
	dict := Dictionary{
		Data:                   RasterizeFont(font, size),
		logger:                 logger,
		stringBlittingCache:    map[string]blitBuffer{},
		paragraphBlittingCache: map[string]blitBuffer{},
	}

	dict.initialize(renderQueue)
	return &dict
}

func LoadAndInitializeDictionary(r io.Reader, renderQueue render.RenderQueueInterface, logger glog.Logger) (*Dictionary, error) {
	d := Dictionary{
		stringBlittingCache:    map[string]blitBuffer{},
		paragraphBlittingCache: map[string]blitBuffer{},
	}
	err := d.Load(r)
	if err != nil {
		return nil, err
	}
	d.logger = logger

	d.initialize(renderQueue)
	return &d, nil
}

func (d *Dictionary) Load(inputStream io.Reader) error {
	if err := gob.NewDecoder(inputStream).Decode(&d.Data); err != nil {
		return err
	}

	d.Data.rebuildAsciiInfo()

	return nil
}

func (d *Dictionary) Store(outputStream io.Writer) error {
	return gob.NewEncoder(outputStream).Encode(d.Data)
}

func (d *Dictionary) initialize(renderQueue render.RenderQueueInterface) {
	d.compileShaders("glop.font", renderQueue)
	d.uploadGlyphTexture(renderQueue)
}

func (d *Dictionary) compileShaders(shaderName string, renderQueue render.RenderQueueInterface) {
	renderQueue.Queue(func(st render.RenderQueueState) {
		if st.Shaders().HasShader(shaderName) {
			return
		}

		err := st.Shaders().RegisterShader(shaderName, font_vertex_shader, font_fragment_shader)
		if err != nil {
			panic(fmt.Errorf("failed to register font %q: %w", shaderName, err))
		}
	})
}

func (d *Dictionary) uploadGlyphTexture(renderQueue render.RenderQueueInterface) {

	renderQueue.Queue(func(render.RenderQueueState) {
		d.texture = gl.GenTexture()
		d.texture.Bind(gl.TEXTURE_2D)
		defer gl.Texture(0).Bind(gl.TEXTURE_2D)
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
			// We use unsigned int here to treat each group of 4 bytes like one big
			// alpha value. Yes, that means we're interpreting red, green and blue
			// components as part of the alpha but, since all of the texture should
			// be grayscale, we can cut this corner.
			gl.UNSIGNED_INT,
			d.Data.Pix)
	})
}
