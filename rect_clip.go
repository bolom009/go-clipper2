package go_clipper2

import "math"

type Location int8

const (
	Left Location = iota
	Top
	Right
	Bottom
	Inside
)

type OutPt2 struct {
	next     *OutPt2
	prev     *OutPt2
	pt       Point64
	ownerIdx int
	edge     []*OutPt2
}

func NewOutPt2(pt Point64) *OutPt2 {
	return &OutPt2{pt: pt}
}

type RectClip64 struct {
	rect       Rect64
	mp         Point64
	rectPath   Path64
	pathBounds Rect64
	results    []*OutPt2
	edges      [][]*OutPt2
	currIdx    int
}

func NewRectClip64(rect Rect64) *RectClip64 {
	vEdges := make([][]*OutPt2, 8)
	for i := 0; i < 8; i++ {
		vEdges[i] = make([]*OutPt2, 0)
	}

	return &RectClip64{
		currIdx:  -1,
		rect:     rect,
		mp:       rect.MidPoint(),
		rectPath: rect.AsPath(),
		results:  make([]*OutPt2, 0),
		edges:    vEdges,
	}
}

func (r *RectClip64) add(pt Point64, startingNewPath bool) *OutPt2 {
	currIdx := len(r.results)
	var result *OutPt2
	if currIdx == 0 || startingNewPath {
		result = NewOutPt2(pt)
		r.results = append(r.results, result)
		result.ownerIdx = currIdx
		result.prev = result
		result.next = result
	} else {
		currIdx--
		prevOp := r.results[currIdx]
		if prevOp.pt == pt {
			return prevOp
		}
		result = &OutPt2{
			pt:       pt,
			ownerIdx: currIdx,
			next:     prevOp.next,
		}
		if prevOp.next != nil {
			prevOp.next.prev = result
		}
		prevOp.next = result
		result.prev = prevOp
		r.results[currIdx] = result
	}
	return result
}

func (r *RectClip64) path1ContainsPath2(path1 Path64, path2 Path64) bool {
	ioCount := 0
	for _, pt := range path2 {
		pip := PointInPolygon(pt, path1)
		switch pip {
		case IsInside:
			ioCount--
		case IsOutside:
			ioCount++
		}
		if absInt(ioCount) > 1 {
			break
		}
	}
	return ioCount <= 0
}

func (r *RectClip64) addCornerLocation(prev, curr Location) {
	if headingClockwise(prev, curr) {
		r.add(r.rectPath[int(prev)], false)
	} else {
		r.add(r.rectPath[int(curr)], false)
	}
}

func (r *RectClip64) addCorner(loc *Location, isClockwise bool) {
	if isClockwise {
		r.add(r.rectPath[int(*loc)], false)
		*loc = getAdjacentLocation(*loc, true)
	} else {
		*loc = getAdjacentLocation(*loc, false)
		r.add(r.rectPath[int(*loc)], false)
	}
}

func (r *RectClip64) getNextLocation(path Path64, loc *Location, i *int, highI int) {
	switch *loc {
	case Left:
		for *i <= highI && path[*i].X <= r.rect.left {
			*i++
		}
		if *i > highI {
			break
		}
		switch {
		case path[*i].X >= r.rect.right:
			*loc = Right
		case path[*i].Y <= r.rect.top:
			*loc = Top
		case path[*i].Y >= r.rect.bottom:
			*loc = Bottom
		default:
			*loc = Inside
		}
	case Top:
		for *i <= highI && path[*i].Y <= r.rect.top {
			*i++
		}
		if *i > highI {
			break
		}
		switch {
		case path[*i].Y >= r.rect.bottom:
			*loc = Bottom
		case path[*i].X <= r.rect.left:
			*loc = Left
		case path[*i].X >= r.rect.right:
			*loc = Right
		default:
			*loc = Inside
		}

	case Right:
		for *i <= highI && path[*i].X >= r.rect.right {
			*i++
		}
		if *i > highI {
			break
		}
		switch {
		case path[*i].X <= r.rect.left:
			*loc = Left
		case path[*i].Y <= r.rect.top:
			*loc = Top
		case path[*i].Y >= r.rect.bottom:
			*loc = Bottom
		default:
			*loc = Inside
		}

	case Bottom:
		for *i <= highI && path[*i].Y >= r.rect.bottom {
			*i++
		}
		if *i > highI {
			break
		}
		switch {
		case path[*i].Y <= r.rect.top:
			*loc = Top
		case path[*i].X <= r.rect.left:
			*loc = Left
		case path[*i].X >= r.rect.right:
			*loc = Right
		default:
			*loc = Inside
		}

	case Inside:
		for *i <= highI {
			switch {
			case path[*i].X < r.rect.left:
				*loc = Left
			case path[*i].X > r.rect.right:
				*loc = Right
			case path[*i].Y > r.rect.bottom:
				*loc = Bottom
			case path[*i].Y < r.rect.top:
				*loc = Top
			default:
				r.add(path[*i], false)
				*i++
				continue
			}
			break
		}
	}
}

