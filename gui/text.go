package gui

import (
	"encoding/gob"
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"math"
	"sort"
	"sync"
	"unsafe"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/render"
)

// Shader stuff - The font stuff requires that we use some simple shaders
// TODO(tmckee): add a #version pragma for OpenGL 2.1
const font_vertex_shader = `
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

// TODO(tmckee): add a #version pragma for OpenGL 2.1
const font_fragment_shader = `
  #version 120
  uniform sampler2D tex;
  uniform float dist_min;
  uniform float dist_max;

  void main() {
    vec2 tpos = gl_TexCoord[0].st;
    float dist = texture2D(tex, tpos).a;
    float alpha = smoothstep(dist_min, dist_max, dist);
    gl_FragColor = gl_Color * vec4(1.0, 1.0, 1.0, alpha);
  }
`

type runeInfo struct {
	Pos         image.Rectangle
	Bounds      image.Rectangle
	Full_bounds image.Rectangle
	Advance     float64
}
type dictData struct {
	// The Pix data from the original image.Rgba
	Pix []byte

	Kerning map[rune]map[rune]int

	// Dx and Dy of the original image.Rgba
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
	Scale float64

	Miny, Maxy int
}
type Dictionary struct {
	Data dictData

	// TODO(tmckee): store a gl.Texture instead of a uint32
	texture uint32

	strs map[string]strBuffer
	pars map[string]strBuffer
}
type strBuffer struct {
	// vertex-buffer
	vbuffer uint32
	vs      []dictVert

	// inidices-buffer
	ibuffer uint32
	is      []uint16
}
type dictVert struct {
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

// Figures out how wide a string will be if rendered at its natural size.
func (d *Dictionary) figureWidth(s string) float64 {
	w := 0.0
	for _, r := range s {
		w += d.getInfo(r).Advance
	}
	return w
}

type Justification int

const (
	Center Justification = iota
	Left
	Right
	Top
	Bottom
)

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

func (d *Dictionary) StringWidth(s string) float64 {
	width := 0.0
	for _, r := range s {
		info := d.getInfo(r)
		width += info.Advance
	}
	return width
}

func (d *Dictionary) RenderString(s string, x, y, z, height float64, just Justification) {
	debug.LogAndClearGlErrors(log.Default())

	if len(s) == 0 {
		return
	}

	stride := unsafe.Sizeof(dictVert{})
	// TODO(tmckee): d.data.Maxy-d.data.Miny is d.MaxHeight() ... need to DRY
	// this out.
	scale := height / float64(d.Data.Maxy-d.Data.Miny)
	width := float32(d.figureWidth(s) * scale)
	log.Printf("scale: %v, stride: %v, width: %v", scale, stride, width)
	log.Printf("d.Data.D{x,y}: %v, %v", d.Data.Dx, d.Data.Dy)
	x_pos := float32(x)
	switch just {
	case Center:
		x_pos -= width / 2
	case Right:
		x_pos -= width
	}

	strbuf, ok := d.strs[s]
	if !ok {
		// We have to actually render a string!
		x_pos = 0
		var prev rune
		for _, r := range s {
			// TODO(tmckee): why toss out the mapped value, then look it up again?!
			if _, ok := d.Data.Kerning[prev]; ok {
				x_pos += float32(d.Data.Kerning[prev][r])
			}
			prev = r
			info := d.getInfo(r)
			log.Printf("render char: x_pos: %f rune: %d\n", x_pos, r)
			log.Printf("info: %+v\n", info)
			log.Printf("d.Data.Maxy: %+v\n", d.Data.Maxy)
			xleft := x_pos + float32(info.Full_bounds.Min.X)      //- float32(info.Full_bounds.Min.X-info.Bounds.Min.X)
			xright := x_pos + float32(info.Full_bounds.Max.X)     //+ float32(info.Full_bounds.Max.X-info.Bounds.Max.X)
			ytop := float32(d.Data.Maxy - info.Full_bounds.Max.Y) //- float32(info.Full_bounds.Min.Y-info.Bounds.Min.Y)
			ybot := float32(d.Data.Maxy - info.Full_bounds.Min.Y) //+ float32(info.Full_bounds.Max.X-info.Bounds.Max.X)
			start := uint16(len(strbuf.vs))
			strbuf.is = append(strbuf.is, start+0)
			strbuf.is = append(strbuf.is, start+1)
			strbuf.is = append(strbuf.is, start+2)
			strbuf.is = append(strbuf.is, start+0)
			strbuf.is = append(strbuf.is, start+2)
			strbuf.is = append(strbuf.is, start+3)
			strbuf.vs = append(strbuf.vs, dictVert{
				x: xleft,
				y: ytop,
				u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
			})
			strbuf.vs = append(strbuf.vs, dictVert{
				x: xleft,
				y: ybot,
				u: float32(info.Pos.Min.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
			})
			strbuf.vs = append(strbuf.vs, dictVert{
				x: xright,
				y: ybot,
				u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Min.Y) / float32(d.Data.Dy),
			})
			strbuf.vs = append(strbuf.vs, dictVert{
				x: xright,
				y: ytop,
				u: float32(info.Pos.Max.X) / float32(d.Data.Dx),
				v: float32(info.Pos.Max.Y) / float32(d.Data.Dy),
			})
			x_pos += float32(info.Advance) // - float32((info.Full_bounds.Dx() - info.Bounds.Dx()))
		}

		// XXX: add a 'letter' that covers the entire viewport to show a slice of
		// the entire glyph-grid texture... plz
		numvs := uint16(len(strbuf.vs))
		strbuf.is = append(strbuf.is,
			numvs+0,
			numvs+1,
			numvs+2,
			numvs+0,
			numvs+2,
			numvs+3,
		)
		strbuf.vs = append(strbuf.vs, []dictVert{
			{
				x: -1,
				y: -1,
				u: 0,
				v: 1,
			},
			{
				x: -1,
				y: 1,
				u: 0,
				v: 0,
			},
			{
				x: 1,
				y: 1,
				u: 1,
				v: 0,
			},
			{
				x: 1,
				y: -1,
				u: 1,
				v: 1,
			},
		}...)

		log.Printf("vxs: %v", strbuf.vs)
		log.Printf("ixs: %v", strbuf.is)
		strbuf.vbuffer = uint32(gl.GenBuffer())
		gl.Buffer(strbuf.vbuffer).Bind(gl.ARRAY_BUFFER)
		gl.BufferData(gl.ARRAY_BUFFER, int(stride)*len(strbuf.vs), strbuf.vs, gl.STATIC_DRAW)

		strbuf.ibuffer = uint32(gl.GenBuffer())
		gl.Buffer(strbuf.ibuffer).Bind(gl.ELEMENT_ARRAY_BUFFER)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(unsafe.Sizeof(strbuf.is[0]))*len(strbuf.is), strbuf.is, gl.STATIC_DRAW)
		d.strs[s] = strbuf
	}

	// Reset x-pos
	x_pos = float32(x)
	switch just {
	case Center:
		x_pos -= width / 2
	case Right:
		x_pos -= width
	}

	debug.LogAndClearGlErrors(log.Default())

	err := render.EnableShader("glop.font")
	if err != nil {
		panic(err)
	}
	defer render.EnableShader("")

	debug.LogAndClearGlErrors(log.Default())

	diff := 20/math.Pow(height, 1.0) + 5*math.Pow(d.Data.Scale, 1.0)/math.Pow(height, 1.0)
	if diff > 0.45 {
		diff = 0.45
	}
	log.Printf("diff: %f", diff)
	render.SetUniformF("glop.font", "dist_min", float32(0.5-diff))
	render.SetUniformF("glop.font", "dist_max", float32(0.5+diff))

	debug.LogAndClearGlErrors(log.Default())

	// We want to use the 0'th texture unit.
	render.SetUniformI("glop.font", "tex", gl.TEXTURE0 + 0)

	debug.LogAndClearGlErrors(log.Default())

	{
		log.Printf("current matrix mode: %q", debug.GetMatrixMode())

		x, y, w, h := debug.GetViewport()
		log.Printf("current viewport: %v %v %v %v", x, y, w, h)

		near, far := debug.GetDepthRange()
		log.Printf("depth range: %v %v", near, far)
	}

	gl.PushAttrib(gl.COLOR_BUFFER_BIT)
	defer gl.PopAttrib()
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// TODO(tmckee): we should do error checking with glGetError:
	// https://docs.gl/gl2/glGetError
	// TODO(tmckee): This seems specific to OpenGL2/2.1: https://docs.gl/gl2/glEnable
	gl.Enable(gl.TEXTURE_2D)
	gl.Texture(d.texture).Bind(gl.TEXTURE_2D)

	gl.EnableClientState(gl.VERTEX_ARRAY)
	defer gl.DisableClientState(gl.VERTEX_ARRAY)
	gl.Buffer(strbuf.vbuffer).Bind(gl.ARRAY_BUFFER)
	gl.VertexPointer(2, gl.FLOAT, int(stride), nil)

	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	gl.Buffer(strbuf.ibuffer).Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.TexCoordPointer(2, gl.FLOAT, int(stride), unsafe.Offsetof(strbuf.vs[0].u))

	// TODO(tmckee): let's use gl.QUADS and simplify indices considerably...
	gl.DrawElements(gl.TRIANGLES, len(strbuf.is), gl.UNSIGNED_SHORT, nil)

	debug.LogAndClearGlErrors(log.Default())
}

type subImage struct {
	im     image.Image
	bounds image.Rectangle
}
type transparent struct{}

func (t transparent) RGBA() (r, g, b, a uint32) {
	return 0, 0, 0, 0
}
func (si *subImage) ColorModel() color.Model {
	return si.im.ColorModel()
}
func (si *subImage) Bounds() image.Rectangle {
	return si.bounds
}
func (si *subImage) At(x, y int) color.Color {
	b := si.bounds
	if (image.Point{x, y}).In(b) {
		return si.im.At(x, y)
	}
	return transparent{}
}

// Returns a sub-image of the input image. The bounding rectangle is the
// smallest possible rectangle that includes all pixels that have alpha > 0,
// with one pixel of border on all sides.
func minimalSubImage(src image.Image) *subImage {
	bounds := src.Bounds()
	var new_bounds image.Rectangle
	new_bounds.Max = bounds.Min
	new_bounds.Min = bounds.Max
	for x := bounds.Min.X; x <= bounds.Max.X; x++ {
		for y := bounds.Min.Y; y <= bounds.Max.Y; y++ {
			c := src.At(x, y)
			_, _, _, a := c.RGBA()
			if a > 0 {
				if x < new_bounds.Min.X {
					new_bounds.Min.X = x
				}
				if y < new_bounds.Min.Y {
					new_bounds.Min.Y = y
				}
				if x > new_bounds.Max.X {
					new_bounds.Max.X = x
				}
				if y > new_bounds.Max.Y {
					new_bounds.Max.Y = y
				}
			}
		}
	}

	// // We want one row/col of boundary between characters so that we don't get
	// // annoying artifacts
	new_bounds.Min.X--
	new_bounds.Min.Y--
	new_bounds.Max.X++
	new_bounds.Max.Y++

	if new_bounds.Min.X > new_bounds.Max.X || new_bounds.Min.Y > new_bounds.Max.Y {
		new_bounds = image.Rect(0, 0, 0, 0)
	}

	return &subImage{src, new_bounds}
}

// This stupid thing is just so that our idiot-packedImage can answer queries
// faster.  If we're going to query every pixel then it makes sense to check
// the largest rectangles first, since they will be the correct response more
// often than the smaller rectangles.
type packedImageSortByArea struct {
	*packedImage
}

func (p *packedImageSortByArea) Len() int {
	return len(p.ims)
}
func (p *packedImageSortByArea) Less(i, j int) bool {
	ai := p.ims[i].Bounds().Dx() * p.ims[i].Bounds().Dy()
	aj := p.ims[j].Bounds().Dx() * p.ims[j].Bounds().Dy()
	return ai > aj
}
func (p *packedImageSortByArea) Swap(i, j int) {
	p.ims[i], p.ims[j] = p.ims[j], p.ims[i]
	p.off[i], p.off[j] = p.off[j], p.off[i]
}

type packedImage struct {
	ims    []image.Image
	off    []image.Point
	bounds image.Rectangle
}

func (p *packedImage) Len() int {
	return len(p.ims)
}
func (p *packedImage) Less(i, j int) bool {
	return p.ims[i].Bounds().Dy() < p.ims[j].Bounds().Dy()
}
func (p *packedImage) Swap(i, j int) {
	p.ims[i], p.ims[j] = p.ims[j], p.ims[i]
	p.off[i], p.off[j] = p.off[j], p.off[i]
}
func (p *packedImage) GetRect(im image.Image) image.Rectangle {
	for i := range p.ims {
		if im == p.ims[i] {
			return p.ims[i].Bounds().Add(p.off[i])
		}
	}
	return image.Rectangle{}
}
func (p *packedImage) ColorModel() color.Model {
	return p.ims[0].ColorModel()
}
func (p *packedImage) Bounds() image.Rectangle {
	return p.bounds
}
func (p *packedImage) At(x, y int) color.Color {
	point := image.Point{x, y}
	for i := range p.ims {
		if point.In(p.ims[i].Bounds().Add(p.off[i])) {
			return p.ims[i].At(x-p.off[i].X, y-p.off[i].Y)
		}
	}
	return transparent{}
}

func packImages(ims []image.Image) *packedImage {
	var p packedImage
	if len(ims) == 0 {
		panic("Cannot pack zero images")
	}
	p.ims = ims
	p.off = make([]image.Point, len(p.ims))
	sort.Sort(&p)

	run := 0
	height := 0
	max_width := 512
	max_height := 0
	for i := 1; i < len(p.off); i++ {
		run += p.ims[i-1].Bounds().Dx()
		if run+p.ims[i].Bounds().Dx() > max_width {
			run = 0
			height += max_height
			max_height = 0
		}
		if p.ims[i].Bounds().Dy() > max_height {
			max_height = p.ims[i].Bounds().Dy()
		}
		p.off[i].X = run
		p.off[i].Y = height
	}
	for i := range p.ims {
		p.off[i] = p.off[i].Sub(p.ims[i].Bounds().Min)
	}

	// Done packing - now figure out the resulting bounds
	p.bounds.Min.X = 1e9 // if we exceed this something else will break first
	p.bounds.Min.Y = 1e9
	p.bounds.Max.X = -1e9
	p.bounds.Max.Y = -1e9
	for i := range p.ims {
		b := p.ims[i].Bounds()
		min := b.Add(p.off[i]).Min
		max := b.Add(p.off[i]).Max
		if min.X < p.bounds.Min.X {
			p.bounds.Min.X = min.X
		}
		if min.Y < p.bounds.Min.Y {
			p.bounds.Min.Y = min.Y
		}
		if max.X > p.bounds.Max.X {
			p.bounds.Max.X = max.X
		}
		if max.Y > p.bounds.Max.Y {
			p.bounds.Max.Y = max.Y
		}
	}

	sort.Sort(&packedImageSortByArea{&p})

	return &p
}

func MakeDictionary(font *truetype.Font, size int) *Dictionary {
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
		adv_x := float64(advance.X) / 256.0
		rune_info[r] = runeInfo{Bounds: sub.bounds, Advance: adv_x}
	}
	packed := packImages(letters)

	for _, r := range alphabet {
		ri := rune_info[r]
		ri.Pos = packed.GetRect(rune_mapping[r])
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

	dict.setupGlStuff()

	return &dict
}

var init_once sync.Once

func LoadDictionary(r io.Reader) (*Dictionary, error) {
	// TODO(tmckee): we shouldn't coulple loading a dictionary to registering
	// shaders.
	init_once.Do(func() {
		render.Queue(func() {
			err := render.RegisterShader("glop.font", []byte(font_vertex_shader), []byte(font_fragment_shader))
			if err != nil {
				panic(err)
			}
		})
		render.Purge()
	})

	var d Dictionary
	err := gob.NewDecoder(r).Decode(&d.Data)
	if err != nil {
		return nil, err
	}
	d.setupGlStuff()
	return &d, nil
}

func (d *Dictionary) Store(outputStream io.Writer) error {
	return gob.NewEncoder(outputStream).Encode(d.Data)
}

// Sets up anything that wouldn't have been loaded from disk, including
// all opengl data.
func (d *Dictionary) setupGlStuff() {
	d.strs = make(map[string]strBuffer)
	d.pars = make(map[string]strBuffer)

	render.Queue(func() {
		gl.Enable(gl.TEXTURE_2D)
		d.texture = uint32(gl.GenTexture())
		gl.Texture(d.texture).Bind(gl.TEXTURE_2D)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

		gl.ActiveTexture(gl.TEXTURE0+0)
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
