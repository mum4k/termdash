// threed/rendering.go

package threed

import (
	"image"
	"math"
	"sort"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
)

// ProjectedFace represents a face projected onto 2D space.
type ProjectedFace struct {
	Points     []Vector2D // Projected 2D points
	Normal     Vector3D   // Normal vector of the face
	Brightness float64    // Brightness for shading
	Depth      float64    // Average depth for sorting
	Char       rune       // Character to render
	RenderMode FaceRenderMode
	Color      Color // Optional base color for shading
	HasColor   bool  // Whether Color should override the widget diffuse color
	ShadeColor Color // Final shaded color used for rendering
}

// computeFaceNormal computes the outward unit normal of a face from its
// first three vertices using the cross-product of two edges.
func computeFaceNormal(vertices []Vector3D) Vector3D {
	if len(vertices) < 3 {
		return Vector3D{}
	}
	edge1 := vertices[1].Subtract(vertices[0])
	edge2 := vertices[2].Subtract(vertices[0])
	return edge1.Cross(edge2).Normalize()
}

// sortFacesByDepth sorts faces from farthest to nearest (back-to-front)
// for correct Painter's Algorithm rendering.
func sortFacesByDepth(faces []ProjectedFace) {
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].Depth > faces[j].Depth
	})
}

// calculatePhongShading calculates the color and character for a face using
// Phong shading. Returns (brightness [0,1], cell color, draw character).
//
// When preferredChar is non-zero it is preserved as the face character. This
// lets callers render symbol-driven models, while callers that leave it unset
// still get the usual block-character brightness ramp.
func calculatePhongShading(normal Vector3D, lightDir Vector3D, baseColor Color, options *Options, preferredChar rune) (float64, Color, rune) {
	normal = normal.Normalize()
	lightDir = lightDir.Normalize()

	// --- Ambient ---
	// Use AmbientColor directly; its per-channel values encode intensity.
	ambientColor := baseColor.Modulate(options.AmbientColor)
	ambientLum := (ambientColor.R + ambientColor.G + ambientColor.B) / 3.0

	// --- Diffuse (Lambertian) ---
	diffuseIntensity := clampFloat(normal.Dot(lightDir), 0.0, 1.0)
	diffuseColor := baseColor.Modulate(options.DiffuseColor).Multiply(diffuseIntensity)

	// --- Specular (Phong) ---
	// viewDir: from surface toward camera.  Camera is at z = -Zoom (negative Z),
	// so the surface-to-camera vector points in the -Z direction.
	viewDir := Vector3D{X: 0, Y: 0, Z: -1}
	reflectDir := normal.Multiply(2 * normal.Dot(lightDir)).Subtract(lightDir)
	spec := math.Pow(clampFloat(viewDir.Dot(reflectDir.Normalize()), 0.0, 1.0), options.Shininess)
	specularColor := options.SpecularColor.Multiply(spec)

	// --- Final color ---
	finalColor := ambientColor.Add(diffuseColor).Add(specularColor)

	// --- Brightness for character selection ---
	// Keep in [0,1]: ambient provides the floor, diffuse drives most of the
	// gradient, specular adds a small highlight contribution.
	brightness := clampFloat(ambientLum+diffuseIntensity*0.8+spec*0.2, 0.0, 1.0)
	char := preferredChar
	if char == 0 {
		char = brightnessToChar(brightness)
	}

	return brightness, finalColor, char
}

// brightnessToChar maps a brightness value in [0,1] to a Unicode block
// character, darkest to brightest.
func brightnessToChar(brightness float64) rune {
	chars := []rune{' ', '░', '▒', '▓', '█'}
	// Clamp defensively so callers need not worry about range.
	b := clampFloat(brightness, 0.0, 1.0)
	index := int(b*float64(len(chars)-1) + 0.5) // round to nearest
	if index < 0 {
		index = 0
	}
	if index >= len(chars) {
		index = len(chars) - 1
	}
	return chars[index]
}

