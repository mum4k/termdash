// threed/types.go

package threed

// FaceRenderMode controls how a face character is drawn.
type FaceRenderMode int

const (
	// FaceRenderFill paints the entire polygon using the face character.
	FaceRenderFill FaceRenderMode = iota
	// FaceRenderGlyph draws the face character once at the projected face center.
	FaceRenderGlyph
)

// Vector3D represents a point or vector in 3D space.
type Vector3D struct {
	X float64 // X coordinate
	Y float64 // Y coordinate
	Z float64 // Z coordinate
}

// Vector2D represents a point or vector in 2D space.
type Vector2D struct {
	X float64 // X coordinate
	Y float64 // Y coordinate
}

// Face represents a polygon face made up of vertices.
type Face struct {
	Vertices   []Vector3D     // Vertices of the face
	Char       rune           // Character to render for this face
	RenderMode FaceRenderMode // How the face character should be drawn
	Color      Color          // Optional base color for this face
	HasColor   bool           // Whether Color should override the widget diffuse color
	Normal     Vector3D       // Pre-computed unit normal in model space; set by Model.AddFace.
}

// Model represents a 3D model composed of faces.
type Model struct {
	Faces []Face // Faces of the model
}
