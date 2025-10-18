package go_clipper2_test

import (
	"fmt"
	"reflect"
	"testing"

	goclipper2 "github.com/bolom009/go-clipper2"
	"github.com/stretchr/testify/assert"
)

func TestBooleanOpPathsD(t *testing.T) {
	var (
		subject = goclipper2.PathsD{
			{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		}
		clip = goclipper2.PathsD{
			{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
		}
	)

	tests := []struct {
		name        string
		clipType    goclipper2.ClipType
		fillRune    goclipper2.FillRule
		overSubject goclipper2.PathsD
		overClip    goclipper2.PathsD
		expect      goclipper2.PathsD
	}{
		{
			name:        "union non zero without clip",
			clipType:    goclipper2.Union,
			fillRune:    goclipper2.NonZero,
			overSubject: goclipper2.PathsD{subject[0], clip[0]},
			expect: goclipper2.PathsD{
				{{100, 50}, {150, 50}, {150, 150}, {50, 150}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:     "union with clip non zero",
			clipType: goclipper2.Union,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.PathsD{
				{{100, 50}, {150, 50}, {150, 150}, {50, 150}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:     "intersection with clip non zero",
			clipType: goclipper2.Intersection,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.PathsD{
				{{100, 100}, {50, 100}, {50, 50}, {100, 50}},
			},
		},
		{
			name:     "difference with clip non zero",
			clipType: goclipper2.Difference,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.PathsD{
				{{100, 50}, {50, 50}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:     "xor with clip non zero",
			clipType: goclipper2.Xor,
			fillRune: goclipper2.NonZero,
			expect: goclipper2.PathsD{
				{{150, 150}, {50, 150}, {50, 100}, {100, 100}, {100, 50}, {150, 50}},
				{{100, 50}, {50, 50}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:     "union with clip even odd",
			clipType: goclipper2.Union,
			fillRune: goclipper2.EvenOdd,
			expect: goclipper2.PathsD{
				{{100, 50}, {150, 50}, {150, 150}, {50, 150}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
		{
			name:        "union even odd without clip",
			clipType:    goclipper2.Union,
			fillRune:    goclipper2.EvenOdd,
			overSubject: goclipper2.PathsD{subject[0], clip[0]},
			expect: goclipper2.PathsD{
				{{150, 150}, {50, 150}, {50, 100}, {100, 100}, {100, 50}, {150, 50}},
				{{100, 50}, {50, 50}, {50, 100}, {0, 100}, {0, 0}, {100, 0}},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, tt.name), func(t *testing.T) {
			var results goclipper2.PathsD
			if tt.overSubject != nil && tt.overClip != nil {
				results = goclipper2.BooleanOpPathsD(tt.clipType, tt.overSubject, tt.overClip, tt.fillRune)
			} else if tt.overSubject != nil && tt.overClip == nil {
				results = goclipper2.BooleanOpPathsD(tt.clipType, tt.overSubject, nil, tt.fillRune)
			} else {
				results = goclipper2.BooleanOpPathsD(tt.clipType, subject, clip, tt.fillRune)
			}

			if !reflect.DeepEqual(results, tt.expect) {
				t.Errorf("got %v, expect %v", results, tt.expect)
			}
		})
	}
}

func TestPolyTreeD(t *testing.T) {
	subject := make(goclipper2.PathsD, 0)
	subject = append(subject, goclipper2.MakePathD(0, 0, 100, 0, 100, 100, 0, 100))

	subject = append(subject, goclipper2.MakePathD(10, 10, 10, 30, 25, 30, 25, 10))
	subject = append(subject, goclipper2.MakePathD(40, 10, 40, 30, 55, 30, 55, 10))
	subject = append(subject, goclipper2.MakePathD(70, 10, 70, 30, 85, 30, 85, 10))
	subject = append(subject, goclipper2.MakePathD(10, 40, 10, 90, 90, 90, 90, 40))

	subject = append(subject, goclipper2.MakePathD(20, 45, 80, 45, 80, 75, 20, 75))
	subject = append(subject, goclipper2.MakePathD(20, 80, 80, 80, 80, 85, 20, 85))

	subject = append(subject, goclipper2.MakePathD(30, 50, 30, 70, 45, 70, 45, 50))
	subject = append(subject, goclipper2.MakePathD(55, 50, 55, 70, 70, 70, 70, 50))

	polytree := goclipper2.BooleanOpPolyTreeD(goclipper2.Union, subject, nil, goclipper2.NonZero)
	assert.Equal(t, 1, len(polytree.GetChildren()))
	assert.Equal(t, 4, len(polytree.GetChildren()[0].GetChildren()))
	assert.Equal(t, 2, len(polytree.GetChildren()[0].GetChildren()[0].GetChildren()))
	assert.Equal(t, 2, len(polytree.GetChildren()[0].GetChildren()[0].GetChildren()[1].GetChildren()))
	/*
		Polytree with 1 polygon.
		  +- polygon (0) contains 4 holes.
			+- hole (0) contains 2 nested polygons.
			  +- polygon (1) contains 2 holes
	*/
}