// drawFillFace renders a non-glyph face directly onto the destination canvas.
// High-density fill faces are expected to go through subcellScene instead.
func drawFillFace(cvs *canvas.Canvas, points []Vector2D, clr Color, char rune, mode FaceRenderMode) {
	if len(points) < 3 {
		return
	}
	if mode == FaceRenderGlyph {
		drawGlyphFace(cvs, points, clr.ToCellColor(), char)
		return
	}

	cellColor := clr.ToCellColor()

	// Find the bounding box of the polygon
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	useBackground := shouldFillFaceBackground(char)

	// Iterate over the bounding box and fill pixels inside the polygon
	for x := int(math.Round(minX)); x <= int(math.Round(maxX)); x++ {
		for y := int(math.Round(minY)); y <= int(math.Round(maxY)); y++ {
			if pointInPolygon(float64(x)+0.5, float64(y)+0.5, points) {
				p := image.Point{X: x, Y: y}
				if x >= 0 && x < cvs.Area().Dx() && y >= 0 && y < cvs.Area().Dy() {
					opts := []cell.Option{cell.FgColor(cellColor)}
					if useBackground {
						opts = append(opts, cell.BgColor(cellColor))
					}
					_, _ = cvs.SetCell(p, char, opts...)
				}
			}
		}
	}
}

// drawGlyphFace renders one glyph at the center of the projected polygon.
func drawGlyphFace(cvs *canvas.Canvas, points []Vector2D, clr cell.Color, char rune) {
	center := polygonCenter(points)
	p := image.Point{
		X: int(math.Round(center.X)),
		Y: int(math.Round(center.Y)),
	}
	if p.X < 0 || p.X >= cvs.Area().Dx() || p.Y < 0 || p.Y >= cvs.Area().Dy() {
		return
	}
	_, _ = cvs.SetCell(p, char, cell.FgColor(clr))
}

// polygonCenter returns the average position of a polygon's vertices.
func polygonCenter(points []Vector2D) Vector2D {
	var center Vector2D
	for _, point := range points {
		center.X += point.X
		center.Y += point.Y
	}
	scale := 1 / float64(len(points))
	center.X *= scale
	center.Y *= scale
	return center
}

// shouldFillFaceBackground reports whether a face character should flood the
// background color as well as the foreground.
//
// Symbol-driven models such as repeated UTF-8 glyphs and emoji look much more
// faithful when only the foreground is colored, while block-shaded meshes use
// matching foreground/background colors for solid fill.
func shouldFillFaceBackground(char rune) bool {
	switch char {
	case ' ', '░', '▒', '▓', '█':
		return true
	default:
		return false
	}
}

// pointInPolygon checks if a point is inside a polygon using the ray-casting algorithm.
func pointInPolygon(x, y float64, polygon []Vector2D) bool {
	n := len(polygon)
	inside := false
	for i := range polygon {
		j := (i + n - 1) % n
		xi, yi := polygon[i].X, polygon[i].Y
		xj, yj := polygon[j].X, polygon[j].Y
		intersect := ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
	}
	return inside
}

