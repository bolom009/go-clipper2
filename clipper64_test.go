package go_clipper2

import (
	"fmt"
	"testing"
)

func TestClipper64(t *testing.T) {
	subject := Paths64{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip := Paths64{
		{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
	}

	union := UnionWithClipPaths64(subject, clip, NonZero)
	fmt.Println("=======> RESULT", union)
}