func (r *RectClip64) executeInternal(path Path64) {
	if len(path) < 3 || r.rect.IsEmpty() {
		return
	}

	startLocs := make([]Location, 0)
	firstCross := Inside
	crossingLoc := firstCross
	prev := firstCross

	highI := len(path) - 1
	var loc Location
	var ok bool
	if loc, ok = getLocation(r.rect, path[highI]); !ok {
		i := highI - 1
		prev, ok = getLocation(r.rect, path[i])
		for i >= 0 && !ok {
			i--
			prev, ok = getLocation(r.rect, path[i])
		}
		if i < 0 {
			for _, pt := range path {
				r.add(pt, false)
			}
			return
		}
		if prev == Inside {
			loc = Inside
		}
	}

	startingLoc := loc

	i := 0
	for i <= highI {
		prev = loc
		prevCrossLoc := Inside
		r.getNextLocation(path, &loc, &i, highI)
		if i > highI {
			break
		}

		prevPt := Point64{}
		if i == 0 {
			prevPt = path[highI]
		} else {
			prevPt = path[i-1]
		}
		crossingLoc = loc

		var ip Point64
		if ip, ok = getIntersection(r.rectPath, path[i], prevPt, &crossingLoc); !ok {
			if prevCrossLoc == Inside {
				isClockw := isClockwise(prev, loc, prevPt, path[i], r.mp)
				for {
					startLocs = append(startLocs, prev)
					prev = getAdjacentLocation(prev, isClockw)
					if prev == loc {
						break
					}
				}
				crossingLoc = prevCrossLoc
			} else if prev != Inside && prev != loc {
				isClockw := isClockwise(prev, loc, prevPt, path[i], r.mp)
				for {
					r.addCorner(&prev, isClockw)
					if prev == loc {
						break
					}
				}
			}
			i++
			continue
		}

		if loc == Inside {
			if firstCross == Inside {
				firstCross = crossingLoc
				startLocs = append(startLocs, prev)
			} else if prev != crossingLoc {
				isClockw := isClockwise(prev, crossingLoc, prevPt, path[i], r.mp)
				for {
					r.addCorner(&prev, isClockw)
					if prev == crossingLoc {
						break
					}
				}
			}
		} else if prev != Inside {
			loc = prev
			ip2, _ := getIntersection(r.rectPath, prevPt, path[i], &loc)
			if prevCrossLoc != Inside && prevCrossLoc != loc {
				r.addCornerLocation(prevCrossLoc, loc)
			}
			if firstCross == Inside {
				firstCross = loc
				startLocs = append(startLocs, prev)
			}
			loc = crossingLoc
			r.add(ip2, false)
			if ip == ip2 {
				loc, _ = getLocation(r.rect, path[i])
				r.addCornerLocation(crossingLoc, loc)
				crossingLoc = loc
				continue
			}

		} else {
			loc = crossingLoc
			if firstCross == Inside {
				firstCross = crossingLoc
			}
		}
		r.add(ip, false)
	}

	if firstCross == Inside {
		if startingLoc == Inside {
			return
		}
		if !r.pathBounds.Contains(r.rect) || !r.path1ContainsPath2(path, r.rectPath) {
			return
		}

		startLocsClockwise := startLocsAreClockwise(startLocs)
		for j := 0; j < 4; j++ {
			var k int
			if startLocsClockwise {
				k = j
			} else {
				k = 3 - j
			}

			r.add(r.rectPath[k], false)
			addToEdge(&r.edges[k*2], r.results[0])
		}
	} else if loc != Inside && (loc != firstCross || len(startLocs) > 2) {
		if len(startLocs) > 0 {
			prev = loc
			for _, loc2 := range startLocs {
				if prev == loc2 {
					continue
				}
				r.addCorner(&prev, headingClockwise(prev, loc2))
				prev = loc2
			}
			loc = prev
		}
		if loc != firstCross {
			r.addCorner(&loc, headingClockwise(loc, firstCross))
		}
	}
}

