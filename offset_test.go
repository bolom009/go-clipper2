package go_clipper2_test

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"testing"

	goclipper2 "github.com/bolom009/go-clipper2"
	"github.com/stretchr/testify/assert"
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
		{
			name:     "offset large uncommon polygon",
			delta:    -3,
			joinType: goclipper2.Miter,
			endType:  goclipper2.Polygon,
			overSubject: goclipper2.PathsD{
				{{X: 398.09, Y: -288.96}, {X: 375.44, Y: -263.02}, {X: 401.06, Y: -242.65}, {X: 387.56, Y: -208.36}, {X: 429.13, Y: -146.17}, {X: 454.03, Y: -146.17}, {X: 454.03, Y: -111.73}, {X: 486.73, Y: -113.24}, {X: 488.21, Y: -109.07}, {X: 529.19, Y: -125.86}, {X: 529.19, Y: -162.1}, {X: 557.37, Y: -162.1}, {X: 557.37, Y: -126.36}, {X: 584.38, Y: -202.03}, {X: 556.79, Y: -202.03}, {X: 556.79, Y: -236.47}, {X: 641.93, Y: -236.47}, {X: 641.93, Y: -202.03}, {X: 674.63, Y: -203.54}, {X: 701.64, Y: -127.88}, {X: 668.94, Y: -126.36}, {X: 668.94, Y: -88.4}, {X: 529.19, Y: -88.4}, {X: 529.19, Y: -94.81}, {X: 497.89, Y: -81.97}, {X: 513.74, Y: -37.58}, {X: 481.04, Y: -36.06}, {X: 481.04, Y: 1.9}, {X: 341.29, Y: 1.9}, {X: 341.29, Y: -5.98}, {X: 291.13, Y: 1.21}, {X: 297.64, Y: 19.42}, {X: 264.94, Y: 20.94}, {X: 264.94, Y: 58.9}, {X: 125.19, Y: 58.9}, {X: 125.19, Y: 44.37}, {X: 85.44, Y: 44.37}, {X: 90.19, Y: 57.68}, {X: 57.5, Y: 59.19}, {X: 57.5, Y: 97.16}, {X: -82.25, Y: 97.16}, {X: -82.25, Y: 23.45}, {X: -54.07, Y: 23.45}, {X: -54.07, Y: 59.19}, {X: -27.06, Y: -16.48}, {X: -54.65, Y: -16.48}, {X: -54.65, Y: -50.91}, {X: -37.98, Y: -50.91}, {X: -71.39, Y: -102.3}, {X: -161.51, Y: -102.3}, {X: -161.51, Y: -176.0}, {X: -133.33, Y: -176.0}, {X: -133.33, Y: -140.26}, {X: -106.32, Y: -215.93}, {X: -133.91, Y: -215.93}, {X: -133.91, Y: -250.37}, {X: -48.77, Y: -250.37}, {X: -48.77, Y: -215.93}, {X: -16.07, Y: -217.44}, {X: -1.6, Y: -176.93}, {X: 45.59, Y: -165.55}, {X: 45.59, Y: -172.6}, {X: 73.77, Y: -172.6}, {X: 73.77, Y: -136.86}, {X: 100.78, Y: -212.53}, {X: 73.19, Y: -212.53}, {X: 73.19, Y: -246.97}, {X: 107.06, Y: -246.97}, {X: 107.06, Y: -280.42}, {X: 95.65, Y: -280.82}, {X: 79.22, Y: -311.72}, {X: 97.77, Y: -341.4}, {X: 132.75, Y: -340.18}, {X: 144.21, Y: -318.61}, {X: 242.42, Y: -282.09}, {X: 264.25, Y: -307.09}, {X: 285.48, Y: -288.56}, {X: 261.97, Y: -261.64}, {X: 332.08, Y: -300.87}, {X: 311.3, Y: -319.02}, {X: 333.95, Y: -344.96}},
			},
			expect: goclipper2.PathsD{
				{{393.85, -288.67}, {371.11, -262.63}, {397.45, -241.67}, {384.19, -208}, {427.53, -143.17}, {451.03, -143.17}, {451.03, -108.59}, {484.64, -110.13}, {486.44, -105.1}, {532.19, -123.85}, {532.19, -159.1}, {554.37, -159.1}, {554.37, -123.83}, {559.35, -122.97}, {588.64, -205.03}, {559.79, -205.03}, {559.79, -233.47}, {638.93, -233.47}, {638.93, -198.89}, {672.54, -200.43}, {697.44, -130.69}, {665.94, -129.22}, {665.94, -91.4}, {532.19, -91.4}, {532.19, -99.28}, {494.1, -83.66}, {509.54, -40.39}, {478.04, -38.92}, {478.04, -1.1}, {344.29, -1.1}, {344.29, -9.44}, {287.07, -1.24}, {293.45, 16.61}, {261.94, 18.08}, {261.94, 55.9}, {128.19, 55.9}, {128.19, 41.37}, {81.18, 41.37}, {85.99, 54.87}, {54.5, 56.33}, {54.5, 94.16}, {-79.25, 94.16}, {-79.25, 26.45}, {-57.07, 26.45}, {-57.07, 61.72}, {-52.09, 62.58}, {-22.8, -19.48}, {-51.65, -19.48}, {-51.65, -47.91}, {-36.19, -47.91}, {-34.49, -51.05}, {-69.76, -105.3}, {-158.51, -105.3}, {-158.51, -173}, {-136.33, -173}, {-136.33, -137.73}, {-131.35, -136.87}, {-102.06, -218.93}, {-130.91, -218.93}, {-130.91, -247.37}, {-51.77, -247.37}, {-51.77, -212.79}, {-18.16, -214.33}, {-3.88, -174.39}, {48.59, -161.74}, {48.59, -169.6}, {70.77, -169.6}, {70.77, -134.33}, {75.75, -133.47}, {105.04, -215.53}, {76.19, -215.53}, {76.19, -243.97}, {110.06, -243.97}, {110.06, -283.32}, {97.49, -283.75}, {82.68, -311.6}, {99.39, -338.33}, {130.92, -337.24}, {142.1, -316.2}, {243.32, -278.56}, {264.53, -302.86}, {281.24, -288.27}, {258.05, -261.71}, {261.23, -257.79}, {337.26, -300.33}, {315.53, -319.3}, {334.23, -340.72}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			if tt.overSubject != nil {
				subject = tt.overSubject
			}

			results := goclipper2.InflatePathsD(subject, tt.delta, tt.joinType, tt.endType)
			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, \n\t\t    expect %v", results, tt.expect)
			}
		})
	}
}

