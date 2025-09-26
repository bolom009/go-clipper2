package go_clipper2

import "math"

type clipperD struct {
	*clipperBase

	scale    float64
	invScale float64
}

func NewClipperD(decimalPrecision int) *clipperD {
	if decimalPrecision == 0 {
		decimalPrecision = 2
	}

	if decimalPrecision < -8 || decimalPrecision > 8 {
		panic(ErrPrecisionRange)
	}

	scale := math.Pow(10, float64(decimalPrecision))

	return &clipperD{
		clipperBase: newClipperBase(),
		scale:       scale,
		invScale:    1 / scale,
	}
}

func (c *clipperD) AddPaths(paths PathsD, polytype PathType, isOpen bool) {
	c.addPaths(ScalePathsDToPaths64(paths, c.scale), polytype, isOpen)
}

func (c *clipperD) Execute(clipType ClipType, fillRule FillRule, solution *PathsD) bool {
	solOpen := make(PathsD, 0)

	return c.ExecuteOC(clipType, fillRule, solution, &solOpen)
}

func (c *clipperD) ExecuteOC(clipType ClipType, fillRule FillRule, solutionClosed, solutionOpen *PathsD) bool {
	solClosed64 := make(Paths64, 0)
	solOpen64 := make(Paths64, 0)

	success := c.clipperBase.execute(clipType, fillRule, &solClosed64, &solOpen64)

	c.clearSolutionOnly()
	if !success {
		return false
	}

	for _, path := range solClosed64 {
		*solutionClosed = append(*solutionClosed, ScalePath64ToPathD(path, c.invScale))
	}

	for _, path := range solOpen64 {
		*solutionOpen = append(*solutionOpen, ScalePath64ToPathD(path, c.invScale))
	}

	return true
}

func UnionPathsD(subject PathsD, fillRule FillRule, precision ...int) PathsD {
	return BooleanOpPathsD(Union, subject, nil, fillRule, precision...)
}

func UnionWithClipPathsD(subject, clip PathsD, fillRule FillRule, precision ...int) PathsD {
	return BooleanOpPathsD(Union, subject, clip, fillRule, precision...)
}

func IntersectWithClipPathsD(subject, clip PathsD, fillRule FillRule, precision ...int) PathsD {
	return BooleanOpPathsD(Intersection, subject, clip, fillRule, precision...)
}

func DifferenceWithClipPathsD(subject, clip PathsD, fillRule FillRule, precision ...int) PathsD {
	return BooleanOpPathsD(Difference, subject, clip, fillRule, precision...)
}

func XorWithClipPathsD(subject, clip PathsD, fillRule FillRule, precision ...int) PathsD {
	return BooleanOpPathsD(Xor, subject, clip, fillRule, precision...)
}

func BooleanOpPathsD(clipType ClipType, subject PathsD, clip PathsD, fillRule FillRule, precisionV ...int) PathsD {
	precision := 2
	if len(precisionV) > 0 {
		precision = precisionV[0]
	}

	solution := make(PathsD, 0)
	c := NewClipperD(precision)
	c.AddPaths(subject, Subject, false)
	if clip != nil {
		c.AddPaths(clip, Clip, false)
	}

	c.Execute(clipType, fillRule, &solution)
	return solution
}
