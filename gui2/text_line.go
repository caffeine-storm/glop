package gui

import (
  "fmt"
  "image"
  "image/draw"
  "freetype-go.googlecode.com/hg/freetype"
  "freetype-go.googlecode.com/hg/freetype/truetype"
  "gl"
  "gl/glu"
  "io/ioutil"
  "os"
)

func mustLoadFont(filename string) *truetype.Font {
  data,err := ioutil.ReadFile(filename)
  if err != nil {
    panic(err.String())
  }
  font,err := freetype.ParseFont(data)
  if err != nil {
    panic(err.String())
  }
  return font
}

func drawText(font *truetype.Font, c *freetype.Context, color image.Color, rgba *image.RGBA, text string) (int,int) {
  fg := image.NewColorImage(color)
  bg := image.Transparent
  draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
  c.SetFont(font)
  c.SetDst(rgba)
  c.SetSrc(fg)
  c.SetClip(rgba.Bounds())
  // height is the fraction of the font that is above the line, 1.0 would mean
  // that the font never falls below the line
  height := 0.75
  pt := freetype.Pt(0, int(float64(c.FUnitToPixelRU(font.UnitsPerEm())) * height) )
  adv,_ := c.DrawString(text, pt)
  pt.X += adv.X
  py := int(float64(pt.Y >> 8) / height + 0.01)
  return int(pt.X >> 8), py
}

var basic_fonts map[string]*truetype.Font

func init() {
  basic_fonts = make(map[string]*truetype.Font)
}

func LoadFont(name,path string) os.Error {
  if _,ok := basic_fonts[name]; ok {
    return os.NewError(fmt.Sprintf("Cannot load two fonts with the same name: '%s'", name))
  }
  basic_fonts[name] = mustLoadFont(path)
  return nil
}

type TextLine struct {
  BasicWidget
  Childless
  NonResponder
  Rectangle
  text      string
  changed   bool
  rdims     Dims
  psize     int
  font      *truetype.Font
  context   *freetype.Context
  glyph_buf *truetype.GlyphBuf
  texture   gl.Texture
  rgba      *image.RGBA
  color     image.Color
}

func nextPowerOf2(n uint32) uint32 {
  if n == 0 { return 1 }
  for i := uint(0); i < 32; i++ {
    p := uint32(1) << i
    if n <= p { return p }
  }
  return 0
}

func (w *TextLine) figureDims() {
  w.Dims.Dx, w.Dims.Dy = drawText(w.font, w.context, w.color, image.NewRGBA(1, 1), w.text)
  w.rdims = Dims{
    Dx : int(nextPowerOf2(uint32(w.Dims.Dx))),
    Dy : int(nextPowerOf2(uint32(w.Dims.Dy))),
  }  
  w.rgba = image.NewRGBA(w.rdims.Dx, w.rdims.Dy)
  drawText(w.font, w.context, w.color, w.rgba, w.text)


  gl.Enable(gl.TEXTURE_2D)
  w.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, w.rdims.Dx, w.rdims.Dy, gl.RGBA, w.rgba.Pix)
  gl.Disable(gl.TEXTURE_2D)
}

func MakeTextLine(font_name,text string, r,g,b,a float64) *TextLine {
  var w TextLine
  w.BasicWidget.CoreWidget = &w
  font,ok := basic_fonts[font_name]
  if !ok {
    panic(fmt.Sprintf("Unable to find a font registered as '%s'.", font_name))
  }
  w.font = font
  w.glyph_buf = truetype.NewGlyphBuf()
  w.text = text
  w.psize = 72
  w.context = freetype.NewContext()
  w.context.SetDPI(132)
  w.context.SetFontSize(18)
  w.texture = gl.GenTexture()
  w.SetColor(r, g, b, a)
  w.figureDims()
  return &w
}

func (w *TextLine) SetColor(r,g,b,a float64) {
  w.color = image.RGBAColor{
    R : uint8(255 * r),
    G : uint8(255 * g),
    B : uint8(255 * b),
    A : uint8(255 * a),
  }
}

func (w *TextLine) SetText(str string) {
  if w.text != str {
    w.text = str
    w.changed = true
  }
}

func (w *TextLine) DoThink(_ int64) {
  if !w.changed { return }
  w.changed = false
  w.figureDims()
}

func (w *TextLine) Draw(region Region) {
  gl.PushMatrix()
  defer gl.PopMatrix()

  gl.PushAttrib(gl.TEXTURE_BIT)
  defer gl.PopAttrib()
  gl.Enable(gl.TEXTURE_2D)
  w.texture.Bind(gl.TEXTURE_2D)

  gl.PushAttrib(gl.COLOR_BUFFER_BIT)
  defer gl.PopAttrib()
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Translated(float64(region.X), float64(region.Y), 0)
  req := w.Rectangle
  req.Constrain(region)
  if req.Dx * w.Dims.Dy < req.Dy * w.Dims.Dx {
    req.Dy = int(float64(w.Dims.Dy) / float64(w.Dims.Dx) * float64(req.Dx))
  } else {
    req.Dx = int(float64(w.Dims.Dx) / float64(w.Dims.Dy) * float64(req.Dy))
  }
  fdx := float64(req.Dx)
  fdy := float64(req.Dy)
  tx := float64(w.Dims.Dx)/float64(w.rdims.Dx)
  ty := float64(w.Dims.Dy)/float64(w.rdims.Dy)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0,0)
    gl.Vertex2d(0, 0)
    gl.TexCoord2d(0,-ty)
    gl.Vertex2d(0, fdy)
    gl.TexCoord2d(tx,-ty)
    gl.Vertex2d(fdx, fdy)
    gl.TexCoord2d(tx,0)
    gl.Vertex2d(fdx, 0)
  gl.End()
}

