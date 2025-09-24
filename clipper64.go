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
