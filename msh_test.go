package msh_test

import (
	"fmt"
	"testing"

	"github.com/Konstantin8105/msh"
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
