package msh_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Konstantin8105/compare"
	"github.com/Konstantin8105/msh"
)

func init() {
	if runtime.GOOS == "windows" {
		// only for work computer
		*msh.GmshApp = "Z:\\Software\\Gmsh\\gmsh.exe"
	}
}

func geo() string {
	var geo string
	geo += fmt.Sprintf("h   = %.5f;\n", 10.0)
	geo += fmt.Sprintf("thk = %.5f;\n", 5.00)
	geo += fmt.Sprintf("Lc  = %.5f;\n", 20.00)

	geo += `
	Point(000) = {+0.0000,+0.0000,+0.0000,Lc};
	Point(001) = {thk    ,+0.0000,+0.0000,Lc};
	Point(002) = {+0.0000,h      ,+0.0000,Lc};
	Point(003) = {thk    ,h      ,+0.0000,Lc};
	Physical Point("NODE002") = {002};
	Line(1) = {1, 3};
	Physical Curve("LINE001") = {1};
	Line(2) = {0, 2};
	Physical Curve("LINE002") = {2};
	Line(3) = {0, 1};
	Physical Curve("LINE003") = {3};
	Line(4) = {2, 3};
	Physical Point("NODE003") = {003};
	Line Loop(5) = {1, -4, -2, 3};
	Physical Point("NODE001") = {001};
	Plane Surface(6) = {5};
	Physical Surface("PLANE006") = {6};`
	return geo
}

// tf is testdata file
func tf(file string) string {
	return filepath.Join("testdata", file)
}

func TestSort(t *testing.T) {
	geo := geo()

	var buf bytes.Buffer
	mesh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(&buf, "\nOriginal:\n%s", mesh)

	// removing
	fmt.Fprintf(&buf, "---------------\n")
	mesh.Sort(msh.Triangle, msh.Quadrangle)
	mesh.Index1()
	fmt.Fprintf(&buf, "\nAfter sort:\n%s", mesh)

	compare.Test(t, tf("Sort"), buf.Bytes())
}

func TestRemoveElements(t *testing.T) {
	geo := geo()

	var buf bytes.Buffer
	mesh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(&buf, "\nOriginal:\n%s", mesh)

	// removing
	fmt.Fprintf(&buf, "---------------\n")
	mesh.RemoveElements(msh.Point, msh.Line, msh.Tetrahedron)
	mesh.Index1()
	fmt.Fprintf(&buf, "\nAfter reindex:\n%s", mesh)

	compare.Test(t, tf("RemoveElements"), buf.Bytes())
}

func TestIndex1(t *testing.T) {
	geo := geo()

	var buf bytes.Buffer
	msh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(&buf, "\nOriginal:\n%s", msh)

	// remove node with index 3
	fmt.Fprintf(&buf, "\nRemoving:\n")
	index := 3
again:
	fmt.Fprintf(&buf, "---------------\n")
	for i, n := range msh.Nodes {
		if msh.Nodes[i].Id == index {
			fmt.Fprintf(&buf, "Remove node id %d: %v\n", n.Id, n)
			msh.Nodes = append(msh.Nodes[:i], msh.Nodes[i+1:]...)
			goto again
		}
	}
	for k, el := range msh.Elements {
		for _, id := range msh.Elements[k].NodeId {
			if id == index {
				fmt.Fprintf(&buf, "Remove element id %d: %v\n", el.Id, el)
				msh.Elements = append(msh.Elements[:k], msh.Elements[k+1:]...)
				goto again
			}
		}
	}
	fmt.Fprintf(&buf, "\nAfter remove:\n%s", msh)

	// reindex
	fmt.Fprintf(&buf, "---------------\n")
	msh.Index1()
	fmt.Fprintf(&buf, "\nAfter reindex:\n%s", msh)

	compare.Test(t, tf("Index1"), buf.Bytes())
}

func Test(t *testing.T) {
	geo := geo()

	var buf bytes.Buffer
	mshContent, err := msh.Generate(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(&buf, "%v", mshContent)

	fmt.Fprintf(&buf, "---------------\n")
	msh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(&buf, "%s", msh)

	compare.Test(t, tf("Test"), buf.Bytes())
}

func TestFail(t *testing.T) {
	if _, err := msh.New("fail"); err == nil {
		t.Fatal("New")
	}
	if _, err := msh.Generate("fail"); err == nil {
		t.Fatal("Generate")
	}
	if _, err := msh.Parse(`
$PhysicalNames
7
0 1 fail "NODE002"
0 5 fail "NODE003"
0 6 fail "NODE001"
1 2 fail "LINE001"
1 3 fail "LINE002"
1 4 fail "LINE003"
2 7 fail "PLANE00"
$EndPhysicalNames`); err == nil {
		t.Fatal("Parse")
	}
}

func TestGetNode(t *testing.T) {
	geo := geo()
	msh, err := msh.New(geo)
	if err != nil {
		t.Fatal(err)
	}
	for i := range msh.Elements {
		for k := range msh.Elements[i].NodeId {
			ni := msh.Elements[i].NodeId[k]
			index := msh.GetNode(ni)
			if index < 0 {
				t.Errorf("Not found")
			}
		}
	}
	if index := msh.GetNode(10000); 0 <= index {
		t.Errorf("Found")
	}
}
