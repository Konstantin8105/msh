// The MSH file format version 2:
//
// https://gmsh.info/doc/texinfo/gmsh.html#MSH-file-format-version-2-_0028Legacy_0029
//
//	$MeshFormat
//	version-number file-type data-size
//	$EndMeshFormat
//
//	$PhysicalNames
//	number-of-names
//	physical-dimension physical-tag "physical-name"
//	…
//	$EndPhysicalNames
//
//	$Nodes
//	number-of-nodes
//	node-number x-coord y-coord z-coord
//	…
//	$EndNodes
//
//	$Elements
//	number-of-elements
//	elm-number elm-type number-of-tags < tag > … node-number-list
//	…
//	$EndElements
package msh

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type ElementType int

const (
	Point       ElementType = 15
	Line                    = 1
	Triangle                = 2
	Quadrangle              = 3
	Tetrahedron             = 4
)

type PhysicalName struct {
	Dimension int
	Tag       int
	Name      string
}

type Element struct {
	Id     int
	EType  ElementType
	Tags   []int
	NodeId []int
}

type Node struct {
	Id    int
	Coord [3]float64
}

type Msh struct {
	PhysicalNames []PhysicalName
	Nodes         []Node
	Elements      []Element
}

func (msh *Msh) Sort(ets ...ElementType) {
	pos := func(et ElementType) int{
		for i :=range ets {
			if ets[i] == et {
				return i
			}
		}
		return len(ets)
	}
	sort.Slice(msh.Elements, func(i,j int) bool {
		return pos(msh.Elements[i].EType) < pos(msh.Elements[j].EType)
	})
}

func (msh *Msh) RemoveElements(ets ...ElementType) {
	for i := len(msh.Elements)-1; 0 <= i; i-- {
		remove := false
		for k := range ets {
			if ets[k] == msh.Elements[i].EType {
				remove =true
			}
		}
		if !remove {
			continue
		}
		msh.Elements = append(msh.Elements[:i],msh.Elements[i+1:]...)
	}
}

func (msh *Msh) Index1() {
	maxIndex := 0
	for _, n := range msh.Nodes {
		if maxIndex < n.Id {
			maxIndex = n.Id
		}
	}
	newId := make([]int, maxIndex+1)
	for id, n := range msh.Nodes {
		newId[n.Id] = id+1
	}
	for i := range msh.Elements {
		for j := range msh.Elements[i].NodeId {
			nid := &msh.Elements[i].NodeId[j]
			*nid = newId[*nid]
		}
	}
	for i := range msh.Nodes {
		msh.Nodes[i].Id = i + 1
	}
	for i := range msh.Elements {
		msh.Elements[i].Id = i + 1
	}
}

func (msh Msh) GetNode(Id int) (index int) {
	index = sort.Search(len(msh.Nodes), func(i int) bool { return msh.Nodes[i].Id >= Id })
	if index < len(msh.Nodes) && msh.Nodes[index].Id == Id {
		// x is present at data[i]
		return
	}
	// x is not present in data,
	// but i is the index where it would be inserted.
	for i := range msh.Nodes {
		if msh.Nodes[i].Id == Id {
			return i
		}
	}
	return -1
}

func (msh Msh) String() string {
	var out string
	if 0 < len(msh.PhysicalNames) {
		out += "$PhysicalNames\n"
		out += fmt.Sprintf("%d\n", len(msh.PhysicalNames))
		for _, pn := range msh.PhysicalNames {
			out += fmt.Sprintf("%v %d \"%s\"\n",
				pn.Dimension, pn.Tag, pn.Name)
		}
		out += "$EndPhysicalNames\n"
	}
	if 0 < len(msh.Nodes) {
		out += "$Nodes\n"
		out += fmt.Sprintf("%d\n", len(msh.Nodes))
		for _, n := range msh.Nodes {
			out += fmt.Sprintf("%d %f %f %f\n",
				n.Id, n.Coord[0], n.Coord[1], n.Coord[2])
		}
		out += "$EndNodes\n"
	}
	if 0 < len(msh.Elements) {
		out += "$Elements\n"
		out += fmt.Sprintf("%d\n", len(msh.Elements))
		for _, el := range msh.Elements {
			out += fmt.Sprintf("%d %d ", el.Id, el.EType)
			out += fmt.Sprintf("%d", len(el.Tags))
			for _, t := range el.Tags {
				out += fmt.Sprintf(" %d", t)
			}
			for _, np := range el.NodeId {
				out += fmt.Sprintf(" %d", np)
			}
			out += "\n"
		}
		out += "$EndElements\n"
	}
	return out
}

