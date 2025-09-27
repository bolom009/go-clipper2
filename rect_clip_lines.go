package go_clipper2

import "math"

type RectClipLines64 struct {
	*RectClip64
}

func NewRectClipLines64(rect Rect64) *RectClipLines64 {
	return &RectClipLines64{
		NewRectClip64(rect),
	}
}

func (r *RectClip64) Execute(paths Paths64) Paths64 {
	result := Paths64{}

	if r.rect.IsEmpty() {
		return result
	}

	for _, path := range paths {
		if len(path) < 3 {
			continue
		}
		r.pathBounds = getBounds(path)

		if !r.rect.Intersects(r.pathBounds) {
			continue
		}
		if r.rect.Contains(r.pathBounds) {
			result = append(result, path)
			continue
		}

		r.executeInternal(path)
		r.checkEdges()

		for i := 0; i < 4; i++ {
			r.tidyEdgePair(i, r.edges[i*2], r.edges[i*2+1])
		}

		for _, op := range r.results {
			if op == nil {
				continue
			}
			tmp := getPath(op)
			if len(tmp) > 0 {
				result = append(result, tmp)
			}
		}

		r.results = r.results[:0]
		for i := 0; i < 8; i++ {
			r.edges[i] = r.edges[i][:0]
		}
	}

	return result
}

func RectClipLinesPaths64(rect Rect64, paths Paths64) Paths64 {
	if rect.IsEmpty() || len(paths) == 0 {
		return Paths64{}
	}

	rc := NewRectClipLines64(rect)
	return rc.Execute(paths)
}

func RectClipLinesPath64(rect Rect64, path Path64) Paths64 {
	if rect.IsEmpty() || len(path) == 0 {
		return Paths64{}
	}

	tmp := Paths64{path}
	return RectClipLinesPaths64(rect, tmp)
}

func RectClipLinesPathsD(rect RectD, paths PathsD, precisionV ...int) PathsD {
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

	rc := NewRectClipLines64(r)
	result := rc.Execute(tmpPaths)
	return ScalePaths64ToPathsD(result, scale)
}

func RectClipLinesPathD(rect RectD, path PathD) PathsD {
	if rect.IsEmpty() || len(path) == 0 {
		return PathsD{}
	}

	tmp := PathsD{path}
	return RectClipLinesPathsD(rect, tmp)
}
