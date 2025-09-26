package go_clipper2_test

import (
	"fmt"
	"reflect"
	"testing"

	goclipper2 "github.com/bolom009/go-clipper2"
)

func TestMinkowskiSum64(t *testing.T) {
	pattern := goclipper2.Ellipse64(goclipper2.Point64{X: 100, Y: 100}, 30, 30, 0)

	tests := []struct {
		name        string
		isClosed    bool
		path        goclipper2.Path64
		overSubject goclipper2.Paths64
		expect      goclipper2.Paths64
	}{
		{
			name:     "minkowski sum64",
			isClosed: false,
			path:     goclipper2.MakePath64(0, 0, 200, 0, 200, 200, 0, 200, 0, 0),
			expect: goclipper2.Paths64{
				{{295, 70}, {305, 70}, {315, 74}, {323, 81}, {328, 90}, {330, 100}, {330, 300}, {328, 310}, {323, 319}, {315, 326}, {305, 330}, {295, 330}, {105, 330}, {95, 330}, {85, 326}, {77, 319}, {72, 310}, {70, 300}, {70, 100}, {72, 90}, {77, 81}, {85, 74}, {95, 70}},
				{{130, 130}, {130, 270}, {270, 270}, {270, 130}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			results := goclipper2.MinkowskiSum64(pattern, tt.path, tt.isClosed)
			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, expect %v", results, tt.expect)
			}
		})
	}
}

func TestMinkowskiSumD(t *testing.T) {
	pattern := goclipper2.EllipseD(goclipper2.PointD{X: 100, Y: 100}, 30, 30, 0)

	tests := []struct {
		name        string
		isClosed    bool
		path        goclipper2.PathD
		overSubject goclipper2.PathsD
		expect      goclipper2.PathsD
	}{
		{
			name:     "minkowski sum64",
			isClosed: false,
			path:     goclipper2.MakePathD(0, 0, 200, 0, 200, 200, 0, 200, 0, 0),
			expect: goclipper2.PathsD{
				{{294.79, 70.46}, {305.21, 70.46}, {315.00, 74.02}, {322.98, 80.72}, {328.19, 89.74}, {330, 100}, {330, 300}, {328.19, 310.26}, {322.98, 319.28}, {315.00, 325.98}, {305.21, 329.54}, {294.79, 329.54}, {105.21, 329.54}, {94.79, 329.54}, {85.00, 325.98}, {77.02, 319.28}, {71.81, 310.26}, {70, 300}, {70, 100}, {71.81, 89.74}, {77.02, 80.72}, {85.00, 74.02}, {94.79, 70.46}},
				{{130, 129.54}, {130, 270.46}, {270, 270.46}, {270, 129.54}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			results := goclipper2.MinkowskiSumD(pattern, tt.path, tt.isClosed)
			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, \n\t\t\t\t expect %v", results, tt.expect)
			}
		})
	}
}

func TestMinkowskiDiff64(t *testing.T) {
	pattern := goclipper2.Ellipse64(goclipper2.Point64{X: 100, Y: 100}, 30, 30, 0)

	tests := []struct {
		name        string
		isClosed    bool
		path        goclipper2.Path64
		overSubject goclipper2.Paths64
		expect      goclipper2.Paths64
	}{
		{
			name:     "minkowski diff64",
			isClosed: true,
			path:     goclipper2.MakePath64(0, 0, 200, 0, 200, 200, 0, 200, 0, 0),
			expect: goclipper2.Paths64{
				{{95, -130}, {105, -130}, {115, -126}, {123, -119}, {128, -110}, {130, -100}, {130, 100}, {128, 110}, {123, 119}, {115, 126}, {105, 130}, {95, 130}, {-95, 130}, {-105, 130}, {-115, 126}, {-123, 119}, {-128, 110}, {-130, 100}, {-130, -100}, {-128, -110}, {-123, -119}, {-115, -126}, {-105, -130}},
				{{-70, -70}, {-70, 70}, {70, 70}, {70, -70}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			results := goclipper2.MinkowskiDiff64(pattern, tt.path, tt.isClosed)
			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, expect %v", results, tt.expect)
			}
		})
	}
}

func TestMinkowskiDiffD(t *testing.T) {
	pattern := goclipper2.EllipseD(goclipper2.PointD{X: 100, Y: 100}, 30, 30, 0)

	tests := []struct {
		name        string
		isClosed    bool
		path        goclipper2.PathD
		overSubject goclipper2.PathsD
		expect      goclipper2.PathsD
	}{
		{
			name:     "minkowski diffD",
			isClosed: true,
			path:     goclipper2.MakePathD(0, 0, 200, 0, 200, 200, 0, 200, 0, 0),
			expect: goclipper2.PathsD{
				{{94.79, -129.54}, {105.21, -129.54}, {115.00, -125.98}, {122.98, -119.28}, {128.19, -110.26}, {130.00, -100.00}, {130.00, 100.00}, {128.19, 110.26}, {122.98, 119.28}, {115.00, 125.98}, {105.21, 129.54}, {94.79, 129.54}, {-94.79, 129.54}, {-105.21, 129.54}, {-115.00, 125.98}, {-122.98, 119.28}, {-128.19, 110.26}, {-130.00, 100.00}, {-130.00, -100.00}, {-128.19, -110.26}, {-122.98, -119.28}, {-115.00, -125.98}, {-105.21, -129.54}},
				{{-70.00, -70.46}, {-70.00, 70.46}, {70.00, 70.46}, {70.00, -70.46}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			results := goclipper2.MinkowskiDiffD(pattern, tt.path, tt.isClosed)
			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, expect %v", results, tt.expect)
			}
		})
	}
}
