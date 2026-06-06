// Copyright 2026 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package threed

import "math"

// CreateCube creates a cube model centered at the specified position.
func CreateCube(position Vector3D, size float64, char rune) *Model {
	halfSize := size / 2
	vertices := []Vector3D{
		{X: position.X - halfSize, Y: position.Y - halfSize, Z: position.Z - halfSize},
		{X: position.X + halfSize, Y: position.Y - halfSize, Z: position.Z - halfSize},
		{X: position.X + halfSize, Y: position.Y + halfSize, Z: position.Z - halfSize},
		{X: position.X - halfSize, Y: position.Y + halfSize, Z: position.Z - halfSize},
		{X: position.X - halfSize, Y: position.Y - halfSize, Z: position.Z + halfSize},
		{X: position.X + halfSize, Y: position.Y - halfSize, Z: position.Z + halfSize},
		{X: position.X + halfSize, Y: position.Y + halfSize, Z: position.Z + halfSize},
		{X: position.X - halfSize, Y: position.Y + halfSize, Z: position.Z + halfSize},
	}

	return facesModel([]Face{
		{Vertices: []Vector3D{vertices[3], vertices[2], vertices[1], vertices[0]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[5], vertices[6], vertices[7]}, Char: char},
		{Vertices: []Vector3D{vertices[7], vertices[3], vertices[0], vertices[4]}, Char: char},
		{Vertices: []Vector3D{vertices[1], vertices[2], vertices[6], vertices[5]}, Char: char},
		{Vertices: []Vector3D{vertices[7], vertices[6], vertices[2], vertices[3]}, Char: char},
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[5], vertices[4]}, Char: char},
	})
}

// CreateTetrahedron creates a tetrahedron model centered at the specified position.
func CreateTetrahedron(position Vector3D, size float64) *Model {
	h := size * math.Sqrt(2.0/3.0)
	vertices := []Vector3D{
		{X: position.X, Y: position.Y + h/2, Z: position.Z},
		{X: position.X - size/2, Y: position.Y - h/2, Z: position.Z - size/(2*math.Sqrt(3))},
		{X: position.X + size/2, Y: position.Y - h/2, Z: position.Z - size/(2*math.Sqrt(3))},
		{X: position.X, Y: position.Y - h/2, Z: position.Z + size/math.Sqrt(3)},
	}

	return facesModel([]Face{
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[2]}, Char: '▲'},
		{Vertices: []Vector3D{vertices[0], vertices[2], vertices[3]}, Char: '▲'},
		{Vertices: []Vector3D{vertices[0], vertices[3], vertices[1]}, Char: '▲'},
		{Vertices: []Vector3D{vertices[1], vertices[2], vertices[3]}, Char: '▲'},
	})
}

// CreatePyramid creates a square pyramid model centered at the specified position.
func CreatePyramid(position Vector3D, size float64, char rune) *Model {
	halfSize := size / 2
	halfHeight := size / 2
	vertices := []Vector3D{
		{X: position.X - halfSize, Y: position.Y - halfHeight, Z: position.Z - halfSize},
		{X: position.X + halfSize, Y: position.Y - halfHeight, Z: position.Z - halfSize},
		{X: position.X + halfSize, Y: position.Y - halfHeight, Z: position.Z + halfSize},
		{X: position.X - halfSize, Y: position.Y - halfHeight, Z: position.Z + halfSize},
		{X: position.X, Y: position.Y + halfHeight, Z: position.Z},
	}

	return facesModel([]Face{
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[2], vertices[3]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[1], vertices[0]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[2], vertices[1]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[3], vertices[2]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[0], vertices[3]}, Char: char},
	})
}

// CreateSphere creates a low-poly sphere model centered at the specified position.
func CreateSphere(position Vector3D, radius float64, latSegments, lonSegments int, char rune) *Model {
	if latSegments < 3 {
		latSegments = 3
	}
	if lonSegments < 4 {
		lonSegments = 4
	}

	point := func(lat, lon int) Vector3D {
		theta := math.Pi * float64(lat) / float64(latSegments)
		phi := 2 * math.Pi * float64(lon) / float64(lonSegments)
		sinTheta := math.Sin(theta)
		return Vector3D{
			X: position.X + radius*sinTheta*math.Cos(phi),
			Y: position.Y + radius*math.Cos(theta),
			Z: position.Z + radius*sinTheta*math.Sin(phi),
		}
	}

	model := NewModel()
	for lat := 0; lat < latSegments; lat++ {
		for lon := 0; lon < lonSegments; lon++ {
			nextLon := (lon + 1) % lonSegments
			topLeft := point(lat, lon)
			topRight := point(lat, nextLon)
			bottomRight := point(lat+1, nextLon)
			bottomLeft := point(lat+1, lon)

			switch lat {
			case 0:
				model.AddFace(Face{Vertices: []Vector3D{topLeft, bottomRight, bottomLeft}, Char: char})
			case latSegments - 1:
				model.AddFace(Face{Vertices: []Vector3D{topLeft, topRight, bottomLeft}, Char: char})
			default:
				model.AddFace(Face{Vertices: []Vector3D{topLeft, topRight, bottomRight, bottomLeft}, Char: char})
			}
		}
	}
	return model
}

// CreateOctahedron creates an octahedron model centered at the specified position.
func CreateOctahedron(position Vector3D, size float64) *Model {
	h := size / math.Sqrt(2)
	vertices := []Vector3D{
		{X: position.X, Y: position.Y + h, Z: position.Z},
		{X: position.X - h, Y: position.Y, Z: position.Z},
		{X: position.X, Y: position.Y, Z: position.Z - h},
		{X: position.X + h, Y: position.Y, Z: position.Z},
		{X: position.X, Y: position.Y, Z: position.Z + h},
		{X: position.X, Y: position.Y - h, Z: position.Z},
	}

	return facesModel([]Face{
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[2]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[0], vertices[2], vertices[3]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[0], vertices[3], vertices[4]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[0], vertices[4], vertices[1]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[2], vertices[1]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[3], vertices[2]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[4], vertices[3]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[1], vertices[4]}, Char: '♦'},
	})
}

// createMarker builds a single marker model at the provided position.
func createMarker(position Vector3D, size float64, char rune) *Model {
	return CreateCube(position, size, char)
}

// createLine converts consecutive coordinates into line faces.
func createLine(coords []Vector3D, char rune) *Model {
	model := NewModel()
	for i := 0; i < len(coords)-1; i++ {
		model.AddFace(Face{
			Vertices: []Vector3D{coords[i], coords[i+1]},
			Char:     char,
		})
	}
	return model
}

// createPolygon converts coordinates into one filled polygon face.
func createPolygon(coords []Vector3D, char rune) *Model {
	model := NewModel()
	if len(coords) < 3 {
		return model
	}
	model.AddFace(Face{Vertices: coords, Char: char})
	return model
}

// facesModel wraps raw faces in a Model and computes their normals.
func facesModel(faces []Face) *Model {
	model := NewModel()
	for _, face := range faces {
		model.AddFace(face)
	}
	return model
}
