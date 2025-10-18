package go_clipper2

import (
	"math"
)

const (
	MaxInt64                 = int64(9223372036854775807)
	MaxCoord                 = MaxInt64 / 4
	max_coord                = float64(MaxCoord)
	min_coord                = -max_coord
	Invalid64                = MaxInt64
	floatingPointTolerance   = 1e-12
	defaultMinimumEdgeLength = 0.1
)

// CrossProduct for three Point64 (pt1->pt2 x pt2->pt3)
func CrossProduct(pt1, pt2, pt3 Point64) float64 {
	return float64((pt2.X-pt1.X)*(pt3.Y-pt2.Y) - (pt2.Y-pt1.Y)*(pt3.X-pt2.X))
}

func checkPrecision(precision int) {
	if precision < -8 || precision > 8 {
		panic(ErrPrecisionRange)
	}
}

func isAlmostZero(value float64) bool {
	return math.Abs(value) <= floatingPointTolerance
}

// triSign returns -1,0,1 but original returns 0,1,-1 mapping; keep same semantics:
// original: if x<0 return -1; return x>1 ? 1 : 0
func triSign(x int64) int {
	if x < 0 {
		return -1
	}
	if x > 1 {
		return 1
	}
	return 0
}

type UInt128Struct struct {
	Lo64 uint64
	Hi64 uint64
}

// MultiplyUInt64 multiplies two uint64 into 128-bit represented by UInt128Struct
func multiplyUInt64(a, b uint64) UInt128Struct {
	x1 := (a & 0xFFFFFFFF) * (b & 0xFFFFFFFF)
	x2 := (a>>32)*(b&0xFFFFFFFF) + (x1 >> 32)
	x3 := (a&0xFFFFFFFF)*(b>>32) + (x2 & 0xFFFFFFFF)
	var result UInt128Struct
	result.Lo64 = ((x3 & 0xFFFFFFFF) << 32) | (x1 & 0xFFFFFFFF)
	result.Hi64 = (a>>32)*(b>>32) + (x2 >> 32) + (x3 >> 32)
	return result
}

// productsAreEqual returns true iff a*b == c*d (exactly) using 128-bit intermediate
func productsAreEqual(a, b, c, d int64) bool {
	absA := uint64(math.Abs(float64(a)))
	absB := uint64(math.Abs(float64(b)))
	absC := uint64(math.Abs(float64(c)))
	absD := uint64(math.Abs(float64(d)))

	mulAB := multiplyUInt64(absA, absB)
	mulCD := multiplyUInt64(absC, absD)

	signAB := triSign(a) * triSign(b)
	signCD := triSign(c) * triSign(d)

	return mulAB.Lo64 == mulCD.Lo64 &&
		mulAB.Hi64 == mulCD.Hi64 && signAB == signCD
}

func isCollinear(pt1, sharedPt, pt2 Point64) bool {
	a := sharedPt.X - pt1.X
	b := pt2.Y - sharedPt.Y
	c := sharedPt.Y - pt1.Y
	d := pt2.X - sharedPt.X

	return productsAreEqual(a, b, c, d)
}

func dotProduct64(pt1, pt2, pt3 Point64) float64 {
	return float64((pt2.X-pt1.X)*(pt3.X-pt2.X) + (pt2.Y-pt1.Y)*(pt3.Y-pt2.Y))
}

func crossProductD(vec1, vec2 PointD) float64 {
	return vec1.Y*vec2.X - vec2.Y*vec1.X
}

func dotProductD(vec1, vec2 PointD) float64 {
	return vec1.X*vec2.X + vec1.Y*vec2.Y
}

func checkCastInt64(val float64) int64 {
	if val >= max_coord || val <= min_coord {
		return Invalid64
	}
	// Round away from zero like MidpointRounding.AwayFromZero
	if val >= 0 {
		return int64(math.Floor(val + 0.5))
	}
	return int64(math.Ceil(val - 0.5))
}