func TestChainInflatePathsD(t *testing.T) {
	circle := circlePath(5.0, 5.0, 3.0, 32)
	circle2 := circlePath(7.0, 7.0, 1.0, 32)
	rectangle := goclipper2.MakePathD(0.0, 0.0, 5.0, 0.0, 5.0, 6.0, 0.0, 6.0)
	result1 := goclipper2.DifferenceWithClipPathsD(goclipper2.PathsD{circle}, goclipper2.PathsD{circle2, rectangle}, goclipper2.EvenOdd)
	result2 := goclipper2.InflatePathsD(result1, 1.0, goclipper2.Round, goclipper2.Polygon, goclipper2.WithMitterLimit(0.0))

	assert.Equal(t, 1, len(result2))
	assert.Equal(t, 117, len(result2[0]))
}

func circlePath(offsetX, offsetY, radius float64, segments int) goclipper2.PathD {
	pts := make(goclipper2.PathD, 0, segments)
	for i := 0; i < segments; i++ {
		angle := float64(i) / float64(segments) * 2.0 * math.Pi
		x := math.Sin(angle)*radius + offsetX
		y := math.Cos(angle)*radius + offsetY
		pts = append(pts, goclipper2.PointD{X: x, Y: y})
	}
	return pts
}

func TestOffsetCallback(t *testing.T) {
	const scale = 10
	const delta = 10.0 * scale

	// ellipse := goclipper2.Ellipse64(goclipper2.Point64{X: 0, Y: 0}, 200*scale, 180*scale, 256)
	ellipse := goclipper2.PathDToPath64(circlePath(0, 0, scale*180, 100))
	solution := goclipper2.Paths64{}

	co := goclipper2.NewClipperOffset(0.0, 10.0, true, false)
	co.AddPaths(goclipper2.Paths64{ellipse}, goclipper2.Miter, goclipper2.RoundET)

	var deltaFunc goclipper2.DeltaCallbackFunc = func(path *goclipper2.Path64, path_normals *goclipper2.PathD, curr_idx, prev_idx uint8) float64 {
		// gradually scale down the offset to a minimum of 25% of delta
		midIndex := (uint8(len(*path)) / 2)
		factor := 1.0 - (float64(curr_idx) / float64(midIndex) * 0.75)
		return delta * factor
	}
	co.SetDeltaCallback(&deltaFunc)
	co.Execute64(10.0, &solution)
	log.Println(solution)

	// Visualize multiple paths
	html := goclipper2.VisualizePaths64HTML(solution, "My Clipped Paths")

	// Save to file
	os.WriteFile("output.html", []byte(html), 0644)
}

