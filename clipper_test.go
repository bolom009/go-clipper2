package go_clipper2_test

import (
	"fmt"
	"reflect"
	"testing"

	goclipper2 "github.com/bolom009/go-clipper2"
)

var (
	subject = goclipper2.Paths64{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip = goclipper2.Paths64{
		{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
	}
)

func TestBooleanOpPaths64(t *testing.T) {
	tests := []struct {
		name        string
		clipType    goclipper2.ClipType
		fillRune    goclipper2.FillRule
		overSubject goclipper2.Paths64
		overClip    goclipper2.Paths64
		expect      goclipper2.Paths64
	}{
		{
			name:        "union non zero without clip",
			clipType:    goclipper2.Union,
			fillRune:    goclipper2.NonZero,
			overSubject: goclipper2.Paths64{subject[0], clip[0]},
			expect: goclipper2.Paths64{
				{{100, 50}, {150, 50}, {150, 150}, {50, 150}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:     "union with clip non zero",
			clipType: goclipper2.Union,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.Paths64{
				{{100, 50}, {150, 50}, {150, 150}, {50, 150}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:     "intersection with clip non zero",
			clipType: goclipper2.Intersection,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.Paths64{
				{{100, 100}, {50, 100}, {50, 50}, {100, 50}},
			},
		},
		{
			name:     "difference with clip non zero",
			clipType: goclipper2.Difference,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.Paths64{
				{{100, 50}, {50, 50}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:     "xor with clip non zero",
			clipType: goclipper2.Xor,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.Paths64{
				{{150, 150}, {50, 150}, {50, 100}, {100, 100}, {100, 50}, {150, 50}},
				{{100, 50}, {50, 50}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			var results goclipper2.Paths64
			if tt.overSubject != nil && tt.overClip != nil {
				results = goclipper2.BooleanOpPaths64(tt.clipType, tt.overSubject, tt.overClip, tt.fillRune)
			} else if tt.overSubject != nil && tt.overClip == nil {
				results = goclipper2.BooleanOpPaths64(tt.clipType, tt.overSubject, nil, tt.fillRune)
			} else {
				results = goclipper2.BooleanOpPaths64(tt.clipType, subject, clip, tt.fillRune)
			}

			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, expect %v", results, tt.expect)
			}
		})
	}
}
