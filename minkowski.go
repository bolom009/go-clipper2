package go_clipper2

import "math"

func MinkowskiSum64(pattern, path Path64, isClosed bool) Paths64 {
	return UnionPaths64(minkowskiInternal(pattern, path, true, isClosed), NonZero)
}

func MinkowskiSumD(pattern, path PathD, isClosed bool, decimalPlaces ...int) PathsD {
	dVal := 2
	if len(decimalPlaces) > 0 {
		dVal = decimalPlaces[0]
	}

	scale := math.Pow(10, float64(dVal))
	sPattern := ScalePathDToPath64(pattern, scale)
	sPath := ScalePathDToPath64(path, scale)

	tmp := UnionPaths64(minkowskiInternal(sPattern, sPath, true, isClosed), NonZero)
	return ScalePaths64ToPathsD(tmp, 1/scale)
}

func MinkowskiDiff64(pattern, path Path64, isClosed bool) Paths64 {
	return UnionPaths64(minkowskiInternal(pattern, path, false, isClosed), NonZero)
}

func MinkowskiDiffD(pattern, path PathD, isClosed bool, decimalPlaces ...int) PathsD {
	dVal := 2
	if len(decimalPlaces) > 0 {
		dVal = decimalPlaces[0]
	}

	scale := math.Pow(10, float64(dVal))
	sPattern := ScalePathDToPath64(pattern, scale)
	sPath := ScalePathDToPath64(path, scale)

	tmp := UnionPaths64(minkowskiInternal(sPattern, sPath, false, isClosed), NonZero)
	return ScalePaths64ToPathsD(tmp, 1/scale)
}

func minkowskiInternal(pattern Path64, path Path64, isSum bool, isClosed bool) Paths64 {
	delta := 1
	if isClosed {
		delta = 0
	}
	patLen := len(pattern)
	pathLen := len(path)

	tmp := make(Paths64, 0, pathLen)
	for _, pathPt := range path {
		path2 := make(Path64, 0, patLen)
		if isSum {
			for _, basePt := range pattern {
				path2 = append(path2, Point64{X: pathPt.X + basePt.X, Y: pathPt.Y + basePt.Y})
			}
		} else {
			for _, basePt := range pattern {
				path2 = append(path2, Point64{X: pathPt.X - basePt.X, Y: pathPt.Y - basePt.Y})
			}
		}
		tmp = append(tmp, path2)
	}

	result := make(Paths64, 0, (pathLen-delta)*patLen)
	g := 0
	if isClosed {
		g = pathLen - 1
	}

	h := patLen - 1
	for i := delta; i < pathLen; i++ {
		for j := 0; j < patLen; j++ {
			quad := Path64{
				tmp[g][h], tmp[i][h], tmp[i][j], tmp[g][j],
			}

			if !IsPositive64(quad) {
				ReversePath64(quad)
				result = append(result, quad)
			} else {
				result = append(result, quad)
			}
			h = j
		}
		g = i
	}

	return result
}
