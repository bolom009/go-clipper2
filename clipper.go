package go_clipper2

import (
	"math"

	"github.com/govalues/decimal"
)

func StripDuplicates(path Path64, isClosedPath bool) Path64 {
	cnt := len(path)
	if cnt == 0 {
		return path
	}

	result := make(Path64, 0, cnt)
	lastPt := path[0]
	result = append(result, lastPt)
	for i := 1; i < cnt; i++ {
		if lastPt.NEquals(path[i]) {
			lastPt = path[i]
			result = append(result, lastPt)
		}
	}

	if isClosedPath && lastPt.Equals(result[0]) {
		var err error
		result, err = removeAtIndex(result, len(result)-1)
		if err != nil {
			panic(ErrInvalidRemoveListIndex)
		}
	}

	return result
}

func Area64(path Path64) float64 {
	if len(path) < 3 {
		return 0.0
	}

	var a int64 = 0
	prevPt := path[len(path)-1]
	for _, pt := range path {
		a += (prevPt.Y + pt.Y) * (prevPt.X - pt.X)
		prevPt = pt
	}
	return float64(a) * 0.5
}

func AreaD(path PathD) float64 {
	if len(path) < 3 {
		return 0.0
	}
	a := 0.0
	prevPt := path[len(path)-1]
	for _, pt := range path {
		a += (prevPt.Y + pt.Y) * (prevPt.X - pt.X)
		prevPt = pt
	}
	return a * 0.5
}

func AreaPaths64(paths Paths64) float64 {
	a := 0.0
	for _, path := range paths {
		a += Area64(path)
	}
	return a
}

func AreaPathsD(paths PathsD) float64 {
	a := 0.0
	for _, path := range paths {
		a += AreaD(path)
	}
	return a
}

func IsPositive64(poly Path64) bool {
	return Area64(poly) >= 0
}

func IsPositiveD(poly PathD) bool {
	return AreaD(poly) >= 0
}

func OffsetPath(path Path64, dx, dy int64) Path64 {
	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point64{X: pt.X + dx, Y: pt.Y + dy}
	}

	return result
}

func ScaleRect64(rec Rect64, scale float64) Rect64 {
	return Rect64{
		left:   int64(float64(rec.left) * scale),
		top:    int64(float64(rec.top) * scale),
		right:  int64(float64(rec.right) * scale),
		bottom: int64(float64(rec.bottom) * scale),
	}
}

func ScalePath64(path Path64, scale float64) Path64 {
	if isAlmostZero(scale - 1) {
		return path
	}

	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point64{X: int64(float64(pt.X) * scale), Y: int64(float64(pt.Y) * scale)}
	}

	return result
}

func ScaleRectD(rec RectD, scale float64) Rect64 {
	return Rect64{
		left:   int64(rec.left * scale),
		top:    int64(rec.top * scale),
		right:  int64(rec.right * scale),
		bottom: int64(rec.bottom * scale),
	}
}

func ScalePathD(path PathD, scale float64) PathD {
	if isAlmostZero(scale - 1) {
		return path
	}

	result := make(PathD, len(path))
	for i, pt := range path {
		result[i] = PointD{X: pt.X * scale, Y: pt.Y * scale}
	}

	return result
}

func ScalePathDToPath64(path PathD, scale float64) Path64 {
	result := make(Path64, len(path))
	for i, pt := range path {
		mulX := pt.X * scale
		mulY := pt.Y * scale

		dx, _ := decimal.NewFromFloat64(mulX)
		dy, _ := decimal.NewFromFloat64(mulY)

		x, _, _ := dx.Int64(0)
		y, _, _ := dy.Int64(0)

		result[i] = Point64{X: x, Y: y}
	}

	return result
}

func ScalePath64ToPathD(path Path64, scale float64) PathD {
	dScale, _ := decimal.NewFromFloat64(scale)

	result := make(PathD, len(path))
	for i, pt := range path {
		ptX, _ := decimal.New(pt.X, 0)
		ptY, _ := decimal.New(pt.Y, 0)

		mulX, _ := ptX.Mul(dScale)
		mulY, _ := ptY.Mul(dScale)

		x, _ := mulX.Float64()
		y, _ := mulY.Float64()

		result[i] = PointD{X: x, Y: y}
	}

	return result
}

func ScalePathsDToPaths64(paths PathsD, scale float64) Paths64 {
	result := make(Paths64, len(paths))
	for i, path := range paths {
		result[i] = ScalePathDToPath64(path, scale)
	}

	return result
}

func ScalePaths64ToPathsD(paths Paths64, scale float64) PathsD {
	result := make(PathsD, len(paths))
	for i, path := range paths {
		result[i] = ScalePath64ToPathD(path, scale)
	}

	return result
}

func PathDToPath64(path PathD) Path64 {
	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point64{X: int64(pt.X), Y: int64(pt.Y)}
	}
	return result
}

func PathsDToPaths64(path PathsD) Paths64 {
	result := make(Paths64, len(path))
	for i, pt := range path {
		result[i] = PathDToPath64(pt)
	}
	return result
}

func Path64ToPathD(path Path64) PathD {
	result := make(PathD, len(path))
	for i, pt := range path {
		result[i] = PointD{X: float64(pt.X), Y: float64(pt.Y)}
	}
	return result
}

