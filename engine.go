package go_clipper2

import (
	"math"
)

var (
	negInf = math.Inf(-1)
	posInf = math.Inf(1)
)

type PointInPolygonResult uint8

const (
	IsOn PointInPolygonResult = iota
	IsInside
	IsOutside
)

type VertexFlags uint8

const (
	None      VertexFlags = 0
	OpenStart             = 1
	OpenEnd               = 2
	LocalMax              = 4
	LocalMin              = 8
)

type JoinWith uint8

const (
	JoinNone JoinWith = iota
	JoinLeft
	JoinRight
)

type OutRec struct {
	idx            int
	owner          *OutRec
	frontEdge      *Active
	backEdge       *Active
	pts            *OutPt
	polypath       *PolyPathBase
	bounds         Rect64
	path           Path64
	isOpen         bool
	splits         []int
	recursiveSplit *OutRec
}

type HorzSegment struct {
	leftOp      *OutPt
	rightOp     *OutPt
	leftToRight bool
}

func NewHorzSegment(op *OutPt) *HorzSegment {
	return &HorzSegment{
		leftOp:      op,
		rightOp:     nil,
		leftToRight: true,
	}
}

type HorzJoin struct {
	op1 *OutPt
	op2 *OutPt
}

func NewHorzJoin(ltor *OutPt, rtol *OutPt) *HorzJoin {
	return &HorzJoin{
		op1: ltor,
		op2: rtol,
	}
}

type Active struct {
	bot        Point64
	top        Point64
	curX       int64 // current (updated at every new scanline)
	dx         float64
	windDx     int // 1 or -1 depending on winding direction
	windCount  int
	windCount2 int // winding count of the opposite polytype
	outrec     *OutRec

	// AEL: 'active edge list' (Vatti's AET - active edge table)
	//     a linked list of all edges (from left to right) that are present
	//     (or 'active') within the current scanbeam (a horizontal 'beam' that
	//     sweeps from bottom to top over the paths in the clipping operation).
	prevInAEL *Active
	nextInAEL *Active

	// SEL: 'sorted edge list' (Vatti's ST - sorted table)
	//     linked list used when sorting edges into their new positions at the
	//     top of scanbeams, but also (re)used to process horizontals.
	prevInSEL   *Active
	nextInSEL   *Active
	jump        *Active
	vertexTop   *Vertex
	localMin    *LocalMinima // the bottom of an edge 'bound' (also Vatti)
	isLeftBound bool
	joinWith    JoinWith
}

type OutPt struct {
	pt     Point64
	next   *OutPt
	prev   *OutPt
	outrec *OutRec
	horz   *HorzSegment
}

func newOutPt(pt Point64, outrec *OutRec) *OutPt {
	out := &OutPt{
		pt:     pt,
		outrec: outrec,
		horz:   nil,
	}

	out.prev = out
	out.next = out

	return out
}

type Vertex struct {
	pt    Point64
	prev  *Vertex
	next  *Vertex
	flags VertexFlags
}

func NewVertex(pt Point64, flags VertexFlags, prev *Vertex) *Vertex {
	return &Vertex{
		pt:    pt,
		prev:  prev,
		next:  nil,
		flags: flags,
	}
}

type VertexPoolList []*Vertex

func (vpl *VertexPoolList) EnsureCapacity(nCap int) {
	if nCap > cap(*vpl) {
		newSlice := make(VertexPoolList, len(*vpl), nCap)
		copy(newSlice, *vpl)
		*vpl = newSlice
	}
}

func (vpl *VertexPoolList) Add(pt Point64, flags VertexFlags, prev *Vertex) *Vertex {
	v := &Vertex{pt: pt, flags: flags, prev: prev}
	*vpl = append(*vpl, v)
	return v
}

type LocalMinima struct {
	Vertex   *Vertex
	PolyType PathType
	IsOpen   bool
}

func NewLocalMinima(vertex *Vertex, polytype PathType, isOpen bool) *LocalMinima {
	return &LocalMinima{
		Vertex:   vertex,
		PolyType: polytype,
		IsOpen:   isOpen,
	}
}