// Test1 - Variable offset callback that gradually scales down
func TestOffsetVariableCallback1(t *testing.T) {
	const scale = 10
	delta := 10.0 * scale

	co := goclipper2.NewClipperOffset(0.0, 10.0, true, false)

	// Create ellipse
	subject := goclipper2.Paths64{goclipper2.Ellipse64(goclipper2.Point64{X: 0, Y: 0}, 200*scale, 180*scale, 256)}
	// Resize to 90% as in C++ version
	if len(subject[0]) > 0 {
		newLen := int(float64(len(subject[0])) * 0.9)
		subject[0] = subject[0][:newLen]
	}

	co.AddPaths(subject, goclipper2.Miter, goclipper2.RoundET)

	var deltaFunc goclipper2.DeltaCallbackFunc = func(path *goclipper2.Path64, path_normals *goclipper2.PathD, curr_idx, prev_idx uint8) float64 {
		// gradually scale down the offset to a minimum of 25% of delta
		high := float64(len(*path)-1) * 1.25
		return (high - float64(curr_idx)) / high * delta
	}

	co.SetDeltaCallback(&deltaFunc)
	solution := goclipper2.Paths64{}
	co.Execute64(1.0, &solution)

	html := goclipper2.VisualizePaths64HTML(append(subject, solution...), "Test1: Variable Offset - Gradual Scale Down")
	os.WriteFile("test1.html", []byte(html), 0644)
	t.Logf("Test1 completed with %d solution paths", len(solution))
}

// Test2 - Variable offset callback based on distance from middle
func TestOffsetVariableCallback2(t *testing.T) {
	const scale = 10
	delta := 10.0 * scale

	co := goclipper2.NewClipperOffset(0.0, 10.0, true, false)

	// Create ellipse
	subject := goclipper2.Paths64{goclipper2.Ellipse64(goclipper2.Point64{X: 0, Y: 0}, 200*scale, 180*scale, 256)}
	// Resize to 90% as in C++ version
	if len(subject[0]) > 0 {
		newLen := int(float64(len(subject[0])) * 0.9)
		subject[0] = subject[0][:newLen]
	}

	co.AddPaths(subject, goclipper2.Miter, goclipper2.RoundET)

	var deltaFunc goclipper2.DeltaCallbackFunc = func(path *goclipper2.Path64, path_normals *goclipper2.PathD, curr_idx, prev_idx uint8) float64 {
		// calculate offset based on distance from the middle of the path
		midIdx := float64(len(*path)) / 2.0
		absDistance := math.Abs(float64(curr_idx) - midIdx)
		return delta * (1.0 - 0.70*(absDistance/midIdx))
	}

	co.SetDeltaCallback(&deltaFunc)
	solution := goclipper2.Paths64{}
	co.Execute64(1.0, &solution)

	html := goclipper2.VisualizePaths64HTML(append(subject, solution...), "Test2: Variable Offset - Middle-based")
	os.WriteFile("test2.html", []byte(html), 0644)
	t.Logf("Test2 completed with %d solution paths", len(solution))
}