// getSegmentIntersectPt computes intersection point ip of two segments ln1 (ln1a->ln1b) and ln2.
// Returns (true, ip) if lines intersect (including endpoints); if parallel, returns (false, zeroPt)
func getSegmentIntersectPt(ln1a, ln1b, ln2a, ln2b Point64) (Point64, bool) {
	dy1 := ln1b.Y - ln1a.Y
	dx1 := ln1b.X - ln1a.X
	dy2 := ln2b.Y - ln2a.Y
	dx2 := ln2b.X - ln2a.X
	det := dy1*dx2 - dy2*dx1
	var ip Point64
	if det == 0 {
		return ip, false
	}

	t := float64(((ln1a.X-ln2a.X)*dy2)-((ln1a.Y-ln2a.Y)*dx2)) / float64(det)
	if t <= 0 {
		ip = ln1a
	} else if t >= 1 {
		ip = ln1b
	} else {
		ip.X = ln1a.X + int64(t*float64(dx1))
		ip.Y = ln1a.Y + int64(t*float64(dy1))
	}
	return ip, true
}

func segsIntersect(seg1a, seg1b, seg2a, seg2b Point64, inclusive bool) bool {
	if !inclusive {
		return (CrossProduct(seg1a, seg2a, seg2b)*CrossProduct(seg1b, seg2a, seg2b) < 0) &&
			(CrossProduct(seg2a, seg1a, seg1b)*CrossProduct(seg2b, seg1a, seg1b) < 0)
	}
	res1 := CrossProduct(seg1a, seg2a, seg2b)
	res2 := CrossProduct(seg1b, seg2a, seg2b)
	if res1*res2 > 0 {
		return false
	}
	res3 := CrossProduct(seg2a, seg1a, seg1b)
	res4 := CrossProduct(seg2b, seg1a, seg1b)
	if res3*res4 > 0 {
		return false
	}

	// ensure NOT collinear
	return res1 != 0 || res2 != 0 || res3 != 0 || res4 != 0
}

func getBounds(path Path64) Rect64 {
	if len(path) == 0 {
		return Rect64{}
	}
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
	return result
}

func getClosestPtOnSegment(offPt, seg1, seg2 Point64) Point64 {
	if seg1.X == seg2.X && seg1.Y == seg2.Y {
		return seg1
	}
	dx := float64(seg2.X - seg1.X)
	dy := float64(seg2.Y - seg1.Y)
	q := ((float64(offPt.X-seg1.X) * dx) + (float64(offPt.Y-seg1.Y) * dy)) / ((dx * dx) + (dy * dy))
	if q < 0 {
		q = 0
	} else if q > 1 {
		q = 1
	}
	rx := seg1.X + int64(roundToEven(q*dx))
	ry := seg1.Y + int64(roundToEven(q*dy))
	return Point64{X: rx, Y: ry}
}

func roundToEven(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	iv, frac := math.Modf(v)
	absFrac := math.Abs(frac)
	if absFrac < 0.5 {
		return iv
	} else if absFrac > 0.5 {
		if v > 0 {
			return iv + 1
		}
		return iv - 1
	} else {
		// exactly .5 -> choose even
		if int64(iv)%2 == 0 {
			return iv
		}
		if v > 0 {
			return iv + 1
		}
		return iv - 1
	}
}