func (r *RectClip64) executeInternalPath64(path Path64) {
	r.results = r.results[:0]
	if len(path) < 2 || r.rect.IsEmpty() {
		return
	}

	prev := Inside
	i := 1
	highI := len(path) - 1

	var loc Location
	var ok bool
	if loc, ok = getLocation(r.rect, path[0]); !ok {
		prev, ok2 := getLocation(r.rect, path[i])
		for i <= highI && !ok2 {
			i++
			prev, ok2 = getLocation(r.rect, path[i])
		}
		if i > highI {
			for _, pt := range path {
				r.add(pt, false)
			}
			return
		}
		if prev == Inside {
			loc = Inside
		}
		i = 1
	}

	if loc == Inside {
		r.add(path[0], false)
	}

	for i <= highI {
		prev = loc
		r.getNextLocation(path, &loc, &i, highI)

		if i > highI {
			break
		}

		prevPt := path[i-1]
		crossingLoc := loc

		ip, ok := getIntersection(r.rectPath, path[i], prevPt, &crossingLoc)
		if !ok {
			i++
			continue
		}

		if loc == Inside {
			r.add(ip, true)
		} else if prev != Inside {
			crossingLoc = prev
			if ip2, ok2 := getIntersection(r.rectPath, prevPt, path[i], &crossingLoc); ok2 {
				r.add(ip2, true)
				r.add(ip, false)
			}
		} else {
			r.add(ip, false)
		}
	}
}

func (r *RectClip64) checkEdges() {
	for i, result := range r.results {
		var op, op2 *OutPt2
		if result == nil {
			continue
		}
		op = result
		op2 = op
		for {
			if isCollinear(op2.prev.pt, op2.pt, op2.next.pt) {
				if op2 == op {
					op2 = unlinkOpBack(op2)
					if op2 == nil {
						break
					}
					op = op2.prev
				} else {
					op2 = unlinkOpBack(op2)
					if op2 == nil {
						break
					}
				}
			} else {
				op2 = op2.next
			}
			if op2 == nil || op2 == op {
				break
			}
		}
		if op2 == nil {
			r.results[i] = nil
			continue
		}

		r.results[i] = op2

		edgeSet1 := getEdgesForPt(op.pt, r.rect)
		op2 = op

		for {
			edgeSet2 := getEdgesForPt(op2.pt, r.rect)
			if edgeSet2 != 0 && op2.edge == nil {
				combinedSet := edgeSet1 & edgeSet2
				for j := 0; j < 4; j++ {
					if (combinedSet & (1 << j)) == 0 {
						continue
					}
					if isHeadingClockwise(op2.prev.pt, op2.pt, j) {
						addToEdge(&r.edges[j*2], op2)
					} else {
						addToEdge(&r.edges[j*2+1], op2)
					}
				}
			}
			edgeSet1 = edgeSet2
			op2 = op2.next
			if op2 == nil || op2 == op {
				break
			}
		}
	}
}

