package go_clipper2

import (
	"fmt"
	"math"
	"slices"
	"sort"
)

type PolyPathBase struct {
	parent   *PolyPathBase
	children []*PolyPathBase
}

func NewPolyPathBase(parent *PolyPathBase) *PolyPathBase {
	return &PolyPathBase{
		parent:   parent,
		children: make([]*PolyPathBase, 0),
	}
}

func (p *PolyPathBase) AddChild(pth Path64) *PolyPathBase {
	child := NewPolyPathBase(p)
	// Возможно, создаёте свой класс, реализующий PolyPathBase
	// или при необходимости возвращать или добавлять конкретный тип
	p.children = append(p.children, child)
	return child
}

func (p *PolyPathBase) Clear() {
	p.children = p.children[:0]
}

func (p *PolyPathBase) GetChildren() []*PolyPathBase {
	return p.children
}

func (p *PolyPathBase) Level() int {
	level := 0
	pp := p
	for pp.parent != nil {
		level++
		pp = pp.parent
	}
	return level
}

func (p *PolyPathBase) IsHole() bool {
	return p.Level()%2 == 0 && p.Level() != 0
}

type ClipperBase struct {
	PreserveCollinear bool
	ReverseSolution   bool

	succeeded          bool
	hasOpenPaths       bool
	isSortedMinimaList bool
	minimaList         []*LocalMinima
	intersectList      []*IntersectNode
	outrecList         []*OutRec
	horzSegList        []*HorzSegment
	horzJoinList       []*HorzJoin
	scanlineList       []int64
	vertexList         VertexPoolList
	fillRule           FillRule
	clipType           ClipType
	currentBotY        int64
	currentLocMin      int
	actives            *Active
	sel                *Active
}

func NewClipperBase() *ClipperBase {
	return &ClipperBase{
		minimaList:        make([]*LocalMinima, 0),
		intersectList:     make([]*IntersectNode, 0),
		vertexList:        make(VertexPoolList, 0),
		outrecList:        make([]*OutRec, 0),
		scanlineList:      make([]int64, 0),
		horzSegList:       make([]*HorzSegment, 0),
		horzJoinList:      make([]*HorzJoin, 0),
		PreserveCollinear: true,
	}
}

func (c *ClipperBase) addSubject(path Paths64) {
	c.addPaths(path, Subject, false)
}

func (c *ClipperBase) addPaths(paths Paths64, polytype PathType, isOpen bool) {
	c.baseAddPaths(paths, polytype, isOpen)
}

func (c *ClipperBase) baseAddPaths(paths Paths64, polytype PathType, isOpen bool) {
	if isOpen {
		c.hasOpenPaths = true
	}

	c.isSortedMinimaList = false
	addPathsToVertexList(paths, polytype, isOpen, &c.minimaList, &c.vertexList)
}

func (c *ClipperBase) execute(clipType ClipType, fillRule FillRule,
	solutionClosed, solutionOpen *Paths64) bool {

	//solutionClosed = make(Paths64, 0)
	//solutionOpen = make(Paths64, 0)

	c.executeInternal(clipType, fillRule)
	c.buildPaths(solutionClosed, solutionOpen)

	c.clearSolutionOnly()
	return c.succeeded
}

func (c *ClipperBase) buildPaths(solutionClosed, solutionOpen *Paths64) bool {
	// Очищаем решения
	*solutionClosed = (*solutionClosed)[:0]
	*solutionOpen = (*solutionOpen)[:0]

	if cap(c.outrecList) < len(*solutionClosed) {
		*solutionOpen = make(Paths64, 0, len(c.outrecList))
	}

	if cap(c.outrecList) < len(*solutionClosed) {
		*solutionClosed = make(Paths64, 0, len(c.outrecList))
	}

	fmt.Println("buildPaths", len(c.outrecList))

	i := 0
	for i < len(c.outrecList) {
		outrec := c.outrecList[i]
		i++
		if outrec.pts == nil {
			continue
		}
		var path []Point64
		if outrec.isOpen {
			if c.buildPath(outrec.pts, c.ReverseSolution, true, &path) {
				*solutionOpen = append(*solutionOpen, path)
			}
		} else {
			c.cleanCollinear(outrec)
			// путя всегда ориентированы по положительной ориентации
			if c.buildPath(outrec.pts, c.ReverseSolution, false, &path) {
				*solutionClosed = append(*solutionClosed, path)
			}
		}
	}
	return true
}

func (c *ClipperBase) cleanCollinear(outrec *OutRec) {
	outrec = getRealOutRec(outrec)
	if outrec == nil || outrec.isOpen {
		return
	}

	if !isValidClosedPath(outrec.pts) {
		outrec.pts = nil
		return
	}

	startOp := outrec.pts
	op2 := startOp

	for {
		prevPt := op2.prev.pt
		currPt := op2.pt
		nextPt := op2.next.pt
		if isCollinear(prevPt, currPt, nextPt) &&
			((currPt == prevPt) || (currPt == nextPt) || !c.PreserveCollinear ||
				dotProduct64(prevPt, currPt, nextPt) < 0) {
			if op2 == outrec.pts {
				outrec.pts = op2.prev
			}
			op2 = disposeOutPt(op2)
			if !isValidClosedPath(op2) {
				outrec.pts = nil
				return
			}
			startOp = op2
			op2 = startOp
			continue
		}
		op2 = op2.next
		if op2 == startOp {
			break
		}
	}

	c.fixSelfIntersects(outrec)
}

func (c *ClipperBase) buildPath(op *OutPt, reverse, isOpen bool, path *[]Point64) bool {
	if op == nil || op.next == op || (!isOpen && op.next == op.prev) {
		return false
	}

	*path = (*path)[:0] // очищаем

	var lastPt Point64
	var op2 *OutPt

	if reverse {
		lastPt = op.pt
		op2 = op.prev
	} else {
		op = op.next
		lastPt = op.pt
		op2 = op.next
	}
	*path = append(*path, lastPt)

	for op2 != op {
		if op2.pt != lastPt {
			lastPt = op2.pt
			*path = append(*path, lastPt)
		}
		if reverse {
			op2 = op2.prev
		} else {
			op2 = op2.next
		}
	}

	if len(*path) != 3 || isOpen || !isVerySmallTriangle(op2) {
		return true
	}
	return false
}

func (c *ClipperBase) executeInternal(ct ClipType, fillRule FillRule) {
	if ct == NoClip {
		return
	}

	c.fillRule = fillRule
	c.clipType = ct

	c.reset()

	y, ok := c.popScanline()
	if !ok {
		return
	}

	for c.succeeded {
		fmt.Println("+++ LOOP", y)

		c.insertLocalMinimaIntoAEL(y)

		for {
			ae, ok := c.popHorz()
			if !ok {
				break
			}

			c.doHorizontal(ae)
		}

		if len(c.horzSegList) > 0 {
			fmt.Println("_horzSegList.Count", len(c.horzSegList))
			c.convertHorzSegsToJoins()
			c.horzSegList = nil
		}

		c.currentBotY = y

		y, ok = c.popScanline()
		if !ok {
			fmt.Println("===> PopScanline END")
			break
		}

		fmt.Println("===> Y =", y, c.sel)

		c.doIntersections(y)
		c.doTopOfScanbeam(y)

		for {
			ae, ok := c.popHorz()
			if !ok {
				break
			}

			fmt.Println("HELLO??")

			c.doHorizontal(ae)
		}
	}

	if c.succeeded {
		c.processHorzJoins()
	}
}

