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
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/Konstantin8105/gog"
	"github.com/Konstantin8105/reindex"
)

var GmshApp = flag.String("gmsh", "", "Example:\n`gmsh` for Linux\n`gmsh.exe` for Windows")

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
	Coord gog.Point3d
}

type Msh struct {
	PhysicalNames []PhysicalName
	Nodes         []Node
	Elements      []Element
}

func (msh *Msh) MergeNodes(minDistance float64) {
	if minDistance < 0.0 {
		return
	}
	if minDistance == 0.0 {
		minDistance = gog.Eps3D
	}
	for iter := 0; iter < 1000; iter++ {
		found := false
		for i := range msh.Nodes {
			for j := range msh.Nodes {
				if j <= i {
					continue
				}
				d := gog.Distance3d(msh.Nodes[i].Coord, msh.Nodes[j].Coord)
				if minDistance < d {
					continue
				}
				from := msh.Nodes[i].Id
				to := msh.Nodes[j].Id
				found = true
				// change nodes id in elements
				for i := range msh.Elements {
					list := &msh.Elements[i].NodeId
					for k := range *list {
						if (*list)[k] == from {
							(*list)[k] = to
							break
						}
					}
				}
				msh.Nodes = append(msh.Nodes[:i], msh.Nodes[i+1:]...)
				break
			}
		}
		if !found {
			return
		}
	}
	msh.Index1()
}

func (src Msh) Clone() Msh {
	var dst Msh
	// physical names
	dst.PhysicalNames = append([]PhysicalName{}, src.PhysicalNames...)
	// nodes
	dst.Nodes = make([]Node, len(src.Nodes))
	for i := range src.Nodes {
		var n Node
		n.Id = src.Nodes[i].Id
		n.Coord = src.Nodes[i].Coord
		dst.Nodes[i] = n
	}
	// elements
	dst.Elements = make([]Element, len(src.Elements))
	for i := range src.Elements {
		var el Element
		el.Id = src.Elements[i].Id
		el.EType = src.Elements[i].EType
		el.Tags = append([]int{}, src.Elements[i].Tags...)
		el.NodeId = append([]int{}, src.Elements[i].NodeId...)
		dst.Elements[i] = el
	}
	return dst
}

// AddMesh add mesh
func (msh *Msh) AddMsh(source Msh) (err error) {
	src := source.Clone()
	src.Index1()

	// nodes
	var riNodes reindex.List[int]
	{
		counter := msh.MaxNodeIndex() + 1
		for _, n := range src.Nodes {
			riNodes.Add(n.Id, counter)
			counter++
		}
	}
	for i := range src.Nodes {
		id := &src.Nodes[i].Id
		*id, err = riNodes.Get(*id)
		if err != nil {
			return
		}
	}
	msh.Nodes = append(msh.Nodes, src.Nodes...)
	sort.Slice(msh.Nodes, func(i, j int) bool {
		return msh.Nodes[i].Id < msh.Nodes[j].Id
	})
	// elements
	var riElements reindex.List[int]
	{
		counter := msh.MaxElementIndex() + 1
		for _, el := range src.Elements {
			riElements.Add(el.Id, counter)
			counter++
		}
	}
	for i := range src.Elements {
		id := &src.Elements[i].Id
		*id, err = riElements.Get(*id)
		if err != nil {
			return
		}
		for j := range src.Elements[i].NodeId {
			id := &src.Elements[i].NodeId[j]
			*id, err = riNodes.Get(*id)
			if err != nil {
				return
			}
		}
	}
	msh.Elements = append(msh.Elements, src.Elements...)
	sort.Slice(msh.Elements, func(i, j int) bool {
		return msh.Elements[i].Id < msh.Elements[j].Id
	})
	// physical names
	msh.PhysicalNames = append(msh.PhysicalNames, src.PhysicalNames...)
	return
}

func (msh *Msh) MaxNodeIndex() int {
	maxIndex := -math.MaxInt
	found := false
	for i := range msh.Nodes {
		if index := msh.Nodes[i].Id; maxIndex < index {
			found = true
			maxIndex = index
		}
	}
	if !found {
		maxIndex = 0
	}
	return maxIndex
}

func (msh *Msh) MaxElementIndex() int {
	maxIndex := -math.MaxInt
	found := false
	for i := range msh.Elements {
		if index := msh.Elements[i].Id; maxIndex < index {
			found = true
			maxIndex = index
		}
	}
	if !found {
		maxIndex = 0
	}
	return maxIndex
}

func (msh *Msh) Sort(ets ...ElementType) {
	pos := func(et ElementType) int {
		for i := range ets {
			if ets[i] == et {
				return i
			}
		}
		return len(ets)
	}
	sort.Slice(msh.Elements, func(i, j int) bool {
		return pos(msh.Elements[i].EType) < pos(msh.Elements[j].EType)
	})
}

