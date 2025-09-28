package go_clipper2_test

import (
	"reflect"
	"testing"

	goclipper2 "github.com/bolom009/go-clipper2"
	"github.com/stretchr/testify/assert"
)

func TestSimplifyPath64(t *testing.T) {
	var (
		path         = goclipper2.MakePath64(0, 0, 1, 1, 0, 20, 0, 21, 1, 40, 0, 41, 0, 60, 0, 61, 0, 80, 1, 81, 0, 100)
		epsilon      = 2.0
		isClosedPath = false
		expect       = goclipper2.Path64{
			{0, 0}, {0, 100},
		}
	)

	got := goclipper2.SimplifyPath64(path, epsilon, isClosedPath)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("SimplifyPath64() = %v, want %v", got, expect)
	}
}

func TestSimplifyPathD(t *testing.T) {
	var (
		path         = goclipper2.MakePathD(0, 0, 1, 1, 0, 20, 0, 21, 1, 40, 0, 41, 0, 60, 0, 61, 0, 80, 1, 81, 0, 100)
		epsilon      = 2.0
		isClosedPath = false
		expect       = goclipper2.PathD{
			{0, 0}, {0, 100},
		}
	)

	got := goclipper2.SimplifyPathD(path, epsilon, isClosedPath)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("SimplifyPathD() = %v, want %v", got, expect)
	}
}

func TestTrimCollinear64(t *testing.T) {
	var (
		path         = goclipper2.MakePath64(10, 10, 10, 10, 50, 10, 100, 10, 100, 100, 10, 100, 10, 10, 20, 10)
		isClosedPath = false
		expect       = goclipper2.Path64{
			{100, 10}, {100, 100}, {10, 100}, {10, 10},
		}
	)

	got := goclipper2.TrimCollinear64(path, isClosedPath)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("SimplifyPathD() = %v, want %v", got, expect)
	}
}

func TestTrimCollinear64Area(t *testing.T) {
	path := goclipper2.MakePath64(2, 3, 3, 4, 4, 4, 4, 5, 7, 5, 8, 4, 8, 3, 9, 3, 8, 3, 7, 3, 6, 3, 5, 3, 4, 3, 3, 3, 2, 3)

	output4a := goclipper2.TrimCollinear64(path, false)
	output4b := goclipper2.TrimCollinear64(output4a, false)
	area4a := int(goclipper2.Area64(output4a))
	area4b := int(goclipper2.Area64(output4b))

	assert.Equal(t, len(output4a), 7)
	assert.Equal(t, area4a, -9)
	assert.Equal(t, len(output4a), len(output4b))
	assert.Equal(t, area4a, area4b)
}

func TestTrimCollinearD(t *testing.T) {
	var (
		path         = goclipper2.MakePathD(10, 10, 10, 10, 50, 10, 100, 10, 100, 100, 10, 100, 10, 10, 20, 10)
		isClosedPath = false
		expect       = goclipper2.PathD{
			{100, 10}, {100, 100}, {10, 100}, {10, 10},
		}
	)

	got := goclipper2.TrimCollinearD(path, 2.0, isClosedPath)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("SimplifyPathD() = %v, want %v", got, expect)
	}
}