func drawLine(cvs *canvas.Canvas, p1, p2 Vector2D, char rune, opts ...cell.Option) {
	x1, y1 := int(math.Round(p1.X)), int(math.Round(p1.Y))
	x2, y2 := int(math.Round(p2.X)), int(math.Round(p2.Y))

	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))

	var sx, sy int
	if x1 < x2 {
		sx = 1
	} else {
		sx = -1
	}
	if y1 < y2 {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy

	for {
		point := image.Point{X: x1, Y: y1}

		// Set the cell if within canvas bounds
		if x1 >= 0 && x1 < cvs.Area().Dx() && y1 >= 0 && y1 < cvs.Area().Dy() {
			cvs.SetCell(point, char, opts...)
		}

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// drawAxes draws the coordinate axes on the canvas.
func drawAxes(cvs *canvas.Canvas) {
	width := cvs.Area().Dx()
	height := cvs.Area().Dy()
	centerX := width / 2
	centerY := height / 2

	// X-axis
	for x := 0; x < width; x++ {
		p := image.Point{X: x, Y: centerY}
		_, err := cvs.SetCell(p, '-', cell.FgColor(cell.ColorRed))
		if err != nil {
			// Handle error if needed.
		}
	}

	// Y-axis
	for y := 0; y < height; y++ {
		p := image.Point{X: centerX, Y: y}
		_, err := cvs.SetCell(p, '|', cell.FgColor(cell.ColorGreen))
		if err != nil {
			// Handle error if needed.
		}
	}
}

// createMarker creates a small cube to represent a point.
func createMarker(position Vector3D, size float64, char rune) *Model {
	return CreateCube(position, size, char)
}

// createLine creates a line model from a list of coordinates.
func createLine(coords []Vector3D, char rune) *Model {
	model := NewModel()
	for i := 0; i < len(coords)-1; i++ {
		face := Face{
			Vertices: []Vector3D{coords[i], coords[i+1]},
			Char:     char,
		}
		model.AddFace(face)
	}
	return model
}

// createPolygon creates a polygon model from a list of coordinates.
func createPolygon(coords []Vector3D, char rune) *Model {
	model := NewModel()
	if len(coords) < 3 {
		return model // Not enough points to form a polygon
	}
	face := Face{
		Vertices: coords,
		Char:     char,
	}
	model.AddFace(face)
	return model
}

// CreateCube creates a cube model centered at the specified position.
func CreateCube(position Vector3D, size float64, char rune) *Model {
	halfSize := size / 2

	// Define vertices centered around position.X, position.Y, position.Z
	vertices := []Vector3D{
		{position.X - halfSize, position.Y - halfSize, position.Z - halfSize}, // 0
		{position.X + halfSize, position.Y - halfSize, position.Z - halfSize}, // 1
		{position.X + halfSize, position.Y + halfSize, position.Z - halfSize}, // 2
		{position.X - halfSize, position.Y + halfSize, position.Z - halfSize}, // 3
		{position.X - halfSize, position.Y - halfSize, position.Z + halfSize}, // 4
		{position.X + halfSize, position.Y - halfSize, position.Z + halfSize}, // 5
		{position.X + halfSize, position.Y + halfSize, position.Z + halfSize}, // 6
		{position.X - halfSize, position.Y + halfSize, position.Z + halfSize}, // 7
	}

	// Define faces with outward-pointing normals (counter-clockwise when viewed from outside)
	faces := []Face{
		// Front face (Z-negative side, normal points -Z)
		{Vertices: []Vector3D{vertices[3], vertices[2], vertices[1], vertices[0]}, Char: char},
		// Back face (Z-positive side, normal points +Z)
		{Vertices: []Vector3D{vertices[4], vertices[5], vertices[6], vertices[7]}, Char: char},
		// Left face (X-negative side, normal points -X)
		{Vertices: []Vector3D{vertices[7], vertices[3], vertices[0], vertices[4]}, Char: char},
		// Right face (X-positive side, normal points +X)
		{Vertices: []Vector3D{vertices[1], vertices[2], vertices[6], vertices[5]}, Char: char},
		// Top face (Y-positive side, normal points +Y)
		{Vertices: []Vector3D{vertices[7], vertices[6], vertices[2], vertices[3]}, Char: char},
		// Bottom face (Y-negative side, normal points -Y)
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[5], vertices[4]}, Char: char},
	}

	model := NewModel()
	model.Faces = faces

	return model
}

// CreateTetrahedron creates a tetrahedron model centered at the specified position.
func CreateTetrahedron(position Vector3D, size float64) *Model {
	h := size * math.Sqrt(2.0/3.0) // Height of a regular tetrahedron

	// Define vertices.
	vertices := []Vector3D{
		{position.X, position.Y + h/2, position.Z},                                  // Top vertex
		{position.X - size/2, position.Y - h/2, position.Z - size/(2*math.Sqrt(3))}, // Base vertex 1
		{position.X + size/2, position.Y - h/2, position.Z - size/(2*math.Sqrt(3))}, // Base vertex 2
		{position.X, position.Y - h/2, position.Z + size/math.Sqrt(3)},              // Base vertex 3
	}

	// Define faces.
	faces := []Face{
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[2]}, Char: '▲'},
		{Vertices: []Vector3D{vertices[0], vertices[2], vertices[3]}, Char: '▲'},
		{Vertices: []Vector3D{vertices[0], vertices[3], vertices[1]}, Char: '▲'},
		{Vertices: []Vector3D{vertices[1], vertices[2], vertices[3]}, Char: '▲'},
	}

	model := NewModel()
	model.Faces = faces

	return model
}

