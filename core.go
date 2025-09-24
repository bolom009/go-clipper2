package go_clipper2

import "math"

// ClipType specifies the type of boolean operation
type ClipType uint8

const (
	NoClip       ClipType = iota
	Intersection          // intersect subject and clip polygons
	Union                 // union (OR) subject and clip polygons
	Difference            // subtract clip polygons from subject polygons
	Xor                   // exclusively or (XOR) subject and clip polygons
)

// FillRule specifies how to determine polygon interiors for self-intersecting polygons
type FillRule uint8

const (
	EvenOdd  FillRule = iota // odd numbered sub-regions are filled
	NonZero                  // non-zero sub-regions are filled
	Positive                 // positive sub-regions are filled
	Negative                 // negative sub-regions are filled
)

type PathType uint8

const (
	Subject PathType = iota
	Clip
)

type MidpointRounding uint8

const (
	ToEven = iota
	AwayFromZero
)

// Point64 represents a point with 64-bit integer coordinates
type Point64 struct {
	X, Y int64
}

func NewFloatPoint64(x, y float64) Point64 {
	rX := 0.0
	rY := 0.0

	if x > 0 {
		rX = math.Floor(x + 0.5)
	} else {
		rX = math.Ceil(x - 0.5)
	}

	if y > 0 {
		rY = math.Floor(y + 0.5)
	} else {
		rY = math.Ceil(y - 0.5)
	}

	return Point64{X: int64(rX), Y: int64(rY)}
}

func (p *Point64) ToPointD() PointD {
	return PointD{X: float64(p.X), Y: float64(p.Y)}
}

func (p *Point64) ToPointDScale(scale float64) PointD {
	return PointD{X: float64(p.X) * scale, Y: float64(p.Y) * scale}
}

func (p *Point64) Equals(p2 Point64) bool {
	return p.X == p2.X && p.Y == p2.Y
}

func (p *Point64) NEquals(p2 Point64) bool {
	return p.X != p2.X || p.Y != p2.Y
}

func (p *Point64) Add(p2 Point64) {
	p.X += p2.X
	p.Y += p2.Y
}

func (p *Point64) Sub(p2 Point64) {
	p.X -= p2.X
	p.Y -= p2.Y
}

func (p *Point64) ToPoint64(pt PointD) Point64 {
	rX := 0.0
	rY := 0.0

	if pt.X > 0 {
		rX = math.Floor(pt.X + 0.5)
	} else {
		rX = math.Ceil(pt.X - 0.5)
	}

	if pt.Y > 0 {
		rY = math.Floor(pt.Y + 0.5)
	} else {
		rY = math.Ceil(pt.Y - 0.5)
	}

	return Point64{X: int64(rX), Y: int64(rY)}
}

type PointD struct {
	X, Y float64
}

func (p *PointD) ToPoint64() Point64 {
	rX := 0.0
	rY := 0.0

	if p.X > 0 {
		rX = math.Floor(p.X + 0.5)
	} else {
		rX = math.Ceil(p.X - 0.5)
	}

	if p.Y > 0 {
		rY = math.Floor(p.Y + 0.5)
	} else {
		rY = math.Ceil(p.Y - 0.5)
	}

	return Point64{X: int64(rX), Y: int64(rY)}
}

func (p *PointD) ToPoint64Scale(scale float64) Point64 {
	rX := 0.0
	rY := 0.0

	if p.X > 0 {
		rX = math.Floor(p.X + 0.5)
	} else {
		rX = math.Ceil(p.X - 0.5)
	}

	if p.Y > 0 {
		rY = math.Floor(p.Y + 0.5)
	} else {
		rY = math.Ceil(p.Y - 0.5)
	}

	return Point64{X: int64(rX * scale), Y: int64(rY * scale)}
}

func (p *PointD) Scale(scale float64) {
	p.X *= scale
	p.Y *= scale
}

func (p *PointD) Equals(p2 PointD) bool {
	return isAlmostZero(p.X-p2.X) && isAlmostZero(p.Y-p2.Y)
}

func (p *PointD) NEquals(p2 PointD) bool {
	return !isAlmostZero(p.X-p2.X) || !isAlmostZero(p.Y-p2.Y)
}

func (p *PointD) Negate() {
	p.X *= -1
	p.Y *= -1
}

type Rect64 struct {
	left   int64
	top    int64
	right  int64
	bottom int64
}

func NewRect64(left, top, right, bottom int64) Rect64 {
	return Rect64{
		left:   left,
		top:    top,
		right:  right,
		bottom: bottom,
	}
}

func NewRect64Invalid(isValid bool) Rect64 {
	if isValid {
		return Rect64{}
	}

	return Rect64{
		left:   math.MaxInt64,
		top:    math.MaxInt64,
		right:  math.MaxInt64,
		bottom: math.MaxInt64,
	}
}

func (r *Rect64) IsEmpty() bool {
	return r.bottom <= r.top || r.right <= r.left
}

func (r *Rect64) IsInvalid() bool {
	return r.left < math.MaxInt64
}

func (r *Rect64) MidPoint() Point64 {
	return Point64{X: (r.left + r.right) / 2, Y: (r.top + r.bottom) / 2}
}

func (r *Rect64) Contains(rec Rect64) bool {
	return rec.left >= r.left && rec.right <= r.right &&
		rec.top >= r.top && rec.bottom <= r.bottom
}

func (r *Rect64) Intersects(rec Rect64) bool {
	return max(r.left, rec.left) <= min(r.right, rec.right) && max(r.top, rec.top) <= min(r.bottom, rec.bottom)
}

func (r *Rect64) AsPath() Path64 {
	return Path64{
		Point64{X: r.left, Y: r.top},
		Point64{X: r.right, Y: r.top},
		Point64{X: r.right, Y: r.bottom},
		Point64{X: r.left, Y: r.bottom},
	}
}

type RectD struct {
	left   float64
	top    float64
	right  float64
	bottom float64
}

func NewRectD(left, top, right, bottom float64) RectD {
	return RectD{
		left:   left,
		top:    top,
		right:  right,
		bottom: bottom,
	}
}

func NewRectDInvalid(isValid bool) RectD {
	if isValid {
		return RectD{}
	}

	return RectD{
		left:   math.MaxFloat64,
		top:    math.MaxFloat64,
		right:  math.MaxFloat64,
		bottom: math.MaxFloat64,
	}
}

func (r *RectD) IsEmpty() bool {
	return r.bottom <= r.top || r.right <= r.left
}

func (r *RectD) IsInvalid() bool {
	return r.left < math.MaxInt64
}

func (r *RectD) MidPoint() PointD {
	return PointD{X: (r.left + r.right) / 2, Y: (r.top + r.bottom) / 2}
}

func (r *RectD) Contains(rec RectD) bool {
	return rec.left >= r.left && rec.right <= r.right &&
		rec.top >= r.top && rec.bottom <= r.bottom
}

func (r *RectD) Intersects(rec RectD) bool {
	return max(r.left, rec.left) <= min(r.right, rec.right) && max(r.top, rec.top) <= min(r.bottom, rec.bottom)
}

func (r *RectD) AsPath() PathD {
	return PathD{
		PointD{X: r.left, Y: r.top},
		PointD{X: r.right, Y: r.top},
		PointD{X: r.right, Y: r.bottom},
		PointD{X: r.left, Y: r.bottom},
	}
}

// Path64 represents a sequence of points forming a path
type Path64 []Point64

// Paths64 represents a collection of paths
type Paths64 []Path64

// PathD represents a sequence of points forming a path
type PathD []PointD

// PathsD represents a collection of paths
type PathsD []PathD
