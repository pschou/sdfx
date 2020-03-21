package sdf

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
)

//-----------------------------------------------------------------------------

// ThreeMFModel top level structure of a model for output to 3mf
type ThreeMFModel struct {
	XMLName xml.Name `xml:"model"`
	Lang    string   `xml:"xml:lang,attr"`
	Schema  string   `xml:"xmlns,attr"`
	Unit    string   `xml:"unit,attr"`
	// These aren't grouping into the array field
	Resources []ThreeMFObject `xml:"resources>object"`
	Build     []ThreeMFItem   `xml:"build>item"`
}

type ThreeMFObject struct {
	XMLName xml.Name    `xml:"object"`
	ID      string      `xml:"id,attr"`
	Type    string      `xml:"type,attr"`
	Mesh    ThreeMFMesh `xml:"mesh"`
}

type ThreeMFMesh struct {
	Vertices  []ThreeMFVertex   `xml:"vertices"`
	Triangles []ThreeMFTriangle `xml:"triangles"`
}

type ThreeMFVertex struct {
	XMLName xml.Name `xml:"vertex"`
	X       float64  `xml:"x,attr"`
	Y       float64  `xml:"y,attr"`
	Z       float64  `xml:"z,attr"`
}

type ThreeMFTriangle struct {
	XMLName xml.Name `xml:"triangle"`
	V1      int      `xml:"v1,attr"`
	V2      int      `xml:"v2,attr"`
	V3      int      `xml:"v3,attr"`
}

type ThreeMFItem struct {
	XMLName  xml.Name `xml:"item"`
	ObjectID string   `xml:"objectid,attr"`
}

//-----------------------------------------------------------------------------

// Save3MF writes a triangle mesh to an STL file.
func Save3MF(path string, mesh []*Triangle3) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// dedupe the vertices
	vertices := map[V3]int{}
	for _, t := range mesh {
		for _, v := range t.V {
			if _, isDuplicate := vertices[v]; isDuplicate {
				continue
			}
			vertices[v] = len(vertices)
		}
	}
	outputVertices := make([]ThreeMFVertex, len(vertices))
	for v, i := range vertices {
		outputVertices[i] = ThreeMFVertex{X: v.X, Y: v.Y, Z: v.Z}
	}

	// TODO: Make this more memory-efficient while encoding
	outputTriangles := make([]ThreeMFTriangle, len(mesh))
	for i, t := range mesh {
		outputTriangles[i].V1 = vertices[t.V[0]]
		outputTriangles[i].V2 = vertices[t.V[1]]
		outputTriangles[i].V3 = vertices[t.V[2]]
	}

	buf := bufio.NewWriter(file)
	fmt.Fprintln(buf, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	err = xml.NewEncoder(buf).Encode(ThreeMFModel{
		Lang:   "en-US",
		Schema: "http://schemas.microsoft.com/3dmanufacturing/core/2015/02",
		Unit:   "mm",
		Resources: []ThreeMFObject{
			{ID: "1", Type: "model", Mesh: ThreeMFMesh{
				Vertices:  outputVertices,
				Triangles: outputTriangles,
			}},
		},
		Build: []ThreeMFItem{
			{ObjectID: "1"},
		},
	})
	if err != nil {
		return err
	}
	return buf.Flush()
}

//-----------------------------------------------------------------------------
