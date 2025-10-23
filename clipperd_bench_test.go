package go_clipper2

import (
	"math"
	"testing"
)

func BenchmarkClipperD(b *testing.B) {
	circle := circlePath(5.0, 5.0, 3.0, 32)
	circle2 := circlePath(7.0, 7.0, 1.0, 32)
	rectangle := MakePathD(0.0, 0.0, 5.0, 0.0, 5.0, 6.0, 0.0, 6.0)

	b.Run("difference", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = DifferenceWithClipPathsD(PathsD{circle}, PathsD{circle2, rectangle}, EvenOdd)
		}
	})
}

func circlePath(offsetX, offsetY, radius float64, segments int) PathD {
	pts := make(PathD, 0, segments)
	for i := 0; i < segments; i++ {
		angle := float64(i) / float64(segments) * 2.0 * math.Pi
		x := math.Sin(angle)*radius + offsetX
		y := math.Cos(angle)*radius + offsetY
		pts = append(pts, PointD{X: x, Y: y})
	}
	return pts
}
