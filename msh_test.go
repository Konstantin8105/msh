package msh_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Konstantin8105/msh"
)

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

func ExampleSort() {
	geo := geo()

	mesh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, "\nOriginal:\n%s", mesh)

	// removing
	mesh.Sort(msh.Triangle, msh.Quadrangle)
	mesh.Index1()
	fmt.Fprintf(os.Stdout, "\nAfter sort:\n%s", mesh)

	// Output:
	// Original:
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 5
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 3 0.000000 10.000000 0.000000
	// 4 5.000000 10.000000 0.000000
	// 5 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 10
	// 1 15 2 6 1 2
	// 2 15 2 1 2 3
	// 3 15 2 5 3 4
	// 4 1 2 2 1 2 4
	// 5 1 2 3 2 1 3
	// 6 1 2 4 3 1 2
	// 7 2 2 7 6 1 2 5
	// 8 2 2 7 6 4 3 5
	// 9 2 2 7 6 3 1 5
	// 10 2 2 7 6 2 4 5
	// $EndElements
	//
	// After sort:
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 5
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 3 0.000000 10.000000 0.000000
	// 4 5.000000 10.000000 0.000000
	// 5 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 10
	// 1 2 2 7 6 1 2 5
	// 2 2 2 7 6 4 3 5
	// 3 2 2 7 6 3 1 5
	// 4 2 2 7 6 2 4 5
	// 5 15 2 6 1 2
	// 6 15 2 1 2 3
	// 7 15 2 5 3 4
	// 8 1 2 2 1 2 4
	// 9 1 2 3 2 1 3
	// 10 1 2 4 3 1 2
	// $EndElements
}

func ExampleRemoveElements() {
	geo := geo()

	mesh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, "\nOriginal:\n%s", mesh)

	// removing
	mesh.RemoveElements(msh.Point, msh.Line, msh.Tetrahedron)
	mesh.Index1()
	fmt.Fprintf(os.Stdout, "\nAfter reindex:\n%s", mesh)

	// Output:
	// Original:
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 5
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 3 0.000000 10.000000 0.000000
	// 4 5.000000 10.000000 0.000000
	// 5 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 10
	// 1 15 2 6 1 2
	// 2 15 2 1 2 3
	// 3 15 2 5 3 4
	// 4 1 2 2 1 2 4
	// 5 1 2 3 2 1 3
	// 6 1 2 4 3 1 2
	// 7 2 2 7 6 1 2 5
	// 8 2 2 7 6 4 3 5
	// 9 2 2 7 6 3 1 5
	// 10 2 2 7 6 2 4 5
	// $EndElements
	//
	// After reindex:
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 5
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 3 0.000000 10.000000 0.000000
	// 4 5.000000 10.000000 0.000000
	// 5 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 4
	// 1 2 2 7 6 1 2 5
	// 2 2 2 7 6 4 3 5
	// 3 2 2 7 6 3 1 5
	// 4 2 2 7 6 2 4 5
	// $EndElements
}

func ExampleIndex1() {
	geo := geo()

	msh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, "\nOriginal:\n%s", msh)

	// remove node with index 3
	fmt.Fprintf(os.Stdout, "\nRemoving:\n")
	index := 3