func Paths64ToPathsD(path Paths64) PathsD {
	result := make(PathsD, len(path))
	for i, pt := range path {
		result[i] = Path64ToPathD(pt)
	}
	return result
}

func GetBounds64(path Path64) Rect64 {
	result := NewRect64Invalid(false)
	for _, pt := range path {
		if pt.X < result.left {
			result.left = pt.X
		}
		if pt.X > result.right {
			result.right = pt.X
		}
		if pt.Y < result.top {
			result.top = pt.Y

		}
		if pt.Y > result.bottom {
			result.bottom = pt.Y
		}
	}

	if result.left == math.MaxInt64 {
		return Rect64{}
	}
	return result
}

func PointsNearEqual(pt1, pt2 PointD, distanceSqrd float64) bool {
	return sqr(pt1.X-pt2.X)+sqr(pt1.Y-pt2.Y) < distanceSqrd
}

func PerpendicDistFromLineSqrD(pt, line1, line2 PointD) float64 {
	a := pt.X - line1.X
	b := pt.Y - line1.Y
	c := line2.X - line1.X
	d := line2.Y - line1.Y

	if c == 0 && d == 0 {
		return 0
	}

	return sqr(a*d-c*b) / (c*c + d*d)
}

func PerpendicDistFromLineSqr64(pt, line1, line2 Point64) float64 {
	a := pt.X - line1.X
	b := pt.Y - line1.Y
	c := line2.X - line1.X
	d := line2.Y - line1.Y

	if c == 0 && d == 0 {
		return 0
	}

	return float64(sqr(a*d-c*b)) / float64(c*c+d*d)
}

func Ellipse64(center Point64, radiusX, radiusY float64, steps int) Path64 {
	if radiusX <= 0 {
		return Path64{}
	}
	if radiusY <= 0 {
		radiusY = radiusX
	}
	if steps <= 2 {
		steps = int(math.Ceil(math.Pi * math.Sqrt((radiusX+radiusY)/2)))
	}

	si := math.Sin(2 * math.Pi / float64(steps))
	co := math.Cos(2 * math.Pi / float64(steps))
	dx := co
	dy := si

	result := make(Path64, 0, steps)
	result = append(result, Point64{
		X: center.X + int64(radiusX),
		Y: center.Y,
	})

	for i := 1; i < steps; i++ {
		xV, _ := decimal.New(center.X, 0)
		yV, _ := decimal.New(center.Y, 0)

		mxV, _ := decimal.NewFromFloat64(radiusX * dx)
		myV, _ := decimal.NewFromFloat64(radiusY * dy)

		xsV, _ := xV.Add(mxV)
		ysV, _ := yV.Add(myV)

		x, _, _ := xsV.Int64(0)
		y, _, _ := ysV.Int64(0)

		pp := Point64{X: x, Y: y}
		result = append(result, pp)

		xx := dx*co - dy*si
		dy = dy*co + dx*si
		dx = xx
	}
	return result
}

func EllipseD(center PointD, radiusX, radiusY float64, steps int) PathD {
	if radiusX <= 0 {
		return PathD{}
	}
	if radiusY <= 0 {
		radiusY = radiusX
	}
	if steps <= 2 {
		steps = int(math.Ceil(math.Pi * math.Sqrt((radiusX+radiusY)/2)))
	}

	si := math.Sin(2 * math.Pi / float64(steps))
	co := math.Cos(2 * math.Pi / float64(steps))
	dx := co
	dy := si

	result := make(PathD, 0, steps)
	result = append(result, PointD{X: center.X + radiusX, Y: center.Y})

	for i := 1; i < steps; i++ {
		result = append(result, PointD{
			X: center.X + radiusX*dx,
			Y: center.Y + radiusY*dy,
		})
		x := dx*co - dy*si
		dy = dy*co + dx*si
		dx = x
	}
	return result
}

func MakePath64(vals ...int64) Path64 {
	l := len(vals) / 2

	result := make(Path64, l)
	for i := 0; i < l; i++ {
		result[i] = Point64{
			X: vals[i*2],
			Y: vals[i*2+1],
		}
	}

	return result
}

func MakePathD(vals ...float64) PathD {
	l := len(vals) / 2

	result := make(PathD, l)
	for i := 0; i < l; i++ {
		result[i] = PointD{
			X: vals[i*2],
			Y: vals[i*2+1],
		}
	}

	return result
}

func TranslatePath64(path Path64, dx, dy int64) Path64 {
	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point64{
			pt.X + dx,
			pt.Y + dy,
		}
	}

	return result
}

func TranslatePaths64(paths Paths64, dx, dy int64) Paths64 {
	result := make(Paths64, len(paths))
	for i, path := range paths {
		result[i] = OffsetPath(path, dx, dy)
	}

	return result
}

func TranslatePathD(path PathD, dx, dy float64) PathD {
	result := make(PathD, len(path))
	for i, pt := range path {
		result[i] = PointD{
			pt.X + dx,
			pt.Y + dy,
		}
	}

	return result
}

func TranslatePathsD(paths PathsD, dx, dy float64) PathsD {
	result := make(PathsD, len(paths))
	for i, path := range paths {
		result[i] = TranslatePathD(path, dx, dy)
	}

	return result
}