// Test3 - Variable offset using normal vectors
func TestOffsetVariableCallback3(t *testing.T) {
	radius := 5000.0
	subject := goclipper2.Paths64{goclipper2.Ellipse64(goclipper2.Point64{X: 0, Y: 0}, radius, radius, 200)}

	co := goclipper2.NewClipperOffset(0.0, 10.0, true, false)
	co.AddPaths(subject, goclipper2.Miter, goclipper2.Polygon)

	var deltaFunc goclipper2.DeltaCallbackFunc = func(path *goclipper2.Path64, path_normals *goclipper2.PathD, curr_idx, prev_idx uint8) float64 {
		// when multiplying the x & y of edge unit normal vectors, the value will be
		// largest (0.5) when edges are at 45 deg. and least (-0.5) at negative 45 deg.
		norm := (*path_normals)[curr_idx]
		delta := norm.Y * norm.X
		return radius*0.5 + radius*delta
	}

	co.SetDeltaCallback(&deltaFunc)
	solution := goclipper2.Paths64{}
	co.Execute64(1.0, &solution)

	html := goclipper2.VisualizePaths64HTML(append(subject, solution...), "Test3: Variable Offset - Normal Vector Based")
	os.WriteFile("test3.html", []byte(html), 0644)
	t.Logf("Test3 completed with %d solution paths", len(solution))
}

// Test4 - Variable offset using sin of edge angle
func TestOffsetVariableCallback4(t *testing.T) {
	const scale = 100
	subject := goclipper2.Paths64{goclipper2.Ellipse64(goclipper2.Point64{X: 10 * scale, Y: 10 * scale}, 40*scale, 40*scale, 256)}

	co := goclipper2.NewClipperOffset(0.0, 10.0, true, false)
	co.AddPaths(subject, goclipper2.Round, goclipper2.RoundET)

	var deltaFunc goclipper2.DeltaCallbackFunc = func(path *goclipper2.Path64, path_normals *goclipper2.PathD, curr_idx, prev_idx uint8) float64 {
		sinEdge := (*path_normals)[curr_idx].Y
		return sinEdge * sinEdge * 3 * scale
	}

	co.SetDeltaCallback(&deltaFunc)
	solution := goclipper2.Paths64{}
	co.Execute64(1.0, &solution)

	html := goclipper2.VisualizePaths64HTML(append(subject, solution...), "Test4: Variable Offset - Sin Edge Based")
	os.WriteFile("test4.html", []byte(html), 0644)
	t.Logf("Test4 completed with %d solution paths", len(solution))
}

// Test5 - Variable offset on a line with quadratic delta
func TestOffsetVariableCallback5(t *testing.T) {
	subject := goclipper2.Paths64{goclipper2.MakePath64(0, 0, 20, 0, 40, 0, 60, 0, 80, 0, 100, 0)}

	co := goclipper2.NewClipperOffset(0.0, 10.0, true, false)
	co.AddPaths(subject, goclipper2.Round, goclipper2.Butt)

	var deltaFunc goclipper2.DeltaCallbackFunc = func(path *goclipper2.Path64, path_normals *goclipper2.PathD, curr_idx, prev_idx uint8) float64 {
		return float64(curr_idx*curr_idx) + 10
	}

	co.SetDeltaCallback(&deltaFunc)
	solution := goclipper2.Paths64{}
	co.Execute64(1.0, &solution)

	html := goclipper2.VisualizePaths64HTML(append(subject, solution...), "Test5: Variable Offset - Quadratic Delta")
	os.WriteFile("test5.html", []byte(html), 0644)
	t.Logf("Test5 completed with %d solution paths", len(solution))
}
