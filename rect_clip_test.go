package go_clipper2_test

import (
	"testing"

	goclipper2 "github.com/bolom009/go-clipper2"
	"github.com/stretchr/testify/assert"
)

func TestRectClipPaths64(t *testing.T) {
	var (
		rect    = goclipper2.NewRect64(1222, 1323, 3247, 3348)
		subject = goclipper2.MakePath64(375, 1680, 1915, 4716, 5943, 586, 3987, 152)
		expect  = goclipper2.Paths64{{{1222, 1323}, {1222, 3348}, {3247, 3348}, {3247, 1323}}}
	)

	solution := goclipper2.RectClipPaths64(rect, goclipper2.Paths64{subject})

	assert.Equal(t, len(solution), 1)
	assert.EqualValues(t, expect, solution)
}