func (c *ClipperBase) doTopOfScanbeam(y int64) {
	fmt.Println("--------------DoTopOfScanbeam", y)
	c.sel = nil // сбросим флаг горизонтальных линий
	ae := c.actives
	for ae != nil {
		fmt.Println("+= doTopOfScanbeam DO ITERATION")
		if ae.top.Y == y {
			ae.curX = ae.top.X
			if isMaximaActive(ae) {
				ae = c.doMaxima(ae)
				continue
			}

			if isHotEdge(ae) {
				addOutPt(ae, ae.top)
			}
			c.updateEdgeIntoAEL(ae)
			if isHorizontal(ae) {
				c.pushHorz(ae)
			}
		} else {
			ae.curX = topX(ae, y)
		}
		ae = ae.nextInAEL
	}
}

func (c *ClipperBase) doMaxima(ae *Active) *Active {
	prevE := ae.prevInAEL
	nextE := ae.nextInAEL

	if isOpenEnd(ae) {
		if isHotEdge(ae) {
			addOutPt(ae, ae.top)
		}
		if isHorizontal(ae) {
			return nextE
		}
		if isHotEdge(ae) {
			if isFront(ae) {
				ae.outrec.frontEdge = nil
			} else {
				ae.outrec.backEdge = nil
			}
			ae.outrec = nil
		}
		c.deleteFromAEL(ae)
		return nextE
	}

	maxPair := getMaximaPair(ae)
	if maxPair == nil {
		return nextE
	}

	if isJoined(ae) {
		c.split(ae, ae.top)
	}
	if isJoined(maxPair) {
		c.split(maxPair, maxPair.top)
	}

	// обрабатываем все Edge между ae и maxPair
	//cur := nextE
	for nextE != maxPair {
		c.intersectEdges(ae, nextE, ae.top)
		c.swapPositionsInAEL(ae, nextE)
		nextE = ae.nextInAEL
	}

	if isOpen(ae) {
		if isHotEdge(ae) {
			c.addLocalMaxPoly(ae, maxPair, ae.top)
		}
		c.deleteFromAEL(maxPair)
		c.deleteFromAEL(ae)
		if prevE != nil {
			return prevE.nextInAEL
		}
		return c.actives
	}

	if isHotEdge(ae) {
		c.addLocalMaxPoly(ae, maxPair, ae.top)
	}
	c.deleteFromAEL(ae)
	c.deleteFromAEL(maxPair)
	if prevE != nil {
		return prevE.nextInAEL
	}
	return c.actives
}

func (c *ClipperBase) doIntersections(topY int64) {
	if !c.buildIntersectList(topY) {
		fmt.Println("------------doIntersections DO ITERATION", topY)
		return
	}
	fmt.Println("------------ DoIntersectionsPROCESS", topY)
	c.processIntersectList()
	c.disposeIntersectNodes()
}

func (c *ClipperBase) processIntersectList() {
	// Сортируем по Y-началу пересечений (или по другой логике)
	sort.Slice(c.intersectList, func(i, j int) bool {
		a, b := c.intersectList[i], c.intersectList[j]
		if a.pt.Y != b.pt.Y {
			return a.pt.Y > b.pt.Y // по убыванию Y
		}
		if a.pt.X == b.pt.X {
			return false // равны
		}
		return a.pt.X < b.pt.X // по возрастанию X
	})

	for i := 0; i < len(c.intersectList); i++ {
		// Обеспечиваем, чтобы пересечения шли между соседними в списке
		if !edgesAdjacentInAEL(c.intersectList[i]) {
			j := i + 1
			for !edgesAdjacentInAEL(c.intersectList[j]) {
				j++
			}
			// swap
			c.intersectList[i], c.intersectList[j] = c.intersectList[j], c.intersectList[i]
		}
		node := c.intersectList[i]
		c.intersectEdges(node.edge1, node.edge2, node.pt)
		c.swapPositionsInAEL(node.edge1, node.edge2)

		node.edge1.curX = node.pt.X
		node.edge2.curX = node.pt.X

		c.checkJoinLeft(node.edge2, node.pt, true)
		c.checkJoinRight(node.edge1, node.pt, true)
	}
}

func (c *ClipperBase) buildIntersectList(topY int64) bool {
	if c.actives == nil || c.actives.nextInAEL == nil {
		return false
	}

	// Производим пересчет текущего X и копирование в селект список
	c.adjustCurrXAndCopyToSEL(topY)

	left := c.sel
	for left != nil && left.jump != nil {
		var prevBase *Active = nil
		for left != nil && left.jump != nil {
			currBase := left
			right := left.jump
			lEnd := right
			rEnd := right.jump

			left.jump = rEnd
			for left != lEnd && right != rEnd {
				if right.curX < left.curX {
					//, добавьте вызов AddNewIntersectNode
					tm := right.prevInSEL
					for {
						c.addNewIntersectNode(tm, right, topY)
						if tm == left {
							break
						}
						tm = tm.prevInSEL
					}

					tm = right
					right = extractFromSEL(tm)
					lEnd = right
					insertBeforeInSEL(tm, left)
					if left == currBase {
						currBase = tm
						currBase.jump = rEnd
						if prevBase == nil {
							c.sel = currBase
						} else {
							prevBase.jump = currBase
						}
					}
				} else {
					left = left.nextInSEL
				}
			}
			prevBase = currBase
			left = rEnd
		}
		left = c.sel
	}
	return len(c.intersectList) > 0
}

func (c *ClipperBase) addNewIntersectNode(ae1, ae2 *Active, topY int64) {
	var ip Point64
	intersectPt, ok := getSegmentIntersectPt(ae1.bot, ae1.top, ae2.bot, ae2.top)
	if !ok {
		ip = Point64{X: ae1.curX, Y: topY}
	} else {
		ip = intersectPt
	}

	if ip.Y > c.currentBotY || ip.Y < topY {
		absDx1 := math.Abs(ae1.dx)
		absDx2 := math.Abs(ae2.dx)

		switch {
		case absDx1 > 100:
			if absDx2 > 100 {
				if absDx1 > absDx2 {
					ip = getClosestPtOnSegment(ip, ae1.bot, ae1.top)
				} else {
					ip = getClosestPtOnSegment(ip, ae2.bot, ae2.top)
				}
			} else {
				ip = getClosestPtOnSegment(ip, ae1.bot, ae1.top)
			}
		case absDx2 > 100:
			ip = getClosestPtOnSegment(ip, ae2.bot, ae2.top)
		default:
			if ip.Y < topY {
				ip.Y = topY
			} else {
				ip.Y = c.currentBotY
			}
			if absDx1 < absDx2 {
				ip.X = topX(ae1, ip.Y)
			} else {
				ip.X = topX(ae2, ip.Y)
			}
		}
	}

	// Создаем узел пересечения
	node := &IntersectNode{
		pt:    ip,
		edge1: ae1,
		edge2: ae2,
	}
	c.intersectList = append(c.intersectList, node)
}