func (l *LocalMinima) Equals(r *LocalMinima) bool {
	return l.Vertex == r.Vertex
}

type reuseableDataContainer64 struct {
	minimaList []*LocalMinima
	vertexList VertexPoolList
}

func newReuseableDataContainer64() *reuseableDataContainer64 {
	return &reuseableDataContainer64{
		minimaList: make([]*LocalMinima, 0),
		vertexList: make(VertexPoolList, 0),
	}
}

func (r *reuseableDataContainer64) AddPaths(paths Paths64, pt PathType, isOpen bool) {
	addPathsToVertexList(paths, pt, isOpen, &r.minimaList, &r.vertexList)
}

func (r *reuseableDataContainer64) Clear() {
	r.minimaList = r.minimaList[:0]
	r.vertexList = r.vertexList[:0]
}

type IntersectNode struct {
	pt    Point64
	edge1 *Active
	edge2 *Active
}

func NewIntersectNode(pt Point64, edge1 *Active, edge2 *Active) *IntersectNode {
	return &IntersectNode{
		pt:    Point64{},
		edge1: edge1,
		edge2: edge2,
	}
}

func addPathsToVertexList(paths Paths64, polytype PathType, isOpen bool, minimaList *[]*LocalMinima, vertexList *VertexPoolList) {
	var totalVertCnt int
	for _, path := range paths {
		totalVertCnt += len(path)
	}
	vertexList.EnsureCapacity(len(*vertexList) + totalVertCnt)

	for _, path := range paths {
		var v0, prevV *Vertex
		for _, pt := range path {
			if v0 == nil {
				v0 = vertexList.Add(pt, None, nil)
				prevV = v0
			} else if prevV.pt != pt {
				currV := vertexList.Add(pt, None, prevV)
				prevV.next = currV
				prevV = currV
			}
		}
		if prevV.prev == nil {
			continue
		}
		if !isOpen && prevV.pt == v0.pt {
			prevV = prevV.prev
		}
		prevV.next = v0
		v0.prev = prevV
		if !isOpen && prevV.next == prevV {
			continue
		}

		var goingUp bool
		if isOpen {
			currV := v0.next
			for currV != nil && currV != v0 && currV.pt.Y == v0.pt.Y {
				currV = currV.next
			}
			if currV != nil {
				goingUp = currV.pt.Y <= v0.pt.Y
			}
			if goingUp {
				v0.flags = OpenStart
				addLocMin(v0, polytype, true, minimaList)
			} else {
				v0.flags = OpenStart | LocalMax
			}
		} else {
			prevV := v0.prev
			for prevV != nil && prevV != v0 && prevV.pt.Y == v0.pt.Y {
				prevV = prevV.prev
			}
			if prevV == nil || prevV == v0 {
				continue // only open path can be flat completely
			}
			goingUp = prevV.pt.Y > v0.pt.Y
		}

		goingUp0 := goingUp
		prevV = v0
		currV := v0.next
		for currV != nil && currV != v0 {
			if currV.pt.Y > prevV.pt.Y && goingUp {
				prevV.flags |= LocalMax
				goingUp = false
			} else if currV.pt.Y < prevV.pt.Y && !goingUp {
				goingUp = true
				addLocMin(prevV, polytype, isOpen, minimaList)
			}
			prevV = currV
			currV = currV.next
		}

		if isOpen {
			prevV.flags |= OpenEnd
			if goingUp {
				prevV.flags |= LocalMax
			} else {
				addLocMin(prevV, polytype, isOpen, minimaList)
			}
		} else if goingUp != goingUp0 {
			if goingUp0 {
				addLocMin(prevV, polytype, false, minimaList)
			} else {
				prevV.flags |= LocalMax
			}
		}
	}
}

func addLocMin(v *Vertex, polytype PathType, isOpen bool, minimaList *[]*LocalMinima) {
	if v.flags&LocalMin != None {
		return
	}
	v.flags |= LocalMin

	*minimaList = append(*minimaList, &LocalMinima{
		Vertex:   v,
		PolyType: polytype,
		IsOpen:   isOpen,
	})
}

func IsOdd(val int) bool {
	return (val & 1) != 0
}

