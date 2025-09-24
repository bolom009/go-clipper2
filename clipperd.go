package go_clipper2

import "math"

type clipperD struct {
	*clipperBase

	scale    float64
	invScale float64
}

func NewClipperD(roundingDecimalPrecision int) *clipperD {
	if roundingDecimalPrecision == 0 {
		roundingDecimalPrecision = 2
	}

	if roundingDecimalPrecision < -8 || roundingDecimalPrecision > 8 {
		panic("precision is out of range")
	}

	scale := math.Pow(10, float64(roundingDecimalPrecision))

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