func (c *ClipperBase) doSplitOp(outrec *OutRec, splitOp *OutPt) {
	prevOp := splitOp.prev
	nextNextOp := splitOp.next.next

	outrec.pts = prevOp

	ip, _ := getSegmentIntersectPt(prevOp.pt, splitOp.pt, splitOp.next.pt, nextNextOp.pt)

	//if c.zCallback != nil {
	//	c.zCallback(prevOp.pt, splitOp.pt, splitOp.next.pt, nextNextOp.pt, &ip)
	//}

	area1 := areaOP(prevOp)
	absArea1 := math.Abs(area1)

	if absArea1 < 2 {
		outrec.pts = nil
		return
	}

	area2 := areaTriangle(ip, splitOp.pt, splitOp.next.pt)
	absArea2 := math.Abs(area2)

	// Проверяем совпадение точек пересечения с текущими вершинами
	if pointsEqual(ip, prevOp.pt) || pointsEqual(ip, nextNextOp.pt) {
		nextNextOp.prev = prevOp
		prevOp.next = nextNextOp
	} else {
		newOp := &OutPt{
			pt:     ip,
			outrec: outrec,
			prev:   prevOp,
			next:   nextNextOp,
		}
		nextNextOp.prev = newOp
		prevOp.next = newOp
	}

	if !(absArea2 > 1) || !(absArea2 > absArea1 && (area2 > 0) != (area1 > 0)) {
		return
	}

	newOutRec := c.newOutRec()
	newOutRec.owner = outrec.owner
	splitOp.outrec = newOutRec
	splitOp.next.outrec = newOutRec

	newOp := &OutPt{
		pt:     ip,
		outrec: newOutRec,
		prev:   splitOp.next,
		next:   splitOp,
	}
	newOutRec.pts = newOp
	splitOp.prev = newOp
	splitOp.next.next = newOp

	//if c.usingPolyTree {
	//	if c.path1InsidePath2(prevOp, newOp) {
	//		if newOutRec.Splits == nil {
	//			newOutRec.Splits = make([]int, 0)
	//		}
	//		newOutRec.Splits = append(newOutRec.Splits, outrec.Index)
	//	} else {
	//		if outrec.Splits == nil {
	//			outrec.Splits = make([]int, 0)
	//		}
	//		outrec.Splits = append(outrec.Splits, newOutRec.Index)
	//	}
	//}
}

func (c *ClipperBase) adjustCurrXAndCopyToSEL(topY int64) {
	ae := c.actives
	c.sel = ae
	for ae != nil {
		ae.prevInSEL = ae.prevInAEL
		ae.nextInSEL = ae.nextInAEL
		ae.jump = ae.nextInSEL
		ae.curX = topX(ae, topY)
		// Не обновляем ae.Curr.Y
		ae = ae.nextInAEL
	}
}

func horzSegSort(hs1, hs2 *HorzSegment) int {
	if hs1 == nil || hs2 == nil {
		return 0
	}

	if hs1.rightOp == nil {
		if hs2.rightOp == nil {
			return 0
		}

		return 1
	}

	if hs2.rightOp == nil {
		return -1
	}

	if hs1.leftOp.pt.X == hs2.leftOp.pt.X {
		return 0
	}

	return -1
}

func (c *ClipperBase) convertHorzSegsToJoins() {
	fmt.Println("+++++========> convertHorzSegsToJoins DO")

	var k int
	for _, hs := range c.horzSegList {
		if c.updateHorzSegment(hs) {
			k++
		}
	}
	if k < 2 {
		return
	}

	slices.SortFunc(c.horzSegList, horzSegSort)

	//sort.Slice(c.horzSegList[:k], func(i, j int) bool {
	//	return horzSegSort(c.horzSegList[i], c.horzSegList[j])
	//})

	for i := 0; i < k-1; i++ {
		hs1 := c.horzSegList[i]
		for j := i + 1; j < k; j++ {
			hs2 := c.horzSegList[j]
			if hs2.leftOp.pt.X >= hs1.rightOp.pt.X ||
				hs2.leftToRight == hs1.leftToRight ||
				hs2.rightOp.pt.X <= hs1.leftOp.pt.X {
				continue
			}
			currY := hs1.leftOp.pt.Y
			if hs1.leftToRight {
				for hs1.leftOp.next.pt.Y == currY &&
					hs1.leftOp.next.pt.X <= hs2.leftOp.pt.X {
					hs1.leftOp = hs1.leftOp.next
				}
				for hs2.leftOp.prev.pt.Y == currY &&
					hs2.leftOp.prev.pt.X <= hs1.leftOp.pt.X {
					hs2.leftOp = hs2.leftOp.prev
				}
				c.horzJoinList = append(c.horzJoinList, &HorzJoin{
					op1: duplicateOp(hs1.leftOp, true),
					op2: duplicateOp(hs2.leftOp, false),
				})
			} else {
				for hs1.leftOp.prev != nil && hs1.leftOp.prev.pt.Y == currY && hs1.leftOp.prev.pt.X <= hs2.leftOp.pt.X {
					hs1.leftOp = hs1.leftOp.prev
				}
				for hs2.leftOp.next != nil && hs2.leftOp.next.pt.Y == currY && hs2.leftOp.next.pt.X <= hs1.leftOp.pt.X {
					hs2.leftOp = hs2.leftOp.next
				}
				c.horzJoinList = append(c.horzJoinList, &HorzJoin{
					op1: duplicateOp(hs2.leftOp, true),
					op2: duplicateOp(hs1.leftOp, false),
				})
			}
		}
	}
}

func (c *ClipperBase) updateHorzSegment(hs *HorzSegment) bool {
	op := hs.leftOp
	outrec := getRealOutRec(op.outrec)
	outrecHasEdges := outrec.frontEdge != nil
	currY := op.pt.Y

	opP := op
	opN := op

	if outrecHasEdges {
		opA := outrec.pts
		opZ := opA.next
		for opP != opZ && opP.prev.pt.Y == currY {
			opP = opP.prev
		}
		for opN != opA && opN.next.pt.Y == currY {
			opN = opN.next
		}
	} else {
		for opP.prev != opN && opP.prev.pt.Y == currY {
			opP = opP.prev
		}
		for opN.next != opP && opN.next.pt.Y == currY {
			opN = opN.next
		}
	}

	result := setHorzSegHeadingForward(hs, opP, opN) && hs.leftOp.horz == nil
	if result {
		hs.leftOp.horz = hs
	} else {
		hs.rightOp = nil
	}

	return result
}