func isHotEdge(ae *Active) bool {
	return ae.outrec != nil
}

func isOpen(ae *Active) bool {
	return ae.localMin.IsOpen
}

func isOpenEnd(ae *Active) bool {
	return ae.localMin.IsOpen && isVertexOpenEnd(ae.vertexTop)
}

func isVertexOpenEnd(v *Vertex) bool {
	return (v.flags & (OpenStart | OpenEnd)) != None
}

func disposeOutPt(op *OutPt) *OutPt {
	var result *OutPt = nil
	if op.next != op {
		result = op.next
	}

	op.prev.next = op.next
	op.next.prev = op.prev

	return result
}

func (c *clipperBase) fixSelfIntersects(outrec *OutRec) {
	op2 := outrec.pts
	if op2.prev == op2.next.next {
		return
	}

	for {
		if segsIntersect(op2.prev.pt, op2.pt, op2.next.pt, op2.next.next.pt, false) {
			if segsIntersect(op2.prev.pt, op2.pt, op2.next.next.next.pt, op2.next.next.next.next.pt, false) {
				op2 = duplicateOp(op2, false)
				op2.pt = op2.next.next.next.pt
				op2 = op2.next
			} else {
				if op2 == outrec.pts || op2.next == outrec.pts {
					outrec.pts = outrec.pts.prev
				}
				c.doSplitOp(outrec, op2)
				if outrec.pts == nil {
					return
				}
				op2 = outrec.pts
				if op2.prev == op2.next.next {
					break
				}
				continue
			}
		}
		op2 = op2.next
		if op2 == outrec.pts {
			break
		}
	}
}

func isValidClosedPath(op *OutPt) bool {
	return op != nil && op.next != op &&
		(op.next != op.prev || !isVerySmallTriangle(op))
}

func pointsEqual(p1, p2 Point64) bool {
	return p1.Equals(p2)
}

func getPrevHotEdge(ae *Active) *Active {
	prev := ae.prevInAEL
	for prev != nil && (isOpen(prev) || !isHotEdge(prev)) {
		prev = prev.prevInAEL
	}

	return prev
}

func isFront(ae *Active) bool {
	return ae == ae.outrec.frontEdge
}

func getDx(pt1, pt2 Point64) float64 {
	dy := pt2.Y - pt1.Y
	if dy != 0 {
		return float64(pt2.X-pt1.X) / float64(dy)
	}

	if pt2.X > pt1.X {
		return negInf
	}

	return posInf
}

func topX(ae *Active, currentY int64) int64 {
	if currentY == ae.top.Y || ae.top.X == ae.bot.X {
		return ae.top.X
	}
	if currentY == ae.bot.Y {
		return ae.bot.X
	}
	return ae.bot.X + int64(math.Round(ae.dx*(float64(currentY)-float64(ae.bot.Y))))
}

func isHorizontal(ae *Active) bool {
	return ae.top.Y == ae.bot.Y
}

func isHeadingRightHorz(ae *Active) bool {
	return math.IsInf(ae.dx, -1)
}

func isHeadingLeftHorz(ae *Active) bool {
	return math.IsInf(ae.dx, 1)
}

func swapActives(ae1, ae2 **Active) {
	*ae1, *ae2 = *ae2, *ae1
}

func getPolyType(ae *Active) PathType {
	return ae.localMin.PolyType
}

func isSamePolyType(ae1, ae2 *Active) bool {
	return ae1.localMin.PolyType == ae2.localMin.PolyType
}

func setDx(ae *Active) {
	ae.dx = getDx(ae.bot, ae.top)
}

func nextVertex(ae *Active) *Vertex {
	if ae.windDx > 0 {
		return ae.vertexTop.next
	}

	return ae.vertexTop.prev
}

func prevPrevVertex(ae *Active) *Vertex {
	if ae.windDx > 0 {
		return ae.vertexTop.prev.prev
	}
	return ae.vertexTop.next.next
}

func isMaxima(vertex *Vertex) bool {
	return (vertex.flags & LocalMax) != None
}

func isMaximaActive(ae *Active) bool {
	return isMaxima(ae.vertexTop)
}

