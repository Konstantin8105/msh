package msh

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Konstantin8105/errors"
	"github.com/Konstantin8105/pow"
)

const (
	meshFormat    = "$MeshFormat"
	meshFormatEnd = "$EndMeshFormat"
	nodes         = "$Nodes"
	nodesEnd      = "$EndNodes"
	elements      = "$Elements"
	elementsEnd   = "$EndElements"
)

type elementType int

const (
	point       elementType = 15
	line                    = 1
	triangle                = 2
	quadrangle              = 3
	tetrahedron             = 4
)

type Point struct {
	Id      int
	X, Y, Z float64
}

type Line struct {
	Id       int
	PointsId [2]int
}

type Triangle struct {
	Id       int
	PointsId [3]int
}

type Msh struct {
	Points    []Point
	Lines     []Line
	Triangles []Triangle
}

func (m Msh) PointsById(pIds [3]int) (ps [3]Point) {
	for index, id := range pIds {
		for j := range m.Points {
			if id == m.Points[j].Id {
				ps[index] = m.Points[j]
				break
			}
		}
	}
	return
}

func (m *Msh) RotateXOY90deg() {
	for i := range m.Points {
		m.Points[i].X, m.Points[i].Y = m.Points[i].Y, m.Points[i].X // swap
	}
}

func (m *Msh) RotateXOY(a float64) {
	for i := range m.Points {
		x, y := m.Points[i].X, m.Points[i].Y
		ampl := math.Sqrt(pow.E2(x) + pow.E2(y))
		angle := math.Atan2(y, x) + a
		m.Points[i].X = ampl * math.Sin(angle)
		m.Points[i].Y = ampl * math.Cos(angle)
	}
}

func (m *Msh) MoveXOY(x, y float64) {
	for i := range m.Points {
		m.Points[i].X += x
		m.Points[i].Y += y
	}
}

func New(geoContent string) (m *Msh, err error) {
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
	if err = exec.Command("gmsh", "-2", geofn, meshfn).Run(); err != nil {
		return
	}

	// read msh
	meshContent, err := ioutil.ReadFile(meshfn)
	if err != nil {
		return
	}

	// create mesh
	var msh Msh

	// Example of meshContent:
	//
	// $MeshFormat
	// 2.2 0 8
	// $EndMeshFormat
	// $Nodes
	// 5
	// 1 0 0 0
	// 2 5 0 0
	// 3 0 10 0
	// 4 5 10 0
	// 5 2.5 5 0
	// $EndNodes
	// $Elements
	// 12
	// 1 15 2 0 0 1
	// 2 15 2 0 1 2
	// 3 15 2 0 2 3
	// 4 15 2 0 3 4
	// 5 1 2 0 1 2 4
	// 6 1 2 0 2 1 3
	// 7 1 2 0 3 1 2
	// 8 1 2 0 4 3 4
	// 9 2 2 0 6 1 2 5
	// 10 2 2 0 6 3 5 4
	// 11 2 2 0 6 1 5 3
	// 12 2 2 0 6 2 4 5
	// $EndElements

	lines := strings.Split(string(meshContent), "\n")

	var nodeLines []string
	{
		var start, end int
		for i := range lines {
			switch strings.TrimSpace(lines[i]) {
			case nodes:
				start = i
			case nodesEnd:
				end = i
			}
		}
		nodeLines = lines[start+1+1 : end]
	}

	var eleLines []string
	{
		var start, end int
		for i := range lines {
			switch strings.TrimSpace(lines[i]) {
			case elements:
				start = i
			case elementsEnd:
				end = i
			}
		}
		eleLines = lines[start+1+1 : end]
	}

	// parse nodes
	if err = msh.parseNodes(nodeLines); err != nil {
		return
	}

	// parse elements
	if err = msh.parseElements(eleLines); err != nil {
		return
	}

	return &msh, err
}

// Example:
//
// 3 0 10 0
// 6 1 2 0 2 1 3
// 7 1 2 0 3 1 2
// 8 1 2 0 4 3 4
// 4 5 10 0
// 5 2.5 5 0
func (msh *Msh) parseNodes(nodeLines []string) error {
	et := errors.New("parse nodes")
	for _, nline := range nodeLines {
		eLine := errors.New(nline)
		fields := strings.Fields(nline)
		if size := len(fields); size != 4 {
			eLine.Add(fmt.Errorf("Amount of node fileds %d is node valid", size))
			continue
		}
		// parsing values
		var (
			id      int
			x, y, z float64
			err     error
		)
		if id, err = strconv.Atoi(fields[0]); err != nil {
			eLine.Add(err)
		}
		if x, err = strconv.ParseFloat(fields[1], 64); err != nil {
			eLine.Add(err)
		}
		if y, err = strconv.ParseFloat(fields[2], 64); err != nil {
			eLine.Add(err)
		}
		if z, err = strconv.ParseFloat(fields[3], 64); err != nil {
			eLine.Add(err)
		}
		if eLine.IsError() {
			et.Add(eLine)
			continue
		}
		msh.Points = append(msh.Points, Point{Id: id, X: x, Y: y, Z: z})
	}
	if et.IsError() {
		return et
	}
	return nil
}

// Example
//
// 4 15 2 0 3 4
// 5 1  2 0 1 2 4
// 9 2  2 0 6 1 2 5
func (msh *Msh) parseElements(eleLines []string) error {
	et := errors.New("parse elements")
	for _, eline := range eleLines {
		eLine := errors.New(eline)
		fields := strings.Fields(eline)
		if size := len(fields); size == 0 {
			eLine.Add(fmt.Errorf("Zero amount of node fileds %d is node valid", size))
			continue
		}
		// parsing values
		var (
			id      int
			eleType int
			err     error
		)
		if id, err = strconv.Atoi(fields[0]); err != nil {
			eLine.Add(err)
		}
		if eleType, err = strconv.Atoi(fields[1]); err != nil {
			eLine.Add(err)
		}

		switch elementType(eleType) {
		case line:
			var e Line
			for i := 0; i < len(e.PointsId); i++ {
				var pid int
				if pid, err = strconv.Atoi(fields[i+len(fields)-len(e.PointsId)]); err != nil {
					eLine.Add(err)
				}
				e.PointsId[i] = pid
			}
			e.Id = id
			msh.Lines = append(msh.Lines, e)
		case triangle:
			var e Triangle
			for i := 0; i < len(e.PointsId); i++ {
				var pid int
				if pid, err = strconv.Atoi(fields[i+len(fields)-len(e.PointsId)]); err != nil {
					eLine.Add(err)
				}
				e.PointsId[i] = pid
			}
			e.Id = id
			msh.Triangles = append(msh.Triangles, e)
		}
	}
	if et.IsError() {
		return et
	}
	return nil
}
