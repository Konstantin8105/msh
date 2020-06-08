package msh_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/Konstantin8105/msh"
	"github.com/Konstantin8105/pow"
)

func Test(t *testing.T) {
	var geo string
	geo += fmt.Sprintf("h   = %.5f;\n", 10.0)
	geo += fmt.Sprintf("thk = %.5f;\n", 5.00)
	geo += fmt.Sprintf("Lc  = %.5f;\n", 20.00)

	geo += `
	Point(000) = {+0.0000,+0.0000,+0.0000,Lc};
	Point(001) = {thk    ,+0.0000,+0.0000,Lc};
	Point(002) = {+0.0000,h      ,+0.0000,Lc};
	Point(003) = {thk    ,h      ,+0.0000,Lc};
	Line(1) = {1, 3};
	Line(2) = {0, 2};
	Line(3) = {0, 1};
	Line(4) = {2, 3};
	Line Loop(5) = {1, -4, -2, 3};
	Plane Surface(6) = {5};
`

	msh, err := msh.New(geo)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", msh)
}

func TestRotateXOY(t *testing.T) {
	tcs := []struct {
		angle  float64
		point  msh.Point
		expect msh.Point
	}{
		{
			angle:  0.0,
			point:  msh.Point{X: 1, Y: 0},
			expect: msh.Point{X: 1, Y: 0},
		},
		{
			angle:  0.0,
			point:  msh.Point{X: 1, Y: 1},
			expect: msh.Point{X: 1, Y: 1},
		},
		{
			angle:  0.0,
			point:  msh.Point{X: 0, Y: 1},
			expect: msh.Point{X: 0, Y: 1},
		},
		{
			angle:  0.0,
			point:  msh.Point{X: -1, Y: 0},
			expect: msh.Point{X: -1, Y: 0},
		},
		{
			angle:  0.0,
			point:  msh.Point{X: -1, Y: -1},
			expect: msh.Point{X: -1, Y: -1},
		},
		{
			angle:  0.0,
			point:  msh.Point{X: 0, Y: -1},
			expect: msh.Point{X: 0, Y: -1},
		},
		{
			angle:  math.Pi / 2.0,
			point:  msh.Point{X: 1, Y: 0},
			expect: msh.Point{X: 0, Y: 1},
		},
		{
			angle:  -math.Pi / 2.0,
			point:  msh.Point{X: 1, Y: 0},
			expect: msh.Point{X: 0, Y: -1},
		},
		{
			angle:  -math.Pi / 2.0,
			point:  msh.Point{X:-0.01730271278289151,Y: -0.04784164326597217},
			expect: msh.Point{X:-0.04784164326597217,Y: +0.01730271278289151},
		},
	}
	eps := 1e-6
	for index, tc := range tcs {
		t.Run(fmt.Sprintf("%d", index), func(t *testing.T) {
			m := msh.Msh{
				Points: []msh.Point{
					tc.point,
				},
			}
			m.RotateXOY(tc.angle)
			act := m.Points[0]
			actEps := math.Sqrt(pow.E2(tc.expect.X-act.X) + pow.E2(tc.expect.Y-act.Y))
			t.Logf("actual exp %e", actEps)
			if actEps > eps {
				t.Errorf("act = %#v\nexpect = %#v", act, tc.expect)
			}
		})
	}
}