func getMaximaPair(ae *Active) *Active {
	ae2 := ae.nextInAEL
	for ae2 != nil {
		if ae2.vertexTop == ae.vertexTop {
			return ae2
		}
		ae2 = ae2.nextInAEL
	}
	return nil
}

func getCurrYMaximaVertexOpen(ae *Active) *Vertex {
	result := ae.vertexTop
	if ae.windDx > 0 {
		for result.next.pt.Y == result.pt.Y &&
			(result.flags&(OpenEnd|LocalMax)) == None {
			result = result.next
		}
	} else {
		for result.prev.pt.Y == result.pt.Y &&
			(result.flags&(OpenEnd|LocalMax)) == None {
			result = result.prev
		}
	}
	if !isMaxima(result) {
		return nil
	}
	return result
}

func getCurrYMaximaVertex(ae *Active) *Vertex {
	result := ae.vertexTop
	if ae.windDx > 0 {
		for result.next.pt.Y == result.pt.Y {
			result = result.next
		}
	} else {
		for result.prev.pt.Y == result.pt.Y {
			result = result.prev
		}
	}
	if !isMaxima(result) {
		return nil
	}
	return result
}

func resetHorzDirection(horz *Active, vertexMax *Vertex) (leftX, rightX int64, leftToRight bool) {
	if horz.bot.X == horz.top.X {
		leftX = horz.curX
		rightX = horz.curX
		ae := horz.nextInAEL
		for ae != nil && ae.vertexTop != vertexMax {
			ae = ae.nextInAEL
		}
		return leftX, rightX, ae != nil
	}
	if horz.curX < horz.top.X {
		leftX = horz.curX
		rightX = horz.top.X
		return leftX, rightX, true
	}
	leftX = horz.top.X
	rightX = horz.curX
	return leftX, rightX, false
}

func setSides(outrec *OutRec, startEdge, endEdge *Active) {
	outrec.frontEdge = startEdge
	outrec.backEdge = endEdge
}

func swapOutRecs(ae1, ae2 *Active) {
	or1 := ae1.outrec
	or2 := ae2.outrec

	if or1 != nil && or2 != nil && or1 == or2 {
		or1.frontEdge, or1.backEdge = or1.backEdge, or1.frontEdge
		return
	}

	if or1 != nil {
		if or1.frontEdge == ae1 {
			or1.frontEdge = ae2
		} else {
			or1.backEdge = ae2
		}
	}

	if or2 != nil {
		if or2.frontEdge == ae2 {
			or2.frontEdge = ae1
		} else {
			or2.backEdge = ae1
		}
	}

	ae1.outrec, ae2.outrec = or2, or1
}

func setOwner(outrec, newOwner *OutRec) {
	for newOwner.owner != nil && newOwner.owner.pts == nil {
		newOwner.owner = newOwner.owner.owner
	}

	tmp := newOwner
	for tmp != nil && tmp != outrec {
		tmp = tmp.owner
	}
	if tmp != nil {
		newOwner.owner = outrec.owner
	}
	outrec.owner = newOwner
}

func areaOP(op *OutPt) float64 {
	// https://en.wikipedia.org/wiki/Shoelace_formula
	var area float64 = 0.0
	op2 := op
	for {
		area += float64(op2.prev.pt.Y+op2.pt.Y) * float64(op2.prev.pt.X-op2.pt.X)
		op2 = op2.next

		if op2 != op {
			break
		}
	}

	return area * 0.5
}

func areaTriangle(pt1, pt2, pt3 Point64) float64 {
	result := float64(pt3.Y+pt1.Y)*(float64(pt3.X)-float64(pt1.X)) +
		float64(pt1.Y+pt2.Y)*(float64(pt1.X)-float64(pt2.X)) +
		float64(pt2.Y+pt3.Y)*(float64(pt2.X)-float64(pt3.X))
	return result * 0.5
}

func getRealOutRec(outRec *OutRec) *OutRec {
	for outRec != nil && outRec.pts == nil {
		outRec = outRec.owner
	}
	return outRec
}

