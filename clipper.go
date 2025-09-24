package go_clipper2

import "math"

func IntersectPaths64(subject Paths64, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Intersection, subject, clip, fillRule)
}

func IntersectPathsD(subject PathsD, clip PathsD, fillRule FillRule, precision int) PathsD {
	return BooleanOpPathsD(Intersection, subject, clip, fillRule, precision)
}

func UnionPaths64(subject Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Union, subject, nil, fillRule)
}

func UnionWithClipPaths64(subject Paths64, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Union, subject, clip, fillRule)
}

func UnionPathsD(subject PathsD, fillRule FillRule, precision int) PathsD {
	return BooleanOpPathsD(Union, subject, nil, fillRule, precision)
}

func UnionWithClipPathsD(subject PathsD, clip PathsD, fillRule FillRule, precision int) PathsD {
	return BooleanOpPathsD(Union, subject, clip, fillRule, precision)
}

func DifferencePaths64(subject Paths64, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Difference, subject, clip, fillRule)
}

func DifferencePathsD(subject PathsD, clip PathsD, fillRule FillRule, precision int) PathsD {
	return BooleanOpPathsD(Difference, subject, clip, fillRule, precision)
}

func XorPaths64(subject Paths64, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Xor, subject, clip, fillRule)
}

func XorPathsD(subject PathsD, clip PathsD, fillRule FillRule, precision int) PathsD {
	return BooleanOpPathsD(Xor, subject, clip, fillRule, precision)
}

func BooleanOpPaths64(clipType ClipType, subject Paths64, clip Paths64, fillRule FillRule) Paths64 {
	if subject == nil {
		return Paths64{}
	}

	solution := make(Paths64, 0)
	c := NewClipper64()
	c.AddPaths(subject, Subject, false)
	if clip != nil {
		c.AddPaths(clip, Clip, false)
	}

	c.Execute(clipType, fillRule, &solution)
	return solution
}

func BooleanOpPathsD(clipType ClipType, subject PathsD, clip PathsD, fillRule FillRule, precision int) PathsD {
	if precision == 0 {
		precision = 2
	}

	solution := make(PathsD, 0)
	c := NewClipperD(precision)
	c.AddPaths(subject, Subject, false)
	if clip != nil {
		c.AddPaths(clip, Clip, false)
	}

	c.Execute(clipType, fillRule, solution)
	return solution
}

// TODO
// InflatePaths64
// InflatePathsD
// RectClipPaths64
// RectClipPath64
// RectClipPathsD
// RectClipPathD
// RectClipLines
// RectClipLinesPaths64
// RectClipLinesPaths64
// RectClipLinesPathsD
// RectClipLinesPathD
// Minkowski methods

func StripDuplicates(path Path64, isClosedPath bool) Path64 {
	cnt := len(path)
	if cnt == 0 {
		return path
	}

	result := make(Path64, 0, cnt)
	lastPt := path[0]
	result = append(result, lastPt)
	for i := 1; i < cnt; i++ {
		if lastPt.X != path[i].X && lastPt.Y != path[i].Y {
			lastPt = path[i]
			result = append(result, lastPt)
		}
	}

	if isClosedPath && lastPt.X == result[0].X && lastPt.Y == result[0].Y {
		result = result[:len(result)-1]
	}

	return result
}

func Area64(path Path64) float64 {
	if len(path) < 3 {
		return 0.0
	}
	a := 0.0
	prevPt := path[len(path)-1]
	for _, pt := range path {
		a += float64(prevPt.Y+pt.Y) * float64(prevPt.X-pt.X)
		prevPt = pt
	}
	return a * 0.5
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
	if IsAlmostZero(scale - 1) {
		return path
	}

	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point64{X: int64(float64(pt.X) * scale), Y: int64(float64(pt.Y) * scale)}
	}

	return result
}

func ScalePathD(path PathD, scale float64) PathD {
	if IsAlmostZero(scale - 1) {
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
		result[i] = Point64{X: int64(pt.X * scale), Y: int64(pt.Y * scale)}
	}

	return result
}

func ScalePath64ToPathD(path Path64, scale float64) PathD {
	result := make(PathD, len(path))
	for i, pt := range path {
		result[i] = PointD{X: float64(pt.X) * scale, Y: float64(pt.Y) * scale}
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

func Sqr[T Numeric](val T) T {
	return val * val
}

func PointsNearEqual(pt1, pt2 PointD, distanceSqrd float64) bool {
	return Sqr(pt1.X-pt2.X)+Sqr(pt1.Y-pt2.Y) < distanceSqrd
}

func PerpendicDistFromLineSqrD(pt, line1, line2 PointD) float64 {
	a := pt.X - line1.X
	b := pt.Y - line1.Y
	c := line2.X - line1.X
	d := line2.Y - line1.Y

	if c == 0 && d == 0 {
		return 0
	}

	return Sqr(a*d-c*b) / (c*c + d*d)
}

func perpendicDistFromLineSqr64(pt, line1, line2 Point64) float64 {
	a := pt.X - line1.X
	b := pt.Y - line1.Y
	c := line2.X - line1.X
	d := line2.Y - line1.Y

	if c == 0 && d == 0 {
		return 0
	}

	return float64(Sqr(a*d-c*b)) / float64(c*c+d*d)
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
	// первый пункт — центр + радиус по X
	result = append(result, Point64{
		X: center.X + int64(radiusX),
		Y: center.Y,
	})

	for i := 1; i < steps; i++ {
		x := dx*co - dy*si
		dy = dy*co + dx*si
		dx = x
		result = append(result, Point64{
			X: center.X + int64(radiusX*dx),
			Y: center.Y + int64(radiusY*dy),
		})
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
	// Первая точка — центр + радиус по X
	result = append(result, PointD{X: center.X + radiusX, Y: center.Y})

	for i := 1; i < steps; i++ {
		result = append(result, PointD{X: center.X + radiusX*dx, Y: center.Y + radiusY*dy})
		x := dx*co - dy*si
		dy = dy*co + dx*si
		dx = x
	}
	return result
}

func ReversePath64(s Path64) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func ReversePathD(s PathD) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