func (msh *Msh) RemoveElements(ets ...ElementType) {
	for i := len(msh.Elements) - 1; 0 <= i; i-- {
		remove := false
		for k := range ets {
			if ets[k] == msh.Elements[i].EType {
				remove = true
			}
		}
		if !remove {
			continue
		}
		msh.Elements = append(msh.Elements[:i], msh.Elements[i+1:]...)
	}
}

func (msh *Msh) Index1() (err error) {
	// reindex nodes
	var riNodes reindex.List[int]
	{
		counter := 1
		for _, n := range msh.Nodes {
			riNodes.Add(n.Id, counter)
			counter++
		}
	}
	for i := range msh.Nodes {
		id := &msh.Nodes[i].Id
		*id, err = riNodes.Get(*id)
		if err != nil {
			return
		}
	}
	for i := range msh.Elements {
		for j := range msh.Elements[i].NodeId {
			id := &msh.Elements[i].NodeId[j]
			*id, err = riNodes.Get(*id)
			if err != nil {
				return
			}
		}
	}
	// reindex elements
	var riElements reindex.List[int]
	{
		counter := 1
		for _, el := range msh.Elements {
			riElements.Add(el.Id, counter)
			counter++
		}
	}
	for i := range msh.Elements {
		id := &msh.Elements[i].Id
		*id, err = riElements.Get(*id)
		if err != nil {
			return
		}
	}
	return
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
	var sb strings.Builder
	sb.WriteString(`$MeshFormat
2.2 0 8
$EndMeshFormat
`)
	if 0 < len(msh.PhysicalNames) {
		sb.WriteString("$PhysicalNames\n")
		sb.WriteString(fmt.Sprintf("%d\n", len(msh.PhysicalNames)))
		for _, pn := range msh.PhysicalNames {
			sb.WriteString(fmt.Sprintf("%v %d \"%s\"\n",
				pn.Dimension, pn.Tag, pn.Name))
		}
		sb.WriteString("$EndPhysicalNames\n")
	}
	if 0 < len(msh.Nodes) {
		sb.WriteString("$Nodes\n")
		sb.WriteString(fmt.Sprintf("%d\n", len(msh.Nodes)))
		for _, n := range msh.Nodes {
			sb.WriteString(fmt.Sprintf("%d %f %f %f\n",
				n.Id, n.Coord[0], n.Coord[1], n.Coord[2]))
		}
		sb.WriteString("$EndNodes\n")
	}
	if 0 < len(msh.Elements) {
		sb.WriteString("$Elements\n")
		sb.WriteString(fmt.Sprintf("%d\n", len(msh.Elements)))
		for _, el := range msh.Elements {
			sb.WriteString(fmt.Sprintf("%d %d ", el.Id, el.EType))
			sb.WriteString(fmt.Sprintf("%d", len(el.Tags)))
			for _, t := range el.Tags {
				sb.WriteString(fmt.Sprintf(" %d", t))
			}
			for _, np := range el.NodeId {
				sb.WriteString(fmt.Sprintf(" %d", np))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("$EndElements\n")
	}
	return sb.String()
}

func New(geoContent string) (m *Msh, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("New: %v", err)
		}
	}()
	msh, err := Generate(geoContent)
	if err != nil {
		return
	}
	return Parse(msh)
}

func DefaultGmsh() {
	if *GmshApp != "" {
		return
	}
	*GmshApp = "gmsh"
	if runtime.GOOS == "windows" {
		*GmshApp = "gmsh.exe"
	}
}

func Generate(geoContent string) (mshContent string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Generate: %v", err)
		}
	}()
	// create temp directory
	var dir string
	dir, err = ioutil.TempDir("", "msh")
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("%v. See dir: %v", err, dir)
		} else {
			os.RemoveAll(dir) // clean up
		}
	}()

	// create geo file
	geofn := filepath.Join(dir, "m.geo")
	if err = ioutil.WriteFile(geofn, []byte(geoContent), 0666); err != nil {
		err = fmt.Errorf("geofn error: %v", err)
		return
	}

	// application
	DefaultGmsh()

	// run gmsh
	meshfn := filepath.Join(dir, "m.msh")
	args := []string{
		*GmshApp,
		"-format", "msh2", // Format: MSH2
		"-smooth", "10", // Smooth mesh
		"-3", // 3D mesh generation
		geofn,
		meshfn,
	}
	var out []byte
	if out, err = exec.Command(args[0], args[1:]...).Output(); err != nil {
		//if !bytes.Contains(out, []byte("Done writing")) {
		err = fmt.Errorf("meshfn error: %v with args: %v\n%s",
			err, args, string(out))
		return
		// }
	}
	// read msh
	meshContent, err := ioutil.ReadFile(meshfn)
	if err != nil {
		err = fmt.Errorf("meshContent error: %v", err)
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