func (c *ClipperBase) doHorizontal(horz *Active) {
	horzIsOpen := isOpen(horz)
	Y := horz.bot.Y

	var vertexMax *Vertex
	if horzIsOpen {
		vertexMax = getCurrYMaximaVertexOpen(horz)
	} else {
		vertexMax = getCurrYMaximaVertex(horz)
	}

	leftX, rightX, isLeftToRight := resetHorzDirection(horz, vertexMax)
	if isHotEdge(horz) {
		op := addOutPt(horz, Point64{X: horz.curX, Y: Y})
		c.addToHorzSegList(op)
	}

	for {
		var ae *Active
		if isLeftToRight {
			ae = horz.nextInAEL
		} else {
			ae = horz.prevInAEL
		}

		for ae != nil {
			if ae.vertexTop == vertexMax {
				if isHotEdge(horz) && isJoined(ae) {
					c.split(ae, ae.top)
				}
				if isHotEdge(horz) {
					for horz.vertexTop != vertexMax {
						addOutPt(horz, horz.top)
						c.updateEdgeIntoAEL(horz)
					}
					if isLeftToRight {
						c.addLocalMaxPoly(horz, ae, horz.top)
					} else {
						c.addLocalMaxPoly(ae, horz, horz.top)
					}
				}
				c.deleteFromAEL(ae)
				c.deleteFromAEL(horz)
				return
			}

			var pt Point64
			if vertexMax != horz.vertexTop || isOpenEnd(horz) {
				if (isLeftToRight && ae.curX > rightX) ||
					(!isLeftToRight && ae.curX < leftX) {
					break
				}
				if ae.curX == horz.top.X && !isHorizontal(ae) {
					pt = nextVertex(horz).pt
					if isOpen(ae) && !isSamePolyType(ae, horz) && !isHotEdge(ae) {
						if (isLeftToRight && topX(ae, pt.Y) > pt.X) ||
							(!isLeftToRight && topX(ae, pt.Y) < pt.X) {
							break
						}
					} else if (isLeftToRight && topX(ae, pt.Y) >= pt.X) ||
						(!isLeftToRight && topX(ae, pt.Y) <= pt.X) {
						break
					}
				}
			}

			pt = Point64{X: ae.curX, Y: Y}

			fmt.Println("== before isLeftToRight", ae.curX, Y, isLeftToRight)

			if isLeftToRight {
				fmt.Println("==========2222=====> isLeftToRight")
				c.intersectEdges(horz, ae, pt)
				c.swapPositionsInAEL(horz, ae)
				c.checkJoinLeft(ae, pt, false)
				horz.curX = ae.curX
				ae = horz.nextInAEL
			} else {
				c.intersectEdges(ae, horz, pt)
				c.swapPositionsInAEL(ae, horz)
				c.checkJoinRight(ae, pt, false)
				horz.curX = ae.curX
				ae = horz.prevInAEL
			}

			if isHotEdge(horz) {
				c.addToHorzSegList(getLastOp(horz))
			}
		}

		if horzIsOpen && isOpenEnd(horz) {
			if isHotEdge(horz) {
				addOutPt(horz, horz.top)
				if isFront(horz) {
					//if horz.outrec != nil {
					horz.outrec.frontEdge = nil
					//}
				} else {
					//if horz.outrec != nil {
					horz.outrec.backEdge = nil
					//}
				}
				horz.outrec = nil
			}
			c.deleteFromAEL(horz)
			return
		}
		if nextVertex(horz).pt.Y != horz.top.Y {
			break
		}

		if isHotEdge(horz) {
			addOutPt(horz, horz.top)
		}

		c.updateEdgeIntoAEL(horz)

		leftX, rightX, isLeftToRight = resetHorzDirection(horz, vertexMax)
	}

	if isHotEdge(horz) {
		op := addOutPt(horz, horz.top)
		c.addToHorzSegList(op)
	}
	c.updateEdgeIntoAEL(horz)
}

func (c *ClipperBase) addToHorzSegList(op *OutPt) {
	if op.outrec.isOpen {
		return
	}

	c.horzSegList = append(c.horzSegList, NewHorzSegment(op))
}

func (c *ClipperBase) updateEdgeIntoAEL(ae *Active) {
	ae.bot = ae.top
	ae.vertexTop = nextVertex(ae)
	ae.top = ae.vertexTop.pt
	ae.curX = ae.bot.X

	setDx(ae)

	if isJoined(ae) {
		c.split(ae, ae.bot)
	}

	fmt.Println("---updateEdgeIntoAELupdateEdgeIntoAEL", ae.top.Y, ae.bot.Y, isHorizontal(ae))
	if isHorizontal(ae) {
		if !isOpen(ae) {
			trimHorz(ae, c.PreserveCollinear)
		}
		return
	}
	c.insertScanline(ae.top.Y)
	c.checkJoinLeft(ae, ae.bot, false)
	c.checkJoinRight(ae, ae.bot, true)
}

func (c *ClipperBase) popHorz() (*Active, bool) {
	if c.sel == nil {
		return nil, false
	}

	ae := c.sel
	c.sel = c.sel.nextInSEL
	return ae, true
}

func (c *ClipperBase) processHorzJoins() {
	for _, j := range c.horzJoinList {
		or1 := getRealOutRec(j.op1.outrec)
		or2 := getRealOutRec(j.op2.outrec)

		op1b := j.op1.next
		op2b := j.op2.prev

		// swap links
		j.op1.next = j.op2
		j.op2.prev = j.op1
		op1b.prev = op2b
		op2b.next = op1b

		if or1 == or2 {
			or2 = c.newOutRec()
			or2.pts = op1b
			fixOutRecPts(or2)

			if or1.pts.outrec == or2 {
				or1.pts = j.op1
				or1.pts.outrec = or1
			}

			//if c.usingPolyTree {
			//	if c.pathInsidePath2(or1.pts, or2.pts) {
			//		// swap
			//		or2.pts, or1.pts = or1.pts, or2.pts
			//		c.fixOutRecPts(or1)
			//		c.fixOutRecPts(or2)
			//		or2.owner = or1
			//	} else if c.pathInsidePath2(or2.pts, or1.pts) {
			//		or2.owner = or1
			//	} else {
			//		or2.owner = or1.owner
			//	}
			//	// add to splits list
			//	if or1.Splits == nil {
			//		or1.Splits = make([]int, 0)
			//	}
			//	or1.Splits = append(or1.Splits, or2.Idx)
			//} else {
			or2.owner = or1
			//}
		} else {
			or2.pts = nil
			//if c.usingPolyTree {
			//	c.setOwner(or2, or1)
			//	c.moveSplits(or2, or1) // реализуется отдельно
			//} else {
			or2.owner = or1
			//}
		}
	}
}

func (c *ClipperBase) reset() {
	if !c.isSortedMinimaList {
		sort.Slice(c.minimaList, func(i, j int) bool {
			return c.minimaList[i].Vertex.pt.Y > c.minimaList[j].Vertex.pt.Y
		})
		c.isSortedMinimaList = true
	}

	//if cap(c.scanlineList) < len(c.minimaList) {
	//	c.scanlineList = make([]int64, 0, len(c.minimaList))
	//} else {
	//	c.scanlineList = c.scanlineList[:0]
	//}

	for i := len(c.minimaList) - 1; i >= 0; i-- {
		c.scanlineList = append(c.scanlineList, c.minimaList[i].Vertex.pt.Y)
	}

	c.currentBotY = 0
	c.currentLocMin = 0
	c.actives = nil
	c.sel = nil
	c.succeeded = true
}

func (c *ClipperBase) clearSolutionOnly() {
	for c.actives != nil {
		c.deleteFromAEL(c.actives)
	}

	c.scanlineList = c.scanlineList[:0]
	c.disposeIntersectNodes()

	c.outrecList = c.outrecList[:0]
	c.horzSegList = c.horzSegList[:0]
	c.horzJoinList = c.horzJoinList[:0]
}