func isValidOwner(outRec, testOwner *OutRec) bool {
	for (testOwner != nil) && (testOwner != outRec) {
		testOwner = testOwner.owner
	}
	return testOwner == nil
}

func uncoupleOutRec(ae *Active) {
	outrec := ae.outrec
	if outrec == nil {
		return
	}
	if outrec.frontEdge != nil {
		outrec.frontEdge.outrec = nil
	}
	if outrec.backEdge != nil {
		outrec.backEdge.outrec = nil
	}
	outrec.frontEdge = nil
	outrec.backEdge = nil
}

func outrecIsAscending(hotEdge *Active) bool {
	return hotEdge == hotEdge.outrec.frontEdge
}

func SwapFrontBackSides(outrec *OutRec) {
	ae2 := outrec.frontEdge
	outrec.frontEdge = outrec.backEdge
	outrec.backEdge = ae2
	if outrec.pts != nil {
		outrec.pts = outrec.pts.next
	}
}

func edgesAdjacentInAEL(inode *IntersectNode) bool {
	return (inode.edge1.nextInAEL == inode.edge2) || (inode.edge1.prevInAEL == inode.edge2)
}

func isValidAelOrder(resident, newcomer *Active) bool {
	// different current X => larger curX goes to the right (newcomer should be right of resident)
	if newcomer.curX != resident.curX {
		return newcomer.curX > resident.curX
	}
	// get turning direction a1.top, a2.bot, a2.top
	d := CrossProduct(resident.top, newcomer.bot, newcomer.top)
	if d != 0 {
		return d < 0
	}

	// edges are collinear here

	// for starting open paths, place them according to the direction they're about to turn
	if !isMaximaActive(resident) && resident.top.Y > newcomer.top.Y {
		return CrossProduct(newcomer.bot, resident.top, nextVertex(resident).pt) <= 0
	}

	if !isMaximaActive(newcomer) && newcomer.top.Y > resident.top.Y {
		return CrossProduct(newcomer.bot, newcomer.top, nextVertex(newcomer).pt) >= 0
	}

	y := newcomer.bot.Y
	newcomerIsLeft := newcomer.isLeftBound

	if resident.bot.Y != y || resident.localMin.Vertex.pt.Y != y {
		return newcomerIsLeft
	}
	// resident must also have just been inserted
	if resident.isLeftBound != newcomerIsLeft {
		return newcomerIsLeft
	}
	if isCollinear(prevPrevVertex(resident).pt, resident.bot, resident.top) {
		return true
	}
	// compare turning direction of the alternate bound
	return (CrossProduct(prevPrevVertex(resident).pt, newcomer.bot,
		prevPrevVertex(newcomer).pt) > 0) == newcomerIsLeft
}

func insertRightEdge(ae, ae2 *Active) {
	ae2.nextInAEL = ae.nextInAEL
	if ae.nextInAEL != nil {
		ae.nextInAEL.prevInAEL = ae2
	}
	ae2.prevInAEL = ae
	ae.nextInAEL = ae2
}

func isJoined(e *Active) bool {
	return e.joinWith != JoinNone
}

func swapFrontBackSides(outrec *OutRec) {
	// while this proc is needed for open paths
	// it's almost never needed for closed paths
	ae2 := outrec.frontEdge
	outrec.frontEdge = outrec.backEdge
	outrec.backEdge = ae2
	if outrec.pts != nil {
		outrec.pts = outrec.pts.next
	}
}

func addOutPt(ae *Active, pt Point64) *OutPt {
	outrec := ae.outrec
	toFront := isFront(ae)
	opFront := outrec.pts
	opBack := opFront.next

	switch {
	case toFront && pt == opFront.pt:
		return opFront
	case !toFront && pt == opBack.pt:
		return opBack
	}

	newOp := newOutPt(pt, outrec)
	opBack.prev = newOp
	newOp.prev = opFront
	newOp.next = opBack
	opFront.next = newOp

	if toFront {
		outrec.pts = newOp
	}

	return newOp
}