func New(geoContent string) (m *Msh, err error) {
	msh, err := Generate(geoContent)
	if err != nil {
		return
	}
	return Parse(msh)
}

func Generate(geoContent string) (mshContent string, err error) {
	// create temp directory
	var dir string
	dir, err = ioutil.TempDir("", "msh")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir) // clean up

	// create geo file
	geofn := filepath.Join(dir, "m.geo")
	if err = ioutil.WriteFile(geofn, []byte(geoContent), 0666); err != nil {
		return
	}

	// run gmsh
	meshfn := filepath.Join(dir, "m.msh")
	if err = exec.Command("gmsh",
		"-format", "msh2", // Format: MSH2
		"-smooth", "10", // Smooth mesh
		"-2", // 2D mesh generation
		geofn, meshfn).Run(); err != nil {
		return
	}
	// read msh
	meshContent, err := ioutil.ReadFile(meshfn)
	if err != nil {
		return
	}
	return string(meshContent), nil
}

func Parse(meshContent string) (msh *Msh, err error) {
	msh = new(Msh)

	// split by lines
	lines := strings.Split(string(meshContent), "\n")

	// PhysicalNames
	for _, line := range getLines(lines, "$PhysicalNames", "$EndPhysicalNames") {
		fs := strings.Fields(line)
		if len(fs) != 3 {
			err = fmt.Errorf("PhysicalNames error: %v", line)
			return
		}
		var (
			dim int
			tag int
		)
		dim, err = strconv.Atoi(fs[0])
		if err != nil {
			err = fmt.Errorf("PhysicalNames error: not valid dim - %v. %v", line, err)
			return
		}
		tag, err = strconv.Atoi(fs[1])
		if err != nil {
			err = fmt.Errorf("PhysicalNames error: not valid tag - %v. %v", line, err)
			return
		}
		name := fs[2][1 : len(fs[2])-1]
		msh.PhysicalNames = append(msh.PhysicalNames, PhysicalName{
			Dimension: dim,
			Tag:       tag,
			Name:      name,
		})
	}

	// Nodes
	for _, line := range getLines(lines, "$Nodes", "$EndNodes") {
		fs := strings.Fields(line)
		if len(fs) != 4 {
			err = fmt.Errorf("PhysicalNames error: %v", line)
			return
		}
		var (
			id    int
			coord [3]float64
		)
		if id, err = strconv.Atoi(fs[0]); err != nil {
			err = fmt.Errorf("Nodes error: not valid id - %v. %v", line, err)
			return
		}
		for i := 0; i < 3; i++ {
			var v float64
			v, err = strconv.ParseFloat(fs[i+1], 64)
			if err != nil {
				err = fmt.Errorf("Nodes error: not valid coord - %v. %v", line, err)
				return
			}
			coord[i] = v
		}
		msh.Nodes = append(msh.Nodes, Node{Id: id, Coord: coord})
	}

	// Elements
	for _, line := range getLines(lines, "$Elements", "$EndElements") {
		var vs []int
		for _, field := range strings.Fields(line) {
			var id int
			id, err = strconv.Atoi(field)
			if err != nil {
				err = fmt.Errorf("Elements error: %v. %v", line, err)
				return
			}
			vs = append(vs, id)
		}
		msh.Elements = append(msh.Elements, Element{
			Id:     vs[0],
			EType:  ElementType(vs[1]),
			Tags:   vs[3 : 3+vs[2]],
			NodeId: vs[3+vs[2]:],
		})
	}
	return
}

func getLine(lines []string, name string) (line int) {
	for i, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "\r", "")
		if name == line {
			return i
		}
	}
	return -1
}

func getLines(lines []string, from, to string) (res []string) {
	fi := getLine(lines, from)
	ti := getLine(lines, to)
	if fi < 0 || ti < 0 {
		return
	}
	if ti < fi {
		return
	}
	return lines[fi+2 : ti]
}