again:
	for i, n := range msh.Nodes {
		if msh.Nodes[i].Id == index {
			fmt.Fprintf(os.Stdout, "Remove node id %d: %v\n", n.Id, n)
			msh.Nodes = append(msh.Nodes[:i], msh.Nodes[i+1:]...)
			goto again
		}
	}
	for k, el := range msh.Elements {
		for _, id := range msh.Elements[k].NodeId {
			if id == index {
				fmt.Fprintf(os.Stdout, "Remove element id %d: %v\n", el.Id, el)
				msh.Elements = append(msh.Elements[:k], msh.Elements[k+1:]...)
				goto again
			}
		}
	}
	fmt.Fprintf(os.Stdout, "\nAfter remove:\n%s", msh)

	// reindex
	msh.Index1()
	fmt.Fprintf(os.Stdout, "\nAfter reindex:\n%s", msh)

	// Output:
	// Original:
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 5
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 3 0.000000 10.000000 0.000000
	// 4 5.000000 10.000000 0.000000
	// 5 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 10
	// 1 15 2 6 1 2
	// 2 15 2 1 2 3
	// 3 15 2 5 3 4
	// 4 1 2 2 1 2 4
	// 5 1 2 3 2 1 3
	// 6 1 2 4 3 1 2
	// 7 2 2 7 6 1 2 5
	// 8 2 2 7 6 4 3 5
	// 9 2 2 7 6 3 1 5
	// 10 2 2 7 6 2 4 5
	// $EndElements
	//
	// Removing:
	// Remove node id 3: {3 [0 10 0]}
	// Remove element id 2: {2 15 [1 2] [3]}
	// Remove element id 5: {5 1 [3 2] [1 3]}
	// Remove element id 8: {8 2 [7 6] [4 3 5]}
	// Remove element id 9: {9 2 [7 6] [3 1 5]}
	//
	// After remove:
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 4
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 4 5.000000 10.000000 0.000000
	// 5 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 6
	// 1 15 2 6 1 2
	// 3 15 2 5 3 4
	// 4 1 2 2 1 2 4
	// 6 1 2 4 3 1 2
	// 7 2 2 7 6 1 2 5
	// 10 2 2 7 6 2 4 5
	// $EndElements
	//
	// After reindex:
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 4
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 3 5.000000 10.000000 0.000000
	// 4 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 6
	// 1 15 2 6 1 2
	// 2 15 2 5 3 3
	// 3 1 2 2 1 2 3
	// 4 1 2 4 3 1 2
	// 5 2 2 7 6 1 2 4
	// 6 2 2 7 6 2 3 4
	// $EndElements
}

func Example() {
	geo := geo()

	mshContent, err := msh.Generate(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, "%v", mshContent)

	msh, err := msh.New(geo)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, "%s", msh)

	// Output:
	// $MeshFormat
	// 2.2 0 8
	// $EndMeshFormat
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 5
	// 1 0 0 0
	// 2 5 0 0
	// 3 0 10 0
	// 4 5 10 0
	// 5 2.5 5 0
	// $EndNodes
	// $Elements
	// 10
	// 1 15 2 6 1 2
	// 2 15 2 1 2 3
	// 3 15 2 5 3 4
	// 4 1 2 2 1 2 4
	// 5 1 2 3 2 1 3
	// 6 1 2 4 3 1 2
	// 7 2 2 7 6 1 2 5
	// 8 2 2 7 6 4 3 5
	// 9 2 2 7 6 3 1 5
	// 10 2 2 7 6 2 4 5
	// $EndElements
	// $PhysicalNames
	// 7
	// 0 1 "NODE002"
	// 0 5 "NODE003"
	// 0 6 "NODE001"
	// 1 2 "LINE001"
	// 1 3 "LINE002"
	// 1 4 "LINE003"
	// 2 7 "PLANE006"
	// $EndPhysicalNames
	// $Nodes
	// 5
	// 1 0.000000 0.000000 0.000000
	// 2 5.000000 0.000000 0.000000
	// 3 0.000000 10.000000 0.000000
	// 4 5.000000 10.000000 0.000000
	// 5 2.500000 5.000000 0.000000
	// $EndNodes
	// $Elements
	// 10
	// 1 15 2 6 1 2
	// 2 15 2 1 2 3
	// 3 15 2 5 3 4
	// 4 1 2 2 1 2 4
	// 5 1 2 3 2 1 3
	// 6 1 2 4 3 1 2
	// 7 2 2 7 6 1 2 5
	// 8 2 2 7 6 4 3 5
	// 9 2 2 7 6 3 1 5
	// 10 2 2 7 6 2 4 5
	// $EndElements
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
