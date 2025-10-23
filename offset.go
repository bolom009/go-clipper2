package go_clipper2

import (
	"math"
)

const (
	Tolerance float64 = 1.0e-12
	arc               = 0.002
)

type JoinType uint8

const (
	Miter JoinType = iota
	Square
	Bevel
	Round
)

type EndType uint8

const (
	Polygon EndType = iota
	Joined
	Butt
	SquareET
	RoundET
)

type InflateOption func(*inflateConfig)

type inflateConfig struct {
	miterLimit   float64
	arcTolerance float64
	precision    int
}

func WithMitterLimit(limit float64) InflateOption {
	return func(config *inflateConfig) {
		config.miterLimit = limit
	}
}

func WithArcTolerance(tolerance float64) InflateOption {
	return func(config *inflateConfig) {
		config.arcTolerance = tolerance
	}
}

func WithPrecision(precision int) InflateOption {
	return func(config *inflateConfig) {
		config.precision = precision
	}
}

func InflatePaths64(paths Paths64, delta float64, joinType JoinType, endType EndType, opts ...InflateOption) Paths64 {
	cfg := &inflateConfig{
		miterLimit:   2.0,
		arcTolerance: 0,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	co := NewClipperOffset(cfg.miterLimit, cfg.arcTolerance, false, false)
	co.AddPaths(paths, joinType, endType)
	solution := make(Paths64, 0)
	co.Execute64(delta, &solution)
	return solution
}

func InflatePathsD(paths PathsD, delta float64, joinType JoinType, endType EndType, opts ...InflateOption) PathsD {
	cfg := &inflateConfig{
		miterLimit:   2.0,
		arcTolerance: 0,
		precision:    2,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// panic if wrong precision
	checkPrecision(cfg.precision)

	scale := math.Pow(10, float64(cfg.precision))
	tmp := ScalePathsDToPaths64(paths, scale)

	co := NewClipperOffset(cfg.miterLimit, scale*cfg.arcTolerance, false, false)
	co.AddPaths(tmp, joinType, endType)
	co.Execute64(delta*scale, &tmp)

	return ScalePaths64ToPathsD(tmp, 1.0/scale)
}

type Group struct {
	inPaths       Paths64
	joinType      JoinType
	endType       EndType
	pathsReversed bool
	lowestPathIdx int
}

func NewGroup(paths Paths64, joinType JoinType, endTypeVal ...EndType) *Group {
	endType := Polygon
	if len(endTypeVal) > 0 {
		endType = endTypeVal[0]
	}

	group := &Group{
		joinType: joinType,
		endType:  endType,
	}
	isGroupJoined := endType == Polygon || endType == Joined

	group.inPaths = make(Paths64, 0, len(paths))
	for _, path := range paths {
		group.inPaths = append(group.inPaths, StripDuplicates(path, isGroupJoined))
	}

	if endType == Polygon {
		var isNegArea bool
		group.lowestPathIdx, isNegArea = group.GetLowestPathInfo()
		group.pathsReversed = (group.lowestPathIdx >= 0) && isNegArea
	} else {
		group.lowestPathIdx = -1
		group.pathsReversed = false
	}

	return group
}

func (g *Group) GetLowestPathInfo() (int, bool) {
	idx := -1
	isNegArea := false
	botPt := Point64{X: math.MaxInt64, Y: math.MinInt64}

	for i, path := range g.inPaths {
		a := math.MaxFloat64
		for _, pt := range path {
			if pt.Y < botPt.Y || (pt.Y == botPt.Y && pt.X >= botPt.X) {
				continue
			}
			if a == math.MaxFloat64 {
				a = Area64(path)
				if a == 0 {
					break
				}
				isNegArea = a < 0
			}
			idx = i
			botPt.X = pt.X
			botPt.Y = pt.Y
		}
	}
	return idx, isNegArea
}

type ClipperOffset struct {
	ArcTolerance      float64
	MergeGroups       bool
	MiterLimit        float64
	PreserveCollinear bool
	ReverseSolution   bool

	groupList   []*Group
	pathOut     Path64
	normals     PathD
	solution    *Paths64
	groupDelta  float64
	delta       float64
	mitLimSqr   float64
	stepsPerRad float64
	stepSin     float64
	stepCos     float64
	joinType    JoinType
	endType     EndType
}

func NewClipperOffset(miterLimit, arcTolerance float64, preserveCollinear, reverseSolution bool) *ClipperOffset {
	if miterLimit == 0 {
		miterLimit = 2.0
	}

	return &ClipperOffset{
		MiterLimit:        miterLimit,
		ArcTolerance:      arcTolerance,
		MergeGroups:       true,
		PreserveCollinear: preserveCollinear,
		ReverseSolution:   reverseSolution,
	}
}

func (co *ClipperOffset) AddPaths(paths Paths64, joinType JoinType, endType EndType) {
	if len(paths) == 0 {
		return
	}

	co.groupList = append(co.groupList, NewGroup(paths, joinType, endType))
}

func (co *ClipperOffset) CalcSolutionCapacity() int {
	result := 0
	for _, g := range co.groupList {
		if g.endType == Joined {
			result += len(g.inPaths) * 2
		} else {
			result += len(g.inPaths)
		}
	}
	return result
}

func (co *ClipperOffset) CheckPathsReversed() bool {
	for _, g := range co.groupList {
		if g.endType == Polygon {
			return g.pathsReversed
		}
	}
	return false
}

func (co *ClipperOffset) Execute64(delta float64, solution *Paths64) {
	co.solution = solution
	co.executeInternal(delta)
}

func (co *ClipperOffset) executeInternal(delta float64) {
	if len(co.groupList) == 0 {
		return
	}

	capacity := co.CalcSolutionCapacity()
	if nCap := cap(*co.solution); nCap < capacity {
		*co.solution = make(Paths64, 0, capacity)
	} else {
		*co.solution = (*co.solution)[:0]
	}

	if math.Abs(delta) < 0.5 {
		for _, group := range co.groupList {
			for _, path := range group.inPaths {
				*co.solution = append(*co.solution, path)
			}
		}
		return
	}

	co.delta = delta

	if co.MiterLimit <= 1 {
		co.mitLimSqr = 2.0
	} else {
		co.mitLimSqr = 2.0 / sqr(co.MiterLimit)
	}

	for _, group := range co.groupList {
		co.doGroupOffset(group)
	}

	if len(co.groupList) == 0 {
		return
	}

	pathsReversed := co.checkPathsReversed()
	var fillRule FillRule
	if pathsReversed {
		fillRule = Negative
	} else {
		fillRule = Positive
	}

	c := NewClipper64()
	c.preserveCollinear = co.PreserveCollinear
	c.reverseSolution = co.ReverseSolution != pathsReversed

	c.addSubject(*co.solution)

	//if co.solutionTree != nil {
	//	c.Execute(Union, fillRule, co.solutionTree)
	//} else {
	c.Execute(Union, fillRule, co.solution)
	//}
}

func (co *ClipperOffset) doGroupOffset(group *Group) {
	if group.endType == Polygon {
		if group.lowestPathIdx < 0 {
			co.delta = math.Abs(co.delta)
		}
		if group.pathsReversed {
			co.groupDelta = -co.delta
		} else {
			co.groupDelta = co.delta
		}
	} else {
		co.groupDelta = math.Abs(co.delta)
	}

	absDelta := math.Abs(co.groupDelta)

	co.joinType = group.joinType
	co.endType = group.endType

	if group.joinType == Round || group.endType == RoundET {
		arcTol := absDelta * arc
		if co.ArcTolerance > 0.01 {
			arcTol = co.ArcTolerance
		}
		stepsPer360 := math.Pi / math.Acos(1-arcTol/absDelta)
		co.stepSin = math.Sin((2 * math.Pi) / stepsPer360)
		co.stepCos = math.Cos((2 * math.Pi) / stepsPer360)
		if co.groupDelta < 0 {
			co.stepSin = -co.stepSin
		}
		co.stepsPerRad = stepsPer360 / (2 * math.Pi)
	}

	for _, p := range group.inPaths {
		co.pathOut = Path64{}
		cnt := len(p)

		switch cnt {
		case 1:
			pt := p[0]
			if group.endType == RoundET {
				steps := int(math.Ceil(co.stepsPerRad * 2 * math.Pi))
				co.pathOut = Ellipse64(pt, absDelta, absDelta, steps)
			} else {
				d := int64(math.Ceil(co.groupDelta))
				r := Rect64{
					pt.X - d,
					pt.Y - d,
					pt.X + d,
					pt.Y + d,
				}
				co.pathOut = r.AsPath()
			}

			*co.solution = append(*co.solution, co.pathOut)
			continue
		case 2:
			if group.endType == Joined {
				if group.joinType == Round {
					co.endType = RoundET
				} else {
					co.endType = SquareET
				}
			}
		}

		co.buildNormals(p)
		switch co.endType {
		case Polygon:
			co.offsetPolygon(group, p)
		case Joined:
			co.offsetOpenJoined(group, p)
		default:
			co.offsetOpenPath(group, p)
		}
	}
}

func (co *ClipperOffset) buildNormals(path Path64) {
	cnt := len(path)
	co.normals = co.normals[:0]
	if cnt == 0 {
		return
	}
	if cap(co.normals) < cnt {
		co.normals = make([]PointD, 0, cnt)
	}
	for i := 0; i < cnt-1; i++ {
		co.normals = append(co.normals, getUnitNormal(path[i], path[i+1]))
	}

	co.normals = append(co.normals, getUnitNormal(path[cnt-1], path[0]))
}

func getUnitNormal(pt1, pt2 Point64) PointD {
	dx := float64(pt2.X - pt1.X)
	dy := float64(pt2.Y - pt1.Y)
	if dx == 0 && dy == 0 {
		return PointD{X: 0, Y: 0}
	}
	f := 1.0 / math.Sqrt(dx*dx+dy*dy)
	dx *= f
	dy *= f
	return PointD{X: dy, Y: -dx}
}

func (co *ClipperOffset) offsetPolygon(group *Group, path Path64) {
	co.pathOut = Path64{}
	cnt := len(path)
	prev := cnt - 1
	for i := 0; i < cnt; i++ {
		co.offsetPoint(group, path, i, &prev)
	}

	*co.solution = append(*co.solution, co.pathOut)
}

func (co *ClipperOffset) offsetOpenJoined(group *Group, path Path64) {
	co.offsetPolygon(group, path)
	rPath := ReversePath(path)
	co.buildNormals(rPath)
	co.offsetPolygon(group, rPath)
}

func (co *ClipperOffset) offsetOpenPath(group *Group, path Path64) {
	co.pathOut = nil
	highI := len(path) - 1
	var delta float64

	if math.Abs(delta) < Tolerance {
		co.pathOut = append(co.pathOut, path[0])
	} else {
		switch co.endType {
		case Butt:
			co.doBevel(path, 0, 0)
		case RoundET:
			co.doRound(path, 0, 0, math.Pi)
		default:
			co.doSquare(path, 0, 0)
		}
	}

	k := 0
	for i := 1; i < highI; i++ {
		co.offsetPoint(group, path, i, &k)
	}

	for i := highI; i > 0; i-- {
		co.normals[i] = PointD{-co.normals[i-1].X, -co.normals[i-1].Y}
	}
	if highI >= 0 {
		co.normals[0] = co.normals[highI]
	}

	if math.Abs(delta) < Tolerance {
		co.pathOut = append(co.pathOut, path[highI])
	} else {
		switch co.endType {
		case Butt:
			co.doBevel(path, highI, highI)
		case RoundET:
			co.doRound(path, highI, highI, math.Pi)
		default:
			co.doSquare(path, highI, highI)
		}
	}

	k = highI
	for i := highI - 1; i > 0; i-- {
		co.offsetPoint(group, path, i, &k)
	}

	*co.solution = append(*co.solution, co.pathOut)
}

func (co *ClipperOffset) offsetPoint(_ *Group, path Path64, j int, k *int) {
	if path[j] == path[*k] {
		*k = j
		return
	}

	sinA := crossProductD(co.normals[j], co.normals[*k])
	cosA := dotProductD(co.normals[j], co.normals[*k])

	if sinA > 1.0 {
		sinA = 1.0
	} else if sinA < -1.0 {
		sinA = -1.0
	}

	if math.Abs(co.groupDelta) < Tolerance {
		co.pathOut = append(co.pathOut, path[j])
		return
	}

	if cosA > -0.999 && (sinA*co.groupDelta < 0) {
		co.pathOut = append(co.pathOut, co.getPerpendic(path[j], co.normals[*k]))
		co.pathOut = append(co.pathOut, path[j])
		co.pathOut = append(co.pathOut, co.getPerpendic(path[j], co.normals[j]))
	} else if cosA > 0.999 && co.joinType != Round {
		co.doMiter(path, j, *k, cosA)
	} else {
		switch co.joinType {
		case Miter:
			if cosA > co.mitLimSqr-1 {
				co.doMiter(path, j, *k, cosA)
			} else {
				co.doSquare(path, j, *k)
			}
		case Round:
			co.doRound(path, j, *k, math.Atan2(sinA, cosA))
		case Bevel:
			co.doBevel(path, j, *k)
		default:
			co.doSquare(path, j, *k)
		}
	}
	*k = j
}

func (co *ClipperOffset) getPerpendic(pt Point64, norm PointD) Point64 {
	offsetX := float64(pt.X) + norm.X*co.groupDelta
	offsetY := float64(pt.Y) + norm.Y*co.groupDelta
	return Point64{
		X: int64(math.Round(offsetX)),
		Y: int64(math.Round(offsetY)),
	}
}

func (co *ClipperOffset) getPerpendicD(pt Point64, norm PointD) PointD {
	x := float64(pt.X) + norm.X*co.groupDelta
	y := float64(pt.Y) + norm.Y*co.groupDelta
	return PointD{X: x, Y: y}
}

func (co *ClipperOffset) checkPathsReversed() bool {
	for _, g := range co.groupList {
		if g.endType == Polygon {
			return g.pathsReversed
		}
	}
	return false
}

func (co *ClipperOffset) doMiter(path Path64, j, k int, cosA float64) {
	q := co.groupDelta / (cosA + 1)

	result := Point64{
		X: int64(math.Round(float64(path[j].X) + (co.normals[k].X+co.normals[j].X)*q)),
		Y: int64(math.Round(float64(path[j].Y) + (co.normals[k].Y+co.normals[j].Y)*q)),
	}

	co.pathOut = append(co.pathOut, result)
}

func (co *ClipperOffset) doBevel(path Path64, j, k int) {
	var pt1, pt2 Point64
	absDelta := math.Abs(co.groupDelta)

	if j == k {
		pt1 = Point64{
			X: int64(float64(path[j].X) - absDelta*co.normals[j].X),
			Y: int64(float64(path[j].Y) - absDelta*co.normals[j].Y),
		}
		pt2 = Point64{
			X: int64(float64(path[j].X) + absDelta*co.normals[j].X),
			Y: int64(float64(path[j].Y) + absDelta*co.normals[j].Y),
		}
	} else {
		pt1 = Point64{
			X: int64(float64(path[j].X) + co.groupDelta*co.normals[k].X),
			Y: int64(float64(path[j].Y) + co.groupDelta*co.normals[k].Y),
		}
		pt2 = Point64{
			X: int64(float64(path[j].X) + co.groupDelta*co.normals[j].X),
			Y: int64(float64(path[j].Y) + co.groupDelta*co.normals[j].Y),
		}
	}

	co.pathOut = append(co.pathOut, pt1, pt2)
}

func (co *ClipperOffset) doRound(path Path64, j, k int, angle float64) {
	pt := path[j]
	offsetVec := PointD{
		X: co.normals[k].X * co.groupDelta,
		Y: co.normals[k].Y * co.groupDelta,
	}
	if j == k {
		offsetVec.X = -offsetVec.X
		offsetVec.Y = -offsetVec.Y
	}

	co.pathOut = append(co.pathOut, Point64{
		X: int64(math.Round(float64(pt.X) + offsetVec.X)),
		Y: int64(math.Round(float64(pt.Y) + offsetVec.Y)),
	})

	steps := int(math.Ceil(co.stepsPerRad * math.Abs(angle)))
	for i := 1; i < steps; i++ {
		newX := offsetVec.X*co.stepCos - co.stepSin*offsetVec.Y
		newY := offsetVec.X*co.stepSin + offsetVec.Y*co.stepCos

		offsetVec = PointD{X: newX, Y: newY}

		co.pathOut = append(co.pathOut, Point64{
			X: int64(math.Round(float64(pt.X) + offsetVec.X)),
			Y: int64(math.Round(float64(pt.Y) + offsetVec.Y)),
		})
	}

	co.pathOut = append(co.pathOut, co.getPerpendic(pt, co.normals[j]))
}

func (co *ClipperOffset) doSquare(path Path64, j, k int) {
	var vec PointD
	if j == k {
		vec = PointD{
			X: co.normals[j].Y,
			Y: -co.normals[j].X,
		}
	} else {
		vec = getAvgUnitVector(
			PointD{
				X: -co.normals[k].Y,
				Y: co.normals[k].X,
			},
			PointD{
				X: co.normals[j].Y,
				Y: -co.normals[j].X,
			},
		)
	}

	absDelta := math.Abs(co.groupDelta)

	ptQ := translatePoint(PointD{X: float64(path[j].X), Y: float64(path[j].Y)}, absDelta*vec.X, absDelta*vec.Y)
	pt1 := translatePoint(ptQ, co.groupDelta*vec.Y, -co.groupDelta*vec.X)
	pt2 := translatePoint(ptQ, -co.groupDelta*vec.Y, co.groupDelta*vec.X)
	pt3 := co.getPerpendicD(path[k], co.normals[k])

	if j == k {
		pt4 := PointD{
			X: pt3.X + vec.X*co.groupDelta,
			Y: pt3.Y + vec.Y*co.groupDelta,
		}
		pt := intersectPoint(pt1, pt2, pt3, pt4)

		rp := reflectPoint(pt, ptQ)
		co.pathOut = append(co.pathOut, rp.ToPoint64())
		co.pathOut = append(co.pathOut, pt.ToPoint64())
	} else {
		pt4 := co.getPerpendicD(path[j], co.normals[k])
		pt := intersectPoint(pt1, pt2, pt3, pt4)

		co.pathOut = append(co.pathOut, pt.ToPoint64())
		rp := reflectPoint(pt, ptQ)
		co.pathOut = append(co.pathOut, rp.ToPoint64())
	}
}

func intersectPoint(pt1a, pt1b, pt2a, pt2b PointD) PointD {
	if isAlmostZero(pt1a.X - pt1b.X) {
		if isAlmostZero(pt2a.X - pt2b.X) {

			return PointD{X: 0, Y: 0}
		}
		m2 := (pt2b.Y - pt2a.Y) / (pt2b.X - pt2a.X)
		b2 := pt2a.Y - m2*pt2a.X
		return PointD{X: pt1a.X, Y: m2*pt1a.X + b2}
	}

	if isAlmostZero(pt2a.X - pt2b.X) {
		m1 := (pt1b.Y - pt1a.Y) / (pt1b.X - pt1a.X)
		b1 := pt1a.Y - m1*pt1a.X
		return PointD{X: pt2a.X, Y: m1*pt2a.X + b1}
	}

	m1 := (pt1b.Y - pt1a.Y) / (pt1b.X - pt1a.X)
	b1 := pt1a.Y - m1*pt1a.X

	m2 := (pt2b.Y - pt2a.Y) / (pt2b.X - pt2a.X)
	b2 := pt2a.Y - m2*pt2a.X

	if isAlmostZero(m1 - m2) {
		return PointD{X: 0, Y: 0}
	}

	x := (b2 - b1) / (m1 - m2)
	y := m1*x + b1
	return PointD{X: x, Y: y}
}

func almostZero(value float64) bool {
	return math.Abs(value) < 0.001
}

func hypotenuse(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}

func translatePoint(pt PointD, dx, dy float64) PointD {
	return PointD{
		X: pt.X + dx,
		Y: pt.Y + dy,
	}
}

func reflectPoint(pt, pivot PointD) PointD {
	return PointD{
		X: pivot.X + (pivot.X - pt.X),
		Y: pivot.Y + (pivot.Y - pt.Y),
	}
}

func getAvgUnitVector(vec1, vec2 PointD) PointD {
	return normalizeVector(PointD{
		X: vec1.X + vec2.X,
		Y: vec1.Y + vec2.Y,
	})
}

func normalizeVector(vec PointD) PointD {
	h := hypotenuse(vec.X, vec.Y)
	if almostZero(h) {
		return PointD{X: 0, Y: 0}
	}

	inverseHypot := 1 / h
	return PointD{
		X: vec.X * inverseHypot,
		Y: vec.Y * inverseHypot,
	}
}
