package sprite

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/draw"
	"os"
	"path/filepath"

	"github.com/go-gl-legacy/gl"
	"github.com/go-gl-legacy/glu"
	"github.com/runningwild/glop/cache"
	"github.com/runningwild/glop/render"
	yed "github.com/runningwild/yedparse"
)

// An id that specifies a specific frame along with its facing.  This is used
// to index into sprite sheets.
type frameId struct {
	facing int
	node   int
}
type frameIdArray []frameId

func (fia frameIdArray) Len() int {
	return len(fia)
}
func (fia frameIdArray) Less(i, j int) bool {
	if fia[i].facing != fia[j].facing {
		return fia[i].facing < fia[j].facing
	}
	return fia[i].node < fia[j].node
}
func (fia frameIdArray) Swap(i, j int) {
	fia[i], fia[j] = fia[j], fia[i]
}

// A sheet contains a group of frames of animations indexed by frameId
type sheet struct {
	rects  map[frameId]FrameRect
	dx, dy int
	// TODO(tmckee): verify correctness.
	// The 'Sprite_path' for an entity as stored in <entitity>.json files.
	spritePath string
	anim       *yed.Graph

	// Unique name that is based on the path of the sprite and the list of
	// frameIds used to generate this sheet.  This name is used to store the
	// sheet on disk when not in use.
	name string

	// Channel for sending reference-count updates (+1/-1 only)
	reference_chan chan int
	// Channel for sending load/unload requests (true: load, false: unload)
	load_chan chan bool
	texture   gl.Texture

	pixelDataCache cache.ByteBank
}

func (s *sheet) Load() {
	s.reference_chan <- 1
}

func (s *sheet) Unload() {
	s.reference_chan <- -1
}

func (s *sheet) getCacheKey() string {
	return s.name
}

func (s *sheet) compose(pixer chan<- []byte) {
	bytes, ok, err := s.pixelDataCache.Read(s.getCacheKey())
	if err != nil {
		panic(fmt.Errorf("couldn't read from pixelDataCache: %w", err))
	}
	if ok {
		pixer <- bytes
		return
	}

	rect := image.Rect(0, 0, s.dx, s.dy)
	canvas := image.NewRGBA(rect)
	for fid, rect := range s.rects {
		name := s.anim.Node(fid.node).Line(0) + ".png"
		file, err := os.Open(filepath.Join(s.spritePath, fmt.Sprintf("%d", fid.facing), name))
		// if a file isn't there that's ok
		if err != nil {
			continue
		}

		im, _, err := image.Decode(file)
		file.Close()
		// if a file can't be read that is *not* ok, TODO: Log an error or something
		if err != nil {
			continue
		}
		draw.Draw(canvas, image.Rect(rect.X, s.dy-rect.Y, rect.X2, s.dy-rect.Y2), im, image.Point{}, draw.Src)
	}

	// Cache the bytes for later use.
	err = s.pixelDataCache.Write(s.getCacheKey(), canvas.Pix)
	if err != nil {
		panic(fmt.Errorf("couldn't write byte source: %w", err))
	}
	pixer <- canvas.Pix
}

// TODO: This was copied from the gui package, probably should just have some basic
// texture loading utils that do this common stuff
func nextPowerOf2(n uint32) uint32 {
	if n == 0 {
		return 1
	}
	for i := uint(0); i < 32; i++ {
		p := uint32(1) << i
		if n <= p {
			return p
		}
	}
	return 0
}

func (s *sheet) makeTexture(pixer <-chan []byte) {
	s.texture = gl.GenTexture()
	s.texture.Bind(gl.TEXTURE_2D)
	gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	// TODO(tmckee): pulling from 'pixer' here will block the render thread on
	// disk IO. We should be pulling the pixels first then passing them over to
	// the render thread instead.
	data := <-pixer

	// TODO(tmckee:#13): ... gl.INT is wrong, I think. We should be getting pixel
	// data from an image.RGBA which stores each pixel in 32 bits but gl.INT
	// means each _component_ is 32 bits... a.k.a. each pixel is 128 bits.
	glu.Build2DMipmaps(gl.TEXTURE_2D, 4, s.dx, s.dy, gl.RGBA, gl.INT, data)
	// glu.Build2DMipmaps(gl.TEXTURE_2D, 4, s.dx, s.dy, gl.RGBA, gl.UNSIGNED_BYTE, data)
}

func (s *sheet) loadRoutine(renderQueue render.RenderQueueInterface) {
	ready := make(chan bool, 1)
	pixer := make(chan []byte)
	for load := range s.load_chan {
		if load {
			go s.compose(pixer)
			// TODO(tmckee): clean: we don't need to spawn a go-routine to send a
			// func on a chan.
			go func() {
				renderQueue.Queue(func(render.RenderQueueState) {
					s.makeTexture(pixer)
					ready <- true
				})
			}()
		} else {
			go func() {
				<-ready
				renderQueue.Queue(func(render.RenderQueueState) {
					s.texture.Delete()
					s.texture = 0
				})
			}()
		}
	}
}

// TODO: Need to set up a finalizer on this thing so that we don't keep this
// texture memory around forever if we forget about it
func (s *sheet) routine(renderQueue render.RenderQueueInterface) {
	go s.loadRoutine(renderQueue)
	references := 0
	for load := range s.reference_chan {
		if load < 0 {
			if references == 0 {
				panic(fmt.Sprintf("Tried to unload a sprite (%s/%s) sheet more times than it was loaded.", s.name, s.spritePath))
			}
			references--
			if references == 0 {
				s.load_chan <- false
			}
		} else if load > 0 {
			if references == 0 {
				s.load_chan <- true
			}
			references++
		} else {
			panic("value of 0 should never be sent along load_chan")
		}
	}
}

func uniqueName(fids []frameId) string {
	var b []byte
	for i := range fids {
		b = append(b, byte(fids[i].facing))
		b = append(b, byte(fids[i].node))
	}
	h := fnv.New64()
	h.Write(b)
	return fmt.Sprintf("%x.gob", h.Sum64())
}

func makeSheet(path string, anim *yed.Graph, fids []frameId, byteBank cache.ByteBank, renderQueue render.RenderQueueInterface) (*sheet, error) {
	s := sheet{
		spritePath:     path,
		anim:           anim,
		name:           uniqueName(fids),
		pixelDataCache: byteBank,
	}
	s.rects = make(map[frameId]FrameRect)
	cy := 0
	cx := 0
	cdy := 0
	tdx := 0
	max_width := 2048
	for _, fid := range fids {
		name := anim.Node(fid.node).Line(0) + ".png"
		file, err := os.Open(filepath.Join(path, fmt.Sprintf("%d", fid.facing), name))
		// if a file isn't there that's ok
		if err != nil {
			continue
		}

		config, _, err := image.DecodeConfig(file)
		file.Close()
		// if a file can't be read that is *not* ok
		if err != nil {
			return nil, err
		}

		if cx+config.Width > max_width {
			cx = 0
			cy += cdy
			cdy = 0
		}
		if config.Height > cdy {
			cdy = config.Height
		}
		s.rects[fid] = FrameRect{X: cx, X2: cx + config.Width, Y: cy, Y2: cy + config.Height}
		cx += config.Width
		if cx > tdx {
			tdx = cx
		}
	}
	s.dx = int(nextPowerOf2(uint32(tdx)))
	s.dy = int(nextPowerOf2(uint32(cy + cdy)))
	s.load_chan = make(chan bool)
	s.reference_chan = make(chan int)
	go s.routine(renderQueue)

	return &s, nil
}