func PointInPolygon(pt Point64, polygon Path64) PointInPolygonResult {
	lenP := len(polygon)
	if lenP < 3 {
		return IsOutside
	}
	start := 0
	for start < lenP && polygon[start].Y == pt.Y {
		start++
	}
	if start == lenP {
		return IsOutside
	}
	isAbove := polygon[start].Y < pt.Y
	startingAbove := isAbove
	val := 0
	i := start + 1
	end := lenP
	for {
		if i == end {
			if end == 0 || start == 0 {
				break
			}
			end = start
			i = 0
		}
		if isAbove {
			for i < end && polygon[i].Y < pt.Y {
				i++
			}
		} else {
			for i < end && polygon[i].Y > pt.Y {
				i++
			}
		}
		if i == end {
			continue
		}
		var curr, prev Point64
		curr = polygon[i]
		if i > 0 {
			prev = polygon[i-1]
		} else {
			prev = polygon[lenP-1]
		}
		if curr.Y == pt.Y {
			if curr.X == pt.X || (curr.Y == prev.Y && ((pt.X < prev.X) != (pt.X < curr.X))) {
				return IsOn
			}
			i++
			if i == start {
				break
			}
			continue
		}
		if pt.X < curr.X && pt.X < prev.X {
			// only interested in edges crossing on the left
		} else if pt.X > prev.X && pt.X > curr.X {
			val = 1 - val
		} else {
			d := CrossProduct(prev, curr, pt)
			if d == 0 {
				return IsOn
			}
			if (d < 0) == isAbove {
				val = 1 - val
			}
		}
		isAbove = !isAbove
		i++
	}
	if isAbove == startingAbove {
		if val == 0 {
			return IsOutside
		}
		return IsInside
	}
	if i == lenP {
		i = 0
	}
	var d float64
	if i == 0 {
		d = CrossProduct(polygon[lenP-1], polygon[0], pt)
	} else {
		d = CrossProduct(polygon[i-1], polygon[i], pt)
	}
	if d == 0 {
		return IsOn
	}
	if (d < 0) == isAbove {
		val = 1 - val
	}
	if val == 0 {
		return IsOutside
	}
	return IsInside
}

func Path2ContainsPath1(path1, path2 Path64) bool {
	pip := IsOn
	for _, pt := range path1 {
		switch PointInPolygon(pt, path2) {
		case IsOutside:
			if pip == IsOutside {
				return false
			}
			pip = IsOutside
		case IsInside:
			if pip == IsInside {
				return true
			}
			pip = IsInside
		default:
			// do nothing
		}
	}

	bounds := getBounds(path1)
	mp := bounds.MidPoint()
	return PointInPolygon(mp, path2) != IsOutside
}

func pointInOpPolygon(pt Point64, op *OutPt) PointInPolygonResult {
	// degenerate polygon
	if op == op.next || op.prev == op.next {
		return IsOutside
	}
	op2 := op
	for {
		if op.pt.Y != pt.Y {
			break
		}
		op = op.next
		if op == op2 {
			break
		}
	}
	if op.pt.Y == pt.Y { // not a proper polygon
		return IsOutside
	}

	// must be above or below to get here
	isAbove := op.pt.Y < pt.Y
	startingAbove := isAbove
	val := 0

	op2 = op.next
	for op2 != op {
		if isAbove {
			for op2 != op && op2.pt.Y < pt.Y {
				op2 = op2.next
			}
		} else {
			for op2 != op && op2.pt.Y > pt.Y {
				op2 = op2.next
			}
		}
		if op2 == op {
			break
		}

		// touching the horizontal
		if op2.pt.Y == pt.Y {
			if op2.pt.X == pt.X || (op2.pt.Y == op2.prev.pt.Y &&
				(pt.X < op2.prev.pt.X) != (pt.X < op2.pt.X)) {
				return IsOn
			}
			op2 = op2.next
			if op2 == op {
				break
			}
			continue
		}

		if op2.pt.X <= pt.X || op2.prev.pt.X <= pt.X {
			if op2.prev.pt.X < pt.X && op2.pt.X < pt.X {
				val = 1 - val
			} else {
				d := CrossProduct(op2.prev.pt, op2.pt, pt)
				if d == 0 {
					return IsOn
				}
				if (d < 0) == isAbove {
					val = 1 - val
				}
			}
		}
		isAbove = !isAbove
		op2 = op2.next
	}

	if isAbove == startingAbove {
		if val == 0 {
			return IsOutside
		}
		return IsInside
	}

	d := CrossProduct(op2.prev.pt, op2.pt, pt)
	if d == 0 {
		return IsOn
	}
	if (d < 0) == isAbove {
		val = 1 - val
	}
	if val == 0 {
		return IsOutside
	}
	return IsInside
}
