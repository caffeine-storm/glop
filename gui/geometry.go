package gui

import (
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
)

type Point struct {
	X, Y int
}

func (p Point) Add(q Point) Point {
	return Point{
		X: p.X + q.X,
		Y: p.Y + q.Y,
	}
}

func (p Point) Inside(r Region) bool {
	if p.X < r.X {
		return false
	}
	if p.Y < r.Y {
		return false
	}
	if p.X >= r.X+r.Dx {
		return false
	}
	if p.Y >= r.Y+r.Dy {
		return false
	}
	return true
}

type Dims struct {
	Dx, Dy int
}

type Region struct {
	Point
	Dims
}

func (r Region) Add(p Point) Region {
	return Region{
		r.Point.Add(p),
		r.Dims,
	}
}

// Returns a region that is no larger than r that fits inside t.  The region
// will be located as closely as possible to r.
func (r Region) Fit(t Region) Region {
	if r.Dx > t.Dx {
		r.Dx = t.Dx
	}
	if r.Dy > t.Dy {
		r.Dy = t.Dy
	}
	if r.X < t.X {
		r.X = t.X
	}
	if r.Y < t.Y {
		r.Y = t.Y
	}
	if r.X+r.Dx > t.X+t.Dx {
		r.X -= (r.X + r.Dx) - (t.X + t.Dx)
	}
	if r.Y+r.Dy > t.Y+t.Dy {
		r.Y -= (r.Y + r.Dy) - (t.Y + t.Dy)
	}
	return r
}

func (r Region) Isect(s Region) Region {
	if r.X < s.X {
		r.Dx -= s.X - r.X
		r.X = s.X
	}
	if r.Y < s.Y {
		r.Dy -= s.Y - r.Y
		r.Y = s.Y
	}
	if r.X+r.Dx > s.X+s.Dx {
		r.Dx -= (r.X + r.Dx) - (s.X + s.Dx)
	}
	if r.Y+r.Dy > s.Y+s.Dy {
		r.Dy -= (r.Y + r.Dy) - (s.Y + s.Dy)
	}
	if r.Dx < 0 {
		r.Dx = 0
	}
	if r.Dy < 0 {
		r.Dy = 0
	}
	return r
}

func (r Region) Size() int {
	return r.Dx * r.Dy
}

// Need a global stack of regions because opengl only handles pushing/popping
// the state of the enable bits for each clip plane, not the planes themselves
var clippers []Region

// If we just declared this in setClipPlanes it would get allocated on the heap
// because we have to take the address of it to pass it to opengl.  By having
// it here we avoid that allocation - it amounts to a lot of someone is calling
// this every frame.
var eqs [4][4]float64

func (r Region) setClipPlanes() {
	leftExtent := float64(r.X)
	rightExtent := float64(r.X + r.Dx)
	bottomExtent := float64(r.Y)
	topExtent := float64(r.Y + r.Dy)

	eqs[0][0], eqs[0][1], eqs[0][2], eqs[0][3] = 1, 0, 0, -leftExtent
	eqs[1][0], eqs[1][1], eqs[1][2], eqs[1][3] = -1, 0, 0, rightExtent
	eqs[2][0], eqs[2][1], eqs[2][2], eqs[2][3] = 0, 1, 0, -bottomExtent
	eqs[3][0], eqs[3][1], eqs[3][2], eqs[3][3] = 0, -1, 0, topExtent
	gl.ClipPlane(gl.CLIP_PLANE0, eqs[0][:])
	gl.ClipPlane(gl.CLIP_PLANE1, eqs[1][:])
	gl.ClipPlane(gl.CLIP_PLANE2, eqs[2][:])
	gl.ClipPlane(gl.CLIP_PLANE3, eqs[3][:])
}

func (r Region) PushClipPlanes() {
	glog.TraceLogger().Trace("pushclip", "clippers", clippers)
	if len(clippers) == 0 {
		gl.Enable(gl.CLIP_PLANE0)
		gl.Enable(gl.CLIP_PLANE1)
		gl.Enable(gl.CLIP_PLANE2)
		gl.Enable(gl.CLIP_PLANE3)
		r.setClipPlanes()
		clippers = append(clippers, r)
	} else {
		cur := clippers[len(clippers)-1]
		clippers = append(clippers, r.Isect(cur))
		clippers[len(clippers)-1].setClipPlanes()
	}
}

func (r Region) PopClipPlanes() {
	clippers = clippers[0 : len(clippers)-1]
	if len(clippers) == 0 {
		gl.Disable(gl.CLIP_PLANE0)
		gl.Disable(gl.CLIP_PLANE1)
		gl.Disable(gl.CLIP_PLANE2)
		gl.Disable(gl.CLIP_PLANE3)
	} else {
		clippers[len(clippers)-1].setClipPlanes()
	}
}