func (r *RectClip64) tidyEdgePair(idx int, cw, ccw []*OutPt2) {
	if len(ccw) == 0 {
		return
	}
	isHorz := idx == 1 || idx == 3
	cwIsTowardLarger := idx == 1 || idx == 2
	i, j := 0, 0

	for i < len(cw) {
		p1 := cw[i]
		if p1 == nil || p1.next == p1.prev {
			cw[i] = nil
			i++
			j = 0
			continue
		}

		jLim := len(ccw)
		for j < jLim && (ccw[j] == nil || ccw[j].next == ccw[j].prev) {
			j++
		}
		if j == jLim {
			i++
			j = 0
			continue
		}

		var p2, p1a, p2a *OutPt2

		if cwIsTowardLarger {
			p1 = cw[i].prev
			p1a = cw[i]
			p2 = ccw[j]
			p2a = ccw[j].prev
		} else {
			p1 = cw[i]
			p1a = cw[i].prev
			p2 = ccw[j].prev
			p2a = ccw[j]
		}

		if (isHorz && !hasHorzOverlap(p1.pt, p1a.pt, p2.pt, p2a.pt)) ||
			(!isHorz && !hasHorzOverlap(p1.pt, p1a.pt, p2.pt, p2a.pt)) {
			j++
			continue
		}

		isRejoining := p2.ownerIdx != p1.ownerIdx

		if isRejoining {
			r.results[p2.ownerIdx] = nil
			setNewOwner(p2, p1.ownerIdx)
		}

		if cwIsTowardLarger {
			p1.next = p2
			p2.prev = p1
			p1a.prev = p2a
			p2a.next = p1a
		} else {
			p1.prev = p2
			p2.next = p1
			p1a.next = p2a
			p2a.prev = p1a
		}

		if !isRejoining {
			newIdx := len(r.results)
			r.results = append(r.results, p1a)
			setNewOwner(p1a, newIdx)
		}

		var op, op2 *OutPt2
		if cwIsTowardLarger {
			op = p2
			op2 = p1a
		} else {
			op = p1
			op2 = p2a
		}

		r.results[op.ownerIdx] = op
		r.results[op2.ownerIdx] = op2

		var opIsLarger, op2IsLarger bool
		if isHorz {
			opIsLarger = op.pt.X > op.prev.pt.X
			op2IsLarger = op2.pt.X > op2.prev.pt.X
		} else {
			opIsLarger = op.pt.Y > op.prev.pt.Y
			op2IsLarger = op2.pt.Y > op2.prev.pt.Y
		}

		if op.next == op.prev || (op.pt == op.prev.pt) {
			if op2IsLarger == cwIsTowardLarger {
				cw[i] = op2
				ccw[j] = nil
				i++
				j++
			} else {
				ccw[j] = nil
				i++
			}
			continue
		}
		if op2.next == op2.prev || (op2.pt == op2.prev.pt) {
			if opIsLarger == cwIsTowardLarger {
				cw[i] = op
				ccw[j] = nil
			} else {
				ccw[j] = op
				cw[i] = nil
			}
			i++
			j++
			continue
		}

		if opIsLarger == op2IsLarger {
			if opIsLarger == cwIsTowardLarger {
				cw[i] = op
				uncoupleEdge(op2)
				addToEdge(&cw, op2)
				ccw[j] = nil
			} else {
				cw[i] = nil
				ccw[j] = op2
				uncoupleEdge(op)
				addToEdge(&ccw, op)
				j = 0
			}
		} else {
			if opIsLarger == cwIsTowardLarger {
				cw[i] = op
			} else {
				ccw[j] = op
			}
			if op2IsLarger == cwIsTowardLarger {
				cw[i] = op2
			} else {
				ccw[j] = op2
			}
		}
		i++
		j++
	}
}

func isClockwise(prev, curr Location, prevPt, currPt, rectMidPoint Point64) bool {
	if areOpposites(prev, curr) {
		return CrossProduct(prevPt, rectMidPoint, currPt) < 0
	}
	return headingClockwise(prev, curr)
}

func hasVertOverlap(top1, bottom1, top2, bottom2 Point64) bool {
	return top1.Y < bottom2.Y && bottom1.Y > top2.Y
}

func addToEdge(edge *[]*OutPt2, op *OutPt2) {
	if op.edge != nil {
		return
	}
	op.edge = *edge
	*edge = append(*edge, op)
}

func uncoupleEdge(op *OutPt2) {
	if op.edge == nil {
		return
	}
	for i, op2 := range op.edge {
		if op2 != op {
			continue
		}
		op.edge[i] = nil
	}
	op.edge = nil
}

func setNewOwner(op *OutPt2, newIdx int) {
	op.ownerIdx = newIdx
	op2 := op.next
	for op2 != op {
		op2.ownerIdx = newIdx
		op2 = op2.next
	}
}