func (c *ClipperBase) disposeIntersectNodes() {
	c.intersectList = c.intersectList[:0]
}

func (c *ClipperBase) deleteFromAEL(ae *Active) {
	prev := ae.prevInAEL
	next := ae.nextInAEL
	if prev == nil && next == nil && (ae != c.actives) {
		return
	}

	if prev != nil {
		prev.nextInAEL = next
	} else {
		c.actives = next
	}
	if next != nil {
		next.prevInAEL = prev
	}
}

func (c *ClipperBase) insertScanline(y int64) {
	index := binarySearch(c.scanlineList, y)
	if index >= 0 {
		return
	}

	index = ^index
	fmt.Println("======== BEFORE InsertScanline", c.scanlineList, index, y)
	var err error
	c.scanlineList, err = insertAtIndex(c.scanlineList, index, y)
	if err != nil {
		panic(err)
	}
	c.scanlineList[index] = y
	fmt.Println("======== AFTER InsertScanline", c.scanlineList)
}

func (c *ClipperBase) popScanline() (int64, bool) {
	cnt := len(c.scanlineList) - 1
	fmt.Println("==> PopScanline BEFORE ITER", cnt, len(c.scanlineList))
	if cnt < 0 {
		return 0, false
	}

	y := c.scanlineList[cnt]
	var err error
	c.scanlineList, err = removeAtIndex(c.scanlineList, cnt)
	if err != nil {
		panic(err)
	}
	//c.scanlineList = c.scanlineList[:cnt]
	cnt--
	for cnt >= 0 && c.scanlineList[cnt] == y {
		c.scanlineList, err = removeAtIndex(c.scanlineList, cnt)
		if err != nil {
			panic(err)
		}
		//c.scanlineList = c.scanlineList[:cnt]
		cnt--
	}

	fmt.Println("==> PopScanline AFTER ITER", cnt, len(c.scanlineList))

	return y, true
}

func removeAtIndex[T any](slice []T, index int) ([]T, error) {
	if index < 0 || index >= len(slice) {
		return slice, fmt.Errorf("index out of bounds") // Or panic, depending on desired behavior
	}

	// Efficiently remove element by shifting elements
	slice = append(slice[:index], slice[index+1:]...)
	return slice, nil
}

func (c *ClipperBase) hasLocMinAtY(y int64) bool {
	return c.currentLocMin < len(c.minimaList) && c.minimaList[c.currentLocMin].Vertex.pt.Y == y
}

func (c *ClipperBase) popLocalMinima() *LocalMinima {
	cur := c.minimaList[c.currentLocMin]
	c.currentLocMin++

	return cur
}

func (c *ClipperBase) AddPath(path Path64, polytype PathType, isOpen bool) {
	tmp := Paths64{path}
	c.baseAddPaths(tmp, polytype, isOpen)
}

func (c *ClipperBase) addReuseableData(reuseableData *ReuseableDataContainer64) {
	if len(reuseableData.minimaList) == 0 {
		return
	}

	// nb: reuseableData will continue to own the vertices, so it's important
	// that the reuseableData object isn't destroyed before the Clipper object
	// that's using the data.
	c.isSortedMinimaList = false
	for _, lm := range reuseableData.minimaList {
		c.minimaList = append(c.minimaList, NewLocalMinima(lm.Vertex, lm.PolyType, lm.IsOpen))
		if lm.IsOpen {
			c.hasOpenPaths = true
		}
	}
}

func (c *ClipperBase) isContributingClosed(ae *Active) bool {
	switch c.fillRule {
	case Positive:
		if ae.windCount != 1 {
			return false
		}
	case Negative:
		if ae.windCount != -1 {
			return false
		}
	case NonZero:
		if math.Abs(float64(ae.windCount)) != 1 {
			return false
		}
	}

	switch c.clipType {
	case Intersection:
		switch c.fillRule {
		case Positive:
			return ae.windCount2 > 0
		case Negative:
			return ae.windCount2 < 0
		default:
			return ae.windCount2 != 0
		}
	case Union:
		switch c.fillRule {
		case Positive:
			return ae.windCount2 <= 0
		case Negative:
			return ae.windCount2 >= 0
		default:
			return ae.windCount2 == 0
		}
	case Difference:
		result := false
		switch c.fillRule {
		case Positive:
			result = ae.windCount2 <= 0
		case Negative:
			result = ae.windCount2 >= 0
		default:
			result = ae.windCount2 == 0
		}
		if getPolyType(ae) == Subject {
			return result
		}
		return !result
	case Xor:
		return true
	default:
		return false
	}
}

func (c *ClipperBase) isContributingOpen(ae *Active) bool {
	var isInSubj, isInClip bool
	switch c.fillRule {
	case Positive:
		isInSubj = ae.windCount > 0
		isInClip = ae.windCount2 > 0
	case Negative:
		isInSubj = ae.windCount < 0
		isInClip = ae.windCount2 < 0
	default:
		isInSubj = ae.windCount != 0
		isInClip = ae.windCount2 != 0
	}

	var result bool
	switch c.clipType {
	case Intersection:
		result = isInClip
	case Union:
		result = !isInSubj && !isInClip
	default:
		result = !isInClip
	}
	return result
}

func (c *ClipperBase) setWindCountForClosedPathEdge(ae *Active) {
	ae2 := ae.prevInAEL
	pt := getPolyType(ae)

	for ae2 != nil && (getPolyType(ae2) != pt || isOpen(ae2)) {
		ae2 = ae2.prevInAEL
	}

	if ae2 == nil {
		ae.windCount = ae.windDx
		ae2 = c.actives
	} else if c.fillRule == EvenOdd {
		ae.windCount = ae.windDx
		ae.windCount2 = ae2.windCount2
		ae2 = ae2.nextInAEL
	} else {
		if ae2.windCount*ae2.windDx < 0 {
			if math.Abs(float64(ae2.windCount)) > 1 {
				if ae2.windDx*ae.windDx < 0 {
					ae.windCount = ae2.windCount
				} else {
					ae.windCount = ae2.windCount + ae.windDx
				}
			} else {
				if isOpen(ae) {
					ae.windCount = 1
				} else {
					ae.windCount = ae.windDx
				}
			}
		} else {
			if ae2.windDx*ae.windDx < 0 {
				ae.windCount = ae2.windCount
			} else {
				ae.windCount = ae2.windCount + ae.windDx
			}
		}
		ae.windCount2 = ae2.windCount2
		ae2 = ae2.nextInAEL
	}

	if c.fillRule == EvenOdd {
		for ae2 != ae {
			if getPolyType(ae2) != pt && !isOpen(ae2) {
				if ae.windCount2 == 0 {
					ae.windCount2 = 1
				} else {
					ae.windCount2 = 0
				}
			}
			ae2 = ae2.nextInAEL
		}
	} else {
		for ae2 != ae {
			if getPolyType(ae2) != pt && !isOpen(ae2) {
				ae.windCount2 += ae2.windDx
			}
			ae2 = ae2.nextInAEL
		}
	}
}