// CreatePyramid creates a square pyramid model centered at the specified position.
func CreatePyramid(position Vector3D, size float64, char rune) *Model {
	halfSize := size / 2
	halfHeight := size / 2

	vertices := []Vector3D{
		{position.X - halfSize, position.Y - halfHeight, position.Z - halfSize}, // 0: front-left base
		{position.X + halfSize, position.Y - halfHeight, position.Z - halfSize}, // 1: front-right base
		{position.X + halfSize, position.Y - halfHeight, position.Z + halfSize}, // 2: back-right base
		{position.X - halfSize, position.Y - halfHeight, position.Z + halfSize}, // 3: back-left base
		{position.X, position.Y + halfHeight, position.Z},                       // 4: apex
	}

	faces := []Face{
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[2], vertices[3]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[1], vertices[0]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[2], vertices[1]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[3], vertices[2]}, Char: char},
		{Vertices: []Vector3D{vertices[4], vertices[0], vertices[3]}, Char: char},
	}

	model := NewModel()
	model.Faces = faces
	return model
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

			if lat == 0 {
				model.AddFace(Face{Vertices: []Vector3D{topLeft, bottomRight, bottomLeft}, Char: char})
				continue
			}
			if lat == latSegments-1 {
				model.AddFace(Face{Vertices: []Vector3D{topLeft, topRight, bottomLeft}, Char: char})
				continue
			}
			model.AddFace(Face{Vertices: []Vector3D{topLeft, topRight, bottomRight, bottomLeft}, Char: char})
		}
	}

	return model
}

// CreateOctahedron creates an octahedron model centered at the specified position.
func CreateOctahedron(position Vector3D, size float64) *Model {
	h := size / math.Sqrt(2) // Height from center to vertex

	// Define vertices.
	vertices := []Vector3D{
		{position.X, position.Y + h, position.Z}, // Top vertex
		{position.X - h, position.Y, position.Z}, // Left vertex
		{position.X, position.Y, position.Z - h}, // Front vertex
		{position.X + h, position.Y, position.Z}, // Right vertex
		{position.X, position.Y, position.Z + h}, // Back vertex
		{position.X, position.Y - h, position.Z}, // Bottom vertex
	}

	// Define faces.
	faces := []Face{
		{Vertices: []Vector3D{vertices[0], vertices[1], vertices[2]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[0], vertices[2], vertices[3]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[0], vertices[3], vertices[4]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[0], vertices[4], vertices[1]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[2], vertices[1]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[3], vertices[2]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[4], vertices[3]}, Char: '♦'},
		{Vertices: []Vector3D{vertices[5], vertices[1], vertices[4]}, Char: '♦'},
	}

	model := NewModel()
	model.Faces = faces

	return model
}

// GenerateLineChartModel generates a 3D line chart model from data.
func GenerateLineChartModel(data []float64) *Model {
	model := NewModel()
	numPoints := len(data)
	vertices := make([]Vector3D, numPoints)
	for i, value := range data {
		x := float64(i) - float64(numPoints)/2
		y := value
		vertices[i] = Vector3D{X: x, Y: y, Z: 0}
	}
	// Create faces between consecutive points.
	for i := 0; i < numPoints-1; i++ {
		face := Face{
			Vertices: []Vector3D{vertices[i], vertices[i+1]},
			Char:     '-', // Use '-' or any other desired rune
		}
		model.AddFace(face)
	}
	return model
}

// GenerateBarChartModel generates a 3D bar chart model from data.
func GenerateBarChartModel(data []float64) *Model {
	model := NewModel()
	numBars := len(data)
	for i, value := range data {
		x := float64(i) - float64(numBars)/2
		height := value

		// Create a cube representing the bar.
		// Passing '█' as the character to render filled bars.
		bar := CreateCube(Vector3D{X: x, Y: 0, Z: 0}, height, '█') // Using '█' to render filled bars
		model.Faces = append(model.Faces, bar.Faces...)
	}
	return model
}