func getLocation(rec Rect64, pt Point64) (Location, bool) {
	if pt.X == rec.left && pt.Y >= rec.top && pt.Y <= rec.bottom {
		return Left, false
	}
	if pt.X == rec.right && pt.Y >= rec.top && pt.Y <= rec.bottom {
		return Right, false
	}
	if pt.Y == rec.top && pt.X >= rec.left && pt.X <= rec.right {
		return Top, false
	}
	if pt.Y == rec.bottom && pt.X >= rec.left && pt.X <= rec.right {
		return Bottom, false
	}
	var loc Location
	switch {
	case pt.X < rec.left:
		loc = Left
	case pt.X > rec.right:
		loc = Right
	case pt.Y < rec.top:
		loc = Top
	case pt.Y > rec.bottom:
		loc = Bottom
	default:
		loc = Inside
	}
	return loc, true
}

func isHorizontalPoint(pt1, pt2 Point64) bool {
	return pt1.Y == pt2.Y
}

func getSegmentIntersection(p1, p2, p3, p4 Point64) (Point64, bool) {
	var ip Point64
	res1 := CrossProduct(p1, p3, p4)
	res2 := CrossProduct(p2, p3, p4)

	if res1 == 0 {
		ip = p1
		if res2 == 0 {
			return Point64{}, false
		}
		if p1 == p3 || p1 == p4 {
			return ip, true
		}
		if isHorizontalPoint(p3, p4) {
			return ip, (p1.X > p3.X) == (p1.X < p4.X)
		}
		return ip, (p1.Y > p3.Y) == (p1.Y < p4.Y)
	}

	if res2 == 0 {
		ip = p2
		if p2 == p3 || p2 == p4 {
			return ip, true
		}
		if isHorizontalPoint(p3, p4) {
			return ip, (p2.X > p3.X) == (p2.X < p4.X)
		}
		return ip, (p2.Y > p3.Y) == (p2.Y < p4.Y)
	}

	if (res1 > 0) == (res2 > 0) {
		return Point64{}, false
	}

	res3 := CrossProduct(p3, p1, p2)
	res4 := CrossProduct(p4, p1, p2)

	if res3 == 0 {
		ip = p3
		if p3 == p1 || p3 == p2 {
			return ip, true
		}
		if isHorizontalPoint(p1, p2) {
			return ip, (p3.X > p1.X) == (p3.X < p2.X)
		}
		return ip, (p3.Y > p1.Y) == (p3.Y < p2.Y)
	}

	if res4 == 0 {
		ip = p4
		if p4 == p1 || p4 == p2 {
			return ip, true
		}
		if isHorizontalPoint(p1, p2) {
			return ip, (p4.X > p1.X) == (p4.X < p2.X)
		}
		return ip, (p4.Y > p1.Y) == (p4.Y < p2.Y)
	}

	if (res3 > 0) == (res4 > 0) {
		return Point64{}, false
	}

	return getSegmentIntersectPt(p1, p2, p3, p4)
}

func getIntersection(rectPath Path64, p, p2 Point64, loc *Location) (Point64, bool) {
	switch *loc {
	case Left:
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[3]); ok {
			return ip2, true
		}
		if p.Y < rectPath[0].Y {
			if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[1]); ok {
				*loc = Top
				return ip2, true
			}
		}
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[2], rectPath[3]); !ok {
			return Point64{}, false
		} else {
			*loc = Bottom
			return ip2, true
		}
	case Right:
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[1], rectPath[2]); ok {
			return ip2, true
		}
		if p.Y < rectPath[0].Y {
			if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[1]); ok {
				*loc = Top
				return ip2, true
			}
		}
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[2], rectPath[3]); !ok {
			return Point64{}, false
		} else {
			*loc = Bottom
			return ip2, true
		}

	case Top:
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[1]); ok {
			return ip2, true
		}
		if p.X < rectPath[0].X {
			if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[3]); ok {
				*loc = Left
				return ip2, true
			}
		}
		if p.X <= rectPath[1].X {
			return Point64{}, false
		}
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[1], rectPath[2]); !ok {
			return Point64{}, false
		} else {
			*loc = Right
			return ip2, true
		}

	case Bottom:
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[2], rectPath[3]); ok {
			return ip2, true
		}
		if p.X < rectPath[3].X {
			if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[3]); ok {
				*loc = Left
				return ip2, true
			}
		}
		if p.X <= rectPath[2].X {
			return Point64{}, false
		}
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[1], rectPath[2]); !ok {
			return Point64{}, false
		} else {
			*loc = Right
			return ip2, true
		}

	default:
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[3]); ok {
			*loc = Left
			return ip2, true
		}
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[0], rectPath[1]); ok {
			*loc = Top
			return ip2, true
		}
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[1], rectPath[2]); ok {
			*loc = Right
			return ip2, true
		}
		if ip2, ok := getSegmentIntersection(p, p2, rectPath[2], rectPath[3]); !ok {
			return Point64{}, false
		} else {
			*loc = Bottom
			return ip2, true
		}
	}
}