func (c *ClipperBase) setWindCountForOpenPathEdge(ae *Active) {
	ae2 := c.actives
	if c.fillRule == EvenOdd {
		cnt1 := 0
		cnt2 := 0
		for ae2 != ae {
			if getPolyType(ae2) == Clip {
				cnt2++
			} else if !isOpen(ae2) {
				cnt1++
			}
			ae2 = ae2.nextInAEL
		}
		if IsOdd(cnt1) {
			ae.windCount = 1
		} else {
			ae.windCount = 0
		}
		if IsOdd(cnt2) {
			ae.windCount2 = 1
		} else {
			ae.windCount2 = 0
		}
	} else {
		for ae2 != ae {
			if getPolyType(ae2) == Clip {
				ae.windCount2 += ae2.windDx
			} else if !isOpen(ae2) {
				ae.windCount += ae2.windDx
			}
			ae2 = ae2.nextInAEL
		}
	}
}

func (c *ClipperBase) insertLeftEdge(ae *Active) {
	// insert into AEL (active edge list) keeping order
	if c.actives == nil {
		ae.prevInAEL = nil
		ae.nextInAEL = nil
		c.actives = ae
		return
	}

	// if new edge goes before head
	if !isValidAelOrder(c.actives, ae) {
		ae.prevInAEL = nil
		ae.nextInAEL = c.actives
		c.actives.prevInAEL = ae
		c.actives = ae
		return
	}

	ae2 := c.actives
	for ae2.nextInAEL != nil && isValidAelOrder(ae2.nextInAEL, ae) {
		ae2 = ae2.nextInAEL
	}
	// don't separate joined edges
	if ae2.joinWith == JoinRight {
		//if ae2.nextInAEL != nil {
		ae2 = ae2.nextInAEL
		//}
	}

	ae.nextInAEL = ae2.nextInAEL
	if ae2.nextInAEL != nil {
		ae2.nextInAEL.prevInAEL = ae
	}
	ae.prevInAEL = ae2
	ae2.nextInAEL = ae
}

func (c *ClipperBase) insertLocalMinimaIntoAEL(botY int64) {
	// Add any local minima (if any) at botY ...
	// NB horizontal local minima edges should contain locMin.vertex.prev
	for c.hasLocMinAtY(botY) {
		locMin := c.popLocalMinima()

		var leftBound *Active
		if (locMin.Vertex.flags & OpenStart) != None {
			leftBound = nil
		} else {
			leftBound = &Active{
				bot:       locMin.Vertex.pt,
				curX:      locMin.Vertex.pt.X,
				windDx:    -1,
				vertexTop: locMin.Vertex.prev,
				top:       locMin.Vertex.prev.pt,
				outrec:    nil,
				localMin:  locMin,
			}
			setDx(leftBound)
		}

		var rightBound *Active
		if (locMin.Vertex.flags & OpenEnd) != None {
			rightBound = nil
		} else {
			rightBound = &Active{
				bot:       locMin.Vertex.pt,
				curX:      locMin.Vertex.pt.X,
				windDx:    1,
				vertexTop: locMin.Vertex.next, // ascending
				top:       locMin.Vertex.next.pt,
				outrec:    nil,
				localMin:  locMin,
			}
			setDx(rightBound)
		}

		// Currently leftBound is descending, rightBound ascending.
		// Ensure leftBound is actually left of rightBound, swap if needed.
		if leftBound != nil && rightBound != nil {
			if isHorizontal(leftBound) {
				if isHeadingRightHorz(leftBound) {
					swapActives(&leftBound, &rightBound)
				}
			} else if isHorizontal(rightBound) {
				if isHeadingLeftHorz(rightBound) {
					swapActives(&leftBound, &rightBound)
				}
			} else if leftBound.dx < rightBound.dx {
				swapActives(&leftBound, &rightBound)
			}
			// when leftBound.windDx == 1 polygon oriented ccw in Cartesian coords
		} else if leftBound == nil {
			leftBound = rightBound
			rightBound = nil
		}

		var contributing bool
		//if leftBound != nil {
		leftBound.isLeftBound = true
		c.insertLeftEdge(leftBound)
		//}

		if isOpen(leftBound) {
			c.setWindCountForOpenPathEdge(leftBound)
			contributing = c.isContributingOpen(leftBound)
		} else {
			c.setWindCountForClosedPathEdge(leftBound)
			contributing = c.isContributingClosed(leftBound)
		}

		if rightBound != nil {
			// copy wind counts
			rightBound.windCount = leftBound.windCount
			rightBound.windCount2 = leftBound.windCount2

			insertRightEdge(leftBound, rightBound)

			if contributing {
				fmt.Println("addLocalMinPoly contributing")
				c.addLocalMinPoly(leftBound, rightBound, leftBound.bot, true)
				if !isHorizontal(leftBound) {
					c.checkJoinLeft(leftBound, leftBound.bot, false)
				}
			}

			for rightBound.nextInAEL != nil && isValidAelOrder(rightBound.nextInAEL, rightBound) {
				c.intersectEdges(rightBound, rightBound.nextInAEL, rightBound.bot)
				c.swapPositionsInAEL(rightBound, rightBound.nextInAEL)
			}

			if isHorizontal(rightBound) {
				c.pushHorz(rightBound)
			} else {
				fmt.Println("--->>insertLocalMinimaIntoAEL111", rightBound.top.Y)
				c.checkJoinRight(rightBound, rightBound.bot, false)
				c.insertScanline(rightBound.top.Y)
			}
		} else if contributing {
			c.startOpenPath(leftBound, leftBound.bot)
		}

		if isHorizontal(leftBound) {
			c.pushHorz(leftBound)
		} else {
			fmt.Println("--->>insertLocalMinimaIntoAEL222", rightBound.top.Y)
			c.insertScanline(leftBound.top.Y)
		}
	}
}

func (c *ClipperBase) swapPositionsInAEL(ae1, ae2 *Active) {
	next := ae2.nextInAEL
	if next != nil {
		next.prevInAEL = ae1
	}

	prev := ae1.prevInAEL
	if prev != nil {
		prev.nextInAEL = ae2
	}

	ae2.prevInAEL = prev
	ae2.nextInAEL = ae1

	ae1.prevInAEL = ae2
	ae1.nextInAEL = next

	if ae2.prevInAEL == nil {
		c.actives = ae2
	}
}

func (c *ClipperBase) checkJoinRight(e *Active, pt Point64, checkCurrX bool) {
	// next edge in AEL
	next := e.nextInAEL
	if next == nil ||
		!isHotEdge(e) || !isHotEdge(next) ||
		isHorizontal(e) || isHorizontal(next) ||
		isOpen(e) || isOpen(next) {
		return
	}

	// avoid trivial joins
	if (pt.Y < e.top.Y+2 || pt.Y < next.top.Y+2) &&
		(e.bot.Y > pt.Y || next.bot.Y > pt.Y) {
		return
	}

	if checkCurrX {
		if perpendicDistFromLineSqr64(pt, next.bot, next.top) > 0.25 {
			return
		}
	} else if e.curX != next.curX {
		return
	}

	if !isCollinear(e.top, pt, next.top) {
		return
	}

	if e.outrec.idx == next.outrec.idx {
		c.addLocalMaxPoly(e, next, pt)
	} else if e.outrec.idx < next.outrec.idx {
		c.joinOutrecPaths(e, next)
	} else {
		c.joinOutrecPaths(next, e)
	}

	e.joinWith = JoinRight
	next.joinWith = JoinLeft
}

