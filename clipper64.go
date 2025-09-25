package go_clipper2

type clipper64 struct {
	*clipperBase
}

func NewClipper64() *clipper64 {
	return &clipper64{
		newClipperBase(),
	}
}

func (c *clipper64) Execute(clipType ClipType, fillRule FillRule, solution *Paths64) bool {
	solOpen := make(Paths64, 0)

	return c.ExecuteOC(clipType, fillRule, solution, &solOpen)
}

func (c *clipper64) ExecuteOC(clipType ClipType, fillRule FillRule, solutionClosed, solutionOpen *Paths64) bool {
	*solutionClosed = (*solutionClosed)[:0]
	*solutionOpen = (*solutionOpen)[:0]

	success := c.clipperBase.execute(clipType, fillRule, solutionClosed, solutionOpen)

	c.clearSolutionOnly()
	return success
}

func (c *clipper64) AddPaths(paths Paths64, polytype PathType, isOpen bool) {
	c.addPaths(paths, polytype, isOpen)
}

func UnionPaths64(subject Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Union, subject, nil, fillRule)
}

func UnionWithClipPaths64(subject, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Union, subject, clip, fillRule)
}

func IntersectWithClipPaths64(subject, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Intersection, subject, clip, fillRule)
}

func DifferenceWithClipPaths64(subject, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Difference, subject, clip, fillRule)
}

func XorWithClipPaths64(subject, clip Paths64, fillRule FillRule) Paths64 {
	return BooleanOpPaths64(Xor, subject, clip, fillRule)
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
