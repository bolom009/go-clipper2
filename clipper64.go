package go_clipper2

type Clipper64 struct {
	*ClipperBase
}

func NewClipper64() *Clipper64 {
	return &Clipper64{
		NewClipperBase(),
	}
}

func (c *Clipper64) Execute(clipType ClipType, fillRule FillRule, solution *Paths64) bool {
	solOpen := make(Paths64, 0)

	return c.ExecuteOC(clipType, fillRule, solution, &solOpen)
}

func (c *Clipper64) ExecuteOC(clipType ClipType, fillRule FillRule, solutionClosed, solutionOpen *Paths64) bool {
	*solutionClosed = (*solutionClosed)[:0]
	*solutionOpen = (*solutionOpen)[:0]

	success := c.ClipperBase.execute(clipType, fillRule, solutionClosed, solutionOpen)

	c.clearSolutionOnly()
	return success
}

func (c *Clipper64) AddPaths(paths Paths64, polytype PathType, isOpen bool) {
	c.addPaths(paths, polytype, isOpen)
}