func (c *ClipperBase) checkJoinLeft(e *Active, pt Point64, checkCurrX bool) {
	prev := e.prevInAEL
	if prev == nil ||
		!isHotEdge(e) || !isHotEdge(prev) ||
		isHorizontal(e) || isHorizontal(prev) ||
		isOpen(e) || isOpen(prev) {
		return
	}
	if (pt.Y < e.top.Y+2 || pt.Y < prev.top.Y+2) && // avoid trivial joins
		(e.bot.Y > pt.Y || prev.bot.Y > pt.Y) {
		return
	}

	if checkCurrX {
		if perpendicDistFromLineSqr64(pt, prev.bot, prev.top) > 0.25 {
			return
		}
	} else if e.curX != prev.curX {
		return
	}

	if !isCollinear(e.top, pt, prev.top) {
		return
	}

	if e.outrec != nil && prev.outrec != nil && e.outrec.idx == prev.outrec.idx {
		c.addLocalMaxPoly(prev, e, pt)
	} else if e.outrec != nil && prev.outrec != nil && e.outrec.idx < prev.outrec.idx {
		c.joinOutrecPaths(e, prev)
	} else {
		c.joinOutrecPaths(prev, e)
	}

	prev.joinWith = JoinRight
	e.joinWith = JoinLeft
}

func (c *ClipperBase) addLocalMinPoly(ae1, ae2 *Active, pt Point64, isNew bool) *OutPt {
	fmt.Println("addLocalMinPoly")
	outrec := c.newOutRec()
	ae1.outrec = outrec
	ae2.outrec = outrec

	if isOpen(ae1) {
		outrec.owner = nil
		outrec.isOpen = true
		if ae1.windDx > 0 {
			setSides(outrec, ae1, ae2)
		} else {
			setSides(outrec, ae2, ae1)
		}
	} else {
		outrec.isOpen = false
		prevHot := getPrevHotEdge(ae1)
		if prevHot != nil {
			//if c.usingPolyTree {
			//	c.setOwner(outrec, prevHot.outrec)
			//}
			outrec.owner = prevHot.outrec
			if outrecIsAscending(prevHot) == isNew {
				setSides(outrec, ae2, ae1)
			} else {
				setSides(outrec, ae1, ae2)
			}
		} else {
			outrec.owner = nil
			if isNew {
				setSides(outrec, ae1, ae2)
			} else {
				setSides(outrec, ae2, ae1)
			}
		}
	}

	op := NewOutPt(pt, outrec)
	outrec.pts = op
	return op
}

func (c *ClipperBase) newOutRec() *OutRec {
	result := &OutRec{idx: len(c.outrecList)}
	c.outrecList = append(c.outrecList, result)

	return result
}

func (c *ClipperBase) pushHorz(ae *Active) {
	ae.nextInSEL = c.sel
	c.sel = ae
}

func (c *ClipperBase) addLocalMaxPoly(ae1, ae2 *Active, pt Point64) *OutPt {
	// If joined edges, split at pt first
	if isJoined(ae1) {
		c.split(ae1, pt)
	}
	if isJoined(ae2) {
		c.split(ae2, pt)
	}
	// If both are front (same side), try to swap front/back for open ends,
	// otherwise failure (in C# sets _succeeded=false and returns null).
	if isFront(ae1) == isFront(ae2) {
		if isOpenEnd(ae1) {
			swapFrontBackSides(ae1.outrec)
		} else if isOpenEnd(ae2) {
			swapFrontBackSides(ae2.outrec)
		} else {
			fmt.Println("END CYCLE WHY?")
			c.succeeded = false
			return nil
		}
	}

	result := addOutPt(ae1, pt)

	// same outrec -> finalize and uncouple
	if ae1.outrec != nil && ae1.outrec == ae2.outrec {
		outrec := ae1.outrec
		outrec.pts = result

		//if c.usingPolyTree {
		//	e := c.getPrevHotEdge(ae1)
		//	if e == nil {
		//		outrec.owner = nil
		//	} else if e.outrec != nil {
		//		c.setOwner(outrec, e.outrec)
		//	}
		//	// note: owner may be fixed later by DeepCheckOwner()
		//}
		uncoupleOutRec(ae1)
	} else {
		// preserve winding orientation of outrec
		if isOpen(ae1) {
			if ae1.windDx < 0 {
				c.joinOutrecPaths(ae1, ae2)
			} else {
				c.joinOutrecPaths(ae2, ae1)
			}
		} else if ae1.outrec != nil && ae2.outrec != nil {
			if ae1.outrec.idx < ae2.outrec.idx {
				c.joinOutrecPaths(ae1, ae2)
			} else {
				c.joinOutrecPaths(ae2, ae1)
			}
		}
	}

	return result
}

func (c *ClipperBase) split(e *Active, currPt Point64) {
	if e.joinWith == JoinRight {
		e.joinWith = JoinNone
		//if e.nextInAEL != nil {
		e.nextInAEL.joinWith = JoinNone
		//}
		c.addLocalMinPoly(e, e.nextInAEL, currPt, true)
	} else {
		e.joinWith = JoinNone
		//if e.prevInAEL != nil {
		e.prevInAEL.joinWith = JoinNone
		//}
		c.addLocalMinPoly(e.prevInAEL, e, currPt, true)
	}
}

func (c *ClipperBase) joinOutrecPaths(ae1, ae2 *Active) {
	// join ae2.outrec path onto ae1.outrec path and then delete ae2.outrec pointers.
	// (Only very rarely do the joining ends share the same coords.)
	p1Start := ae1.outrec.pts
	p2Start := ae2.outrec.pts
	p1End := p1Start.next
	p2End := p2Start.next
	if isFront(ae1) {
		// link p2End <- p1Start -> p2End.prev = p1Start
		p2End.prev = p1Start
		p1Start.next = p2End
		// link p2Start -> p1End
		p2Start.next = p1End
		p1End.prev = p2Start
		ae1.outrec.pts = p2Start

		// reassign frontEdge
		ae1.outrec.frontEdge = ae2.outrec.frontEdge
		if ae1.outrec.frontEdge != nil {
			ae1.outrec.frontEdge.outrec = ae1.outrec
		}
	} else {
		// other orientation
		p1End.prev = p2Start
		p2Start.next = p1End
		p1Start.next = p2End
		p2End.prev = p1Start

		ae1.outrec.backEdge = ae2.outrec.backEdge
		if ae1.outrec.backEdge != nil {
			ae1.outrec.backEdge.outrec = ae1.outrec
		}
	}

	// after joining, the ae2.OutRec must contain no vertices ...
	ae2.outrec.frontEdge = nil
	ae2.outrec.backEdge = nil
	ae2.outrec.pts = nil
	setOwner(ae2.outrec, ae1.outrec)

	if isOpenEnd(ae1) {
		ae2.outrec.pts = ae1.outrec.pts
		ae1.outrec.pts = nil
	}

	// ae1 and ae2 are maxima and are about to be dropped from Actives list.
	ae1.outrec = nil
	ae2.outrec = nil
}

