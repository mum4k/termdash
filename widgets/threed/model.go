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

// Append adds all faces from the supplied models.
func (m *Model) Append(models ...*Model) {
	for _, model := range models {
		if model == nil {
			continue
		}
		for _, face := range model.Faces {
			m.AddFace(face)
		}
	}
}

// Clone returns a deep copy of the model.
func (m *Model) Clone() *Model {
	if m == nil {
		return nil
	}
	cloned := NewModel()
	for _, face := range m.Faces {
		verts := make([]Vector3D, len(face.Vertices))
		copy(verts, face.Vertices)
		cloned.AddFace(Face{
			Vertices:   verts,
			Char:       face.Char,
			RenderMode: face.RenderMode,
			Color:      face.Color,
			HasColor:   face.HasColor,
			Normal:     face.Normal,
		})
	}
	return cloned
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
	if count == 0 {
		return Vector3D{}
	}
	return Vector3D{
		X: sumX / float64(count),
		Y: sumY / float64(count),
		Z: sumZ / float64(count),
	}
}

// Move moves the entire model by the given delta.
func (m *Model) Move(delta Vector3D) {
	for fi, face := range m.Faces {
		for vi, vertex := range face.Vertices {
			m.Faces[fi].Vertices[vi] = Vector3D{
				X: vertex.X + delta.X,
				Y: vertex.Y + delta.Y,
				Z: vertex.Z + delta.Z,
			}
		}
	}
}

// Scale uniformly scales the model around the origin.
func (m *Model) Scale(factor float64) {
	for fi, face := range m.Faces {
		for vi, vertex := range face.Vertices {
			m.Faces[fi].Vertices[vi] = Vector3D{
				X: vertex.X * factor,
				Y: vertex.Y * factor,
				Z: vertex.Z * factor,
			}
		}
	}
}

// Translate recenters the model by subtracting the supplied offset.
//
// Prefer Move for user-facing movement. Translate is kept for older callers
// and geospatial normalization code that use this subtractive behavior.
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
