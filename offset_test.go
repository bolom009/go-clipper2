package go_clipper2_test

import (
	"fmt"
	"reflect"
	"testing"

	goclipper2 "github.com/bolom009/go-clipper2"
)

func TestInflatePaths64(t *testing.T) {
	subject := goclipper2.Paths64{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}

	tests := []struct {
		name        string
		delta       float64
		joinType    goclipper2.JoinType
		endType     goclipper2.EndType
		overSubject goclipper2.Paths64
		overClip    goclipper2.Paths64
		expect      goclipper2.Paths64
	}{
		{
			name:     "offset polygon with squire join type",
			delta:    40,
			joinType: goclipper2.Square,
			endType:  goclipper2.Polygon,
			expect: goclipper2.Paths64{
				{{140, -17}, {140, 117}, {117, 140}, {-17, 140}, {-40, 117}, {-40, -17}, {-17, -40}, {117, -40}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			results := goclipper2.InflatePaths64(subject, tt.delta, tt.joinType, tt.endType)
			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, expect %v", results, tt.expect)
			}
		})
	}
}

func TestInflatePathsD(t *testing.T) {
	subject := goclipper2.PathsD{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}

	tests := []struct {
		name        string
		delta       float64
		joinType    goclipper2.JoinType
		endType     goclipper2.EndType
		overSubject goclipper2.PathsD
		overClip    goclipper2.PathsD
		expect      goclipper2.PathsD
	}{
		{
			name:     "offset polygon with squire join type",
			delta:    40,
			joinType: goclipper2.Square,
			endType:  goclipper2.Polygon,
			expect: goclipper2.PathsD{
				{{140.00, -16.57}, {140.00, 116.57}, {116.57, 140.00}, {-16.57, 140.00}, {-40.00, 116.57}, {-40.00, -16.57}, {-16.57, -40.00}, {116.57, -40.00}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			results := goclipper2.InflatePathsD(subject, tt.delta, tt.joinType, tt.endType)
			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, expect %v", results, tt.expect)
			}
		})
	}
}
