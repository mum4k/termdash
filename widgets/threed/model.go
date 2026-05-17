// threed/model.go

package threed

// NewModel creates a new empty model.
func NewModel() *Model {
	return &Model{
		Faces: []Face{},
	}
}

// AddFace adds a face to the model. It pre-computes the unit normal for faces
// with at least 3 vertices so the Draw loop can skip the per-frame cross-product.
func (m *Model) AddFace(face Face) {
	if len(face.Vertices) >= 3 && face.Normal == (Vector3D{}) {
		face.Normal = computeFaceNormal(face.Vertices)
	}
	m.Faces = append(m.Faces, face)
}

// SetColor applies the same base color to every face in the model.
func (m *Model) SetColor(color Color) {
	for i := range m.Faces {
		m.Faces[i].Color = color
		m.Faces[i].HasColor = true
	}
}

// Center calculates the center point of the model.
func (m *Model) Center() Vector3D {
	var sumX, sumY, sumZ float64
	var count int
	for _, face := range m.Faces {
		for _, vertex := range face.Vertices {
			sumX += vertex.X
			sumY += vertex.Y
			sumZ += vertex.Z
			count++
		}
	}
	return Vector3D{
		X: sumX / float64(count),
		Y: sumY / float64(count),
		Z: sumZ / float64(count),
	}
}

// Translate moves the entire model by the given offset.
func (m *Model) Translate(offset Vector3D) {
	for fi, face := range m.Faces {
		for vi, vertex := range face.Vertices {
			m.Faces[fi].Vertices[vi] = Vector3D{
				X: vertex.X - offset.X,
				Y: vertex.Y - offset.Y,
				Z: vertex.Z - offset.Z,
			}
		}
	}
}