func swapOutrecs(ae1, ae2 *Active) {
	or1 := ae1.outrec
	or2 := ae2.outrec
	if or1 != nil && or2 != nil && or1 == or2 {
		or1.frontEdge, or1.backEdge = or1.backEdge, or1.frontEdge
		return
	}

	if or1 != nil {
		if or1.frontEdge == ae1 {
			or1.frontEdge = ae2
		} else {
			or1.backEdge = ae2
		}
	}

	if or2 != nil {
		if or2.frontEdge == ae2 {
			or2.frontEdge = ae1
		} else {
			or2.backEdge = ae1
		}
	}

	ae1.outrec = or2
	ae2.outrec = or1
}

func trimHorz(horzEdge *Active, preserveCollinear bool) {
	wasTrimmed := false
	pt := nextVertex(horzEdge).pt

	for pt.Y == horzEdge.top.Y {
		if preserveCollinear && ((pt.X < horzEdge.top.X) != (horzEdge.bot.X < horzEdge.top.X)) {
			break
		}

		horzEdge.vertexTop = nextVertex(horzEdge)
		horzEdge.top = pt
		wasTrimmed = true
		if isMaximaActive(horzEdge) {
			break
		}
		pt = nextVertex(horzEdge).pt
	}
	if wasTrimmed {
		setDx(horzEdge)
	}
}

func getLastOp(hotEdge *Active) *OutPt {
	outrec := hotEdge.outrec
	if outrec == nil {
		return nil
	}
	if outrec.frontEdge == hotEdge {
		return outrec.pts
	}
	return outrec.pts.next
}

func extractFromSEL(ae *Active) *Active {
	res := ae.nextInSEL
	if res != nil {
		res.prevInSEL = ae.prevInSEL
	}
	if ae.prevInSEL != nil {
		ae.prevInSEL.nextInSEL = res
	}
	return res
}

func insertBeforeInSEL(ae1, ae2 *Active) {
	ae1.prevInSEL = ae2.prevInSEL
	if ae1.prevInSEL != nil {
		ae1.prevInSEL.nextInSEL = ae1
	}
	ae1.nextInSEL = ae2
	ae2.prevInSEL = ae1
}

func findEdgeWithMatchingLocMin(e *Active) *Active {
	result := e.nextInAEL
	for result != nil {
		if &result.localMin == &e.localMin {
			return result
		}
		if !isHorizontal(result) && (e.bot != result.bot) {
			result = nil
		} else {
			result = result.nextInAEL
		}
	}

	result = e.prevInAEL
	for result != nil {
		if &result.localMin == &e.localMin {
			return result
		}
		if !isHorizontal(result) && (e.bot != result.bot) {
			return nil
		}
		result = result.prevInAEL
	}
	return nil
}

func duplicateOp(op *OutPt, insertAfter bool) *OutPt {
	result := newOutPt(op.pt, op.outrec)
	if insertAfter {
		result.next = op.next
		//if op.next != nil {
		op.next.prev = result
		//}
		result.prev = op
		op.next = result
	} else {
		result.prev = op.prev
		//if result.prev != nil {
		result.prev.next = result
		//}
		result.next = op
		op.prev = result
	}
	return result
}

func setHorzSegHeadingForward(hs *HorzSegment, opP, opN *OutPt) bool {
	if opP.pt.X == opN.pt.X {
		return false
	}
	if opP.pt.X < opN.pt.X {
		hs.leftOp = opP
		hs.rightOp = opN
		hs.leftToRight = true
	} else {
		hs.leftOp = opN
		hs.rightOp = opP
		hs.leftToRight = false
	}
	return true
}

func fixOutRecPts(outrec *OutRec) {
	if outrec.pts == nil {
		return
	}
	op := outrec.pts
	start := op
	for {
		op.outrec = outrec
		op = op.next
		if op == start {
			break
		}
	}
}

func ptsReallyClose(pt1, pt2 Point64) bool {
	return (absInt(pt1.X-pt2.X) < 2) && (absInt(pt1.Y-pt2.Y) < 2)
}

func isVerySmallTriangle(op *OutPt) bool {
	return op.next.next == op.prev &&
		(ptsReallyClose(op.prev.pt, op.next.pt) ||
			ptsReallyClose(op.pt, op.next.pt) ||
			ptsReallyClose(op.pt, op.prev.pt))
}