func headingClockwise(prev, curr Location) bool {
	return (int(prev)+1)%4 == int(curr)
}

func getAdjacentLocation(loc Location, isClockwise bool) Location {
	delta := 3
	if isClockwise {
		delta = 1
	}

	return Location((int(loc) + delta) % 4)
}

func unlinkOp(op *OutPt2) *OutPt2 {
	if op.next == op {
		return nil
	}
	op.prev.next = op.next
	op.next.prev = op.prev
	return op.next
}

func unlinkOpBack(op *OutPt2) *OutPt2 {
	if op.next == op {
		return nil
	}
	op.prev.next = op.next
	op.next.prev = op.prev
	return op.prev
}

func getEdgesForPt(pt Point64, rec Rect64) uint {
	var result uint = 0
	if pt.X == rec.left {
		result = 1
	} else if pt.X == rec.right {
		result = 4
	}
	if pt.Y == rec.top {
		result += 2
	} else if pt.Y == rec.bottom {
		result += 8
	}
	return result
}

func isHeadingClockwise(pt1, pt2 Point64, edgeIdx int) bool {
	switch edgeIdx {
	case 0:
		return pt2.Y < pt1.Y
	case 1:
		return pt2.X > pt1.X
	case 2:
		return pt2.Y > pt1.Y
	default:
		return pt2.X < pt1.X
	}
}

func hasHorzOverlap(left1, right1, left2, right2 Point64) bool {
	return left1.X < right2.X && right1.X > left2.X
}

func areOpposites(prev, curr Location) bool {
	return int(math.Abs(float64(int(prev)-int(curr)))) == 2
}

func startLocsAreClockwise(startLocs []Location) bool {
	result := 0
	for i := 1; i < len(startLocs); i++ {
		d := int(startLocs[i]) - int(startLocs[i-1])
		switch d {
		case -1:
			result -= 1
		case 1:
			result += 1
		case -3:
			result += 1
		case 3:
			result -= 1
		}
	}
	return result > 0
}

func getPath(op *OutPt2) Path64 {
	var result Path64
	if op == nil || op == op.next {
		return result
	}
	op = op.next
	result = append(result, op.pt)
	op2 := op.next
	for op2 != op {
		result = append(result, op2.pt)
		op2 = op2.next
	}
	return result
}

func RectClipPaths64(rect Rect64, paths Paths64) Paths64 {
	if rect.IsEmpty() || len(paths) == 0 {
		return Paths64{}
	}

	rc := NewRectClip64(rect)
	return rc.Execute(paths)
}

func RectClipPath64(rect Rect64, path Path64) Paths64 {
	if rect.IsEmpty() || len(path) == 0 {
		return Paths64{}
	}

	tmp := Paths64{path}
	return RectClipPaths64(rect, tmp)
}

func RectClipPathsD(rect RectD, paths PathsD, precisionV ...int) PathsD {
	precision := 2
	if len(precisionV) > 0 {
		precision = precisionV[0]
	}

	checkPrecision(precision)
	if rect.IsEmpty() || len(paths) == 0 {
		return PathsD{}
	}

	scale := math.Pow(10, float64(precision))
	r := ScaleRectD(rect, scale)
	tmpPaths := ScalePathsDToPaths64(paths, scale)

	rc := NewRectClip64(r)
	result := rc.Execute(tmpPaths)
	return ScalePaths64ToPathsD(result, scale)
}

func RectClipPathD(rect RectD, path PathD) PathsD {
	if rect.IsEmpty() || len(path) == 0 {
		return PathsD{}
	}

	tmp := PathsD{path}
	return RectClipPathsD(rect, tmp)
}