func (c *ClipperBase) intersectEdges(ae1, ae2 *Active, pt Point64) {
	fmt.Println("=====> IntersectEdges")
	var resultOp *OutPt
	if c.hasOpenPaths && (isOpen(ae1) || isOpen(ae2)) {
		if isOpen(ae1) && isOpen(ae2) {
			return
		}
		if isOpen(ae2) {
			swapActives(&ae1, &ae2)
		}
		if isJoined(ae2) {
			c.split(ae2, pt)
		}
		if c.clipType == Union {
			if !isHotEdge(ae2) {
				return
			}
		} else if ae2.localMin.PolyType == Subject {
			return
		}

		switch c.fillRule {
		case Positive:
			if ae2.windCount != 1 {
				return
			}
		case Negative:
			if ae2.windCount != -1 {
				return
			}
		default:
			if absInt(ae2.windCount) != 1 {
				return
			}
		}

		if isHotEdge(ae1) {
			resultOp = addOutPt(ae1, pt)
			if isFront(ae1) {
				//if ae1.outrec.frontEdge != nil {
				ae1.outrec.frontEdge = nil
				//}
			} else {
				//if ae1.outrec.backEdge != nil {
				ae1.outrec.backEdge = nil
				//}
			}
			ae1.outrec = nil
		} else if pt == ae1.localMin.Vertex.pt &&
			!isVertexOpenEnd(ae1.localMin.Vertex) {
			ae3 := findEdgeWithMatchingLocMin(ae1)
			if ae3 != nil && isHotEdge(ae3) {
				ae1.outrec = ae3.outrec
				if ae1.windDx > 0 {
					setSides(ae3.outrec, ae1, ae3)
				} else {
					setSides(ae3.outrec, ae3, ae1)
				}
				return
			}
			resultOp = c.startOpenPath(ae1, pt)
		} else {
			resultOp = c.startOpenPath(ae1, pt)
		}
		return
	}

	if isJoined(ae1) {
		c.split(ae1, pt)
	}
	if isJoined(ae2) {
		c.split(ae2, pt)
	}

	if ae1.localMin.PolyType == ae2.localMin.PolyType {
		if c.fillRule == EvenOdd {
			oldE1WindCount := ae1.windCount
			ae1.windCount = ae2.windCount
			ae2.windCount = oldE1WindCount
		} else {
			if ae1.windCount+ae2.windDx == 0 {
				ae1.windCount = -ae1.windCount
			} else {
				ae1.windCount += ae2.windDx
			}
			if ae2.windCount-ae1.windDx == 0 {
				ae2.windCount = -ae2.windCount
			} else {
				ae2.windCount -= ae1.windDx
			}
		}
	} else {
		if c.fillRule != EvenOdd {
			ae1.windCount2 += ae2.windDx
		} else {
			if ae1.windCount2 == 0 {
				ae1.windCount2 = 1
			} else {
				ae1.windCount2 = 0
			}
		}
		if c.fillRule != EvenOdd {
			ae2.windCount2 -= ae1.windDx
		} else {
			if ae2.windCount2 == 0 {
				ae2.windCount2 = 1
			} else {
				ae2.windCount2 = 0
			}
		}
	}

	var oldE1WindCount, oldE2WindCount int
	switch c.fillRule {
	case Positive:
		oldE1WindCount = ae1.windCount
		oldE2WindCount = ae2.windCount
	case Negative:
		oldE1WindCount = -ae1.windCount
		oldE2WindCount = -ae2.windCount
	default:
		oldE1WindCount = absInt(ae1.windCount)
		oldE2WindCount = absInt(ae2.windCount)
	}

	e1WindCountIs0or1 := oldE1WindCount == 0 || oldE1WindCount == 1
	e2WindCountIs0or1 := oldE2WindCount == 0 || oldE2WindCount == 1

	if (!isHotEdge(ae1) && !e1WindCountIs0or1) ||
		(!isHotEdge(ae2) && !e2WindCountIs0or1) {
		return
	}

	// Обработка максимума, если оба hot
	if isHotEdge(ae1) && isHotEdge(ae2) {
		fmt.Println("===============> IsHotEdge1 IsHotEdge2")
		if (oldE1WindCount != 0 && oldE1WindCount != 1) ||
			(oldE2WindCount != 0 && oldE2WindCount != 1) ||
			(ae1.localMin.PolyType != ae2.localMin.PolyType && c.clipType != Xor) {
			c.addLocalMaxPoly(ae1, ae2, pt)
		} else if isFront(ae1) || (ae1.outrec == ae2.outrec) {
			c.addLocalMaxPoly(ae1, ae2, pt)
			c.addLocalMinPoly(ae1, ae2, pt, false)
		} else {
			addOutPt(ae1, pt)
			addOutPt(ae2, pt)
			swapOutrecs(ae1, ae2)
		}

		fmt.Println("========== 66666666666")
		return
	}

	// Обработка hot edges
	if isHotEdge(ae1) {
		addOutPt(ae1, pt)
		swapOutrecs(ae1, ae2)
		return
	}
	if isHotEdge(ae2) {
		addOutPt(ae2, pt)
		swapOutrecs(ae1, ae2)
		return
	}

	// Обработка не-hot edges
	var e1Wc2, e2Wc2 int
	switch c.fillRule {
	case Positive:
		e1Wc2 = ae1.windCount2
		e2Wc2 = ae2.windCount2
	case Negative:
		e1Wc2 = -ae1.windCount2
		e2Wc2 = -ae2.windCount2
	default:
		e1Wc2 = absInt(ae1.windCount2)
		e2Wc2 = absInt(ae2.windCount2)
	}

	fmt.Println("========== 555555555555")
	if !isSamePolyType(ae1, ae2) {
		c.addLocalMinPoly(ae1, ae2, pt, false)
	} else if oldE1WindCount == 1 && oldE2WindCount == 1 {
		switch c.clipType {
		case Union:
			if e1Wc2 > 0 && e2Wc2 > 0 {
				return
			}
			c.addLocalMinPoly(ae1, ae2, pt, false)
		case Difference:
			if (getPolyType(ae1) == Clip && e1Wc2 > 0 && e2Wc2 > 0) ||
				(getPolyType(ae1) == Subject && e1Wc2 <= 0 && e2Wc2 <= 0) {
				c.addLocalMinPoly(ae1, ae2, pt, false)
			}
		case Xor:
			c.addLocalMinPoly(ae1, ae2, pt, false)
		default: // ClipTypeIntersection
			if e1Wc2 <= 0 || e2Wc2 <= 0 {
				return
			}
			c.addLocalMinPoly(ae1, ae2, pt, false)
		}
	}

	// TODO NOT SURE FOR WHAT THIS VAR
	_ = resultOp
}

func (c *ClipperBase) startOpenPath(ae *Active, pt Point64) *OutPt {
	outrec := c.newOutRec()
	outrec.isOpen = true

	if ae.windDx > 0 {
		outrec.frontEdge = ae
		outrec.backEdge = nil
	} else {
		outrec.frontEdge = nil
		outrec.backEdge = ae
	}

	ae.outrec = outrec

	op := NewOutPt(pt, outrec)
	outrec.pts = op

	return op
}
