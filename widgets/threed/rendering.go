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
	Depths     []float64  // Per-point transformed Z values used for depth tests
	Normal     Vector3D   // Normal vector of the face
	Brightness float64    // Brightness for shading
	Depth      float64    // Average depth for sorting
	Char       rune       // Character to render
	RenderMode FaceRenderMode
	Color      Color // Optional base color for shading
	HasColor   bool  // Whether Color should override the widget diffuse color
	ShadeColor Color // Final shaded color used for rendering
}

// computeFaceNormal returns the outward unit normal of a face from its first
// three vertices using the cross-product of two edges.
func computeFaceNormal(vertices []Vector3D) Vector3D {
	if len(vertices) < 3 {
		return Vector3D{}
	}
	edge1 := vertices[1].Subtract(vertices[0])
	edge2 := vertices[2].Subtract(vertices[0])
	return edge1.Cross(edge2).Normalize()
}

// insertionSortThreshold is the maximum face count for which the inline
// insertion sort is used instead of sort.SliceStable.  Insertion sort is
// stable, allocation-free, and faster than the standard library sort for
// small slices.  Most models have far fewer than 32 faces, so this path is
// hit almost exclusively in practice.
const insertionSortThreshold = 32

// sortFacesByDepth sorts faces back-to-front (farthest first) for the
// Painter's Algorithm.  Stability is required so coplanar faces don't
// flicker between frames due to a non-deterministic sort order.
func sortFacesByDepth(faces []ProjectedFace) {
	n := len(faces)
	if n < 2 {
		return
	}
	if n <= insertionSortThreshold {
		// Insertion sort: O(n²) but stable, allocation-free, and cache-
		// friendly for the small N that characterises typical 3-D models.
		for i := 1; i < n; i++ {
			key := faces[i]
			j := i - 1
			for j >= 0 && faces[j].Depth < key.Depth {
				faces[j+1] = faces[j]
				j--
			}
			faces[j+1] = key
		}
		return
	}
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].Depth > faces[j].Depth
	})
}

// calculatePhongShading computes the shade color and draw character for a face.
// Returns (brightness [0,1], shaded color, draw rune).
//
// Preconditions: normal and lightDir must already be unit vectors — the
// caller (Draw) guarantees this, so we skip the redundant Normalize here.
//
// When preferredChar is non-zero it is preserved as the face character,
// letting callers render symbol-driven models directly. A zero preferredChar
// triggers the standard block-character brightness ramp.
func calculatePhongShading(normal, lightDir Vector3D, baseColor Color, options *Options, preferredChar rune) (float64, Color, rune) {
	// --- Ambient ---
	// AmbientColor's per-channel values encode intensity directly.
	ambientColor := baseColor.Modulate(options.AmbientColor)
	ambientLum := (ambientColor.R + ambientColor.G + ambientColor.B) / 3.0

	// --- Diffuse (Lambertian) ---
	diffuseIntensity := clampFloat(normal.Dot(lightDir), 0.0, 1.0)
	diffuseColor := baseColor.Modulate(options.DiffuseColor).Multiply(diffuseIntensity)

	// --- Specular (Phong) ---
	// Camera sits at z = -Zoom, so the surface-to-camera vector is -Z.
	viewDir := Vector3D{X: 0, Y: 0, Z: -1}
	reflectDir := normal.Multiply(2 * normal.Dot(lightDir)).Subtract(lightDir)
	spec := math.Pow(clampFloat(viewDir.Dot(reflectDir.Normalize()), 0.0, 1.0), options.Shininess)
	specularColor := options.SpecularColor.Multiply(spec)

	// --- Final color ---
	finalColor := ambientColor.Add(diffuseColor).Add(specularColor)

	// --- Brightness for character selection ---
	// Ambient sets the floor; diffuse drives most of the gradient;
	// specular contributes a small highlight.
	brightness := clampFloat(ambientLum+diffuseIntensity*0.8+spec*0.2, 0.0, 1.0)
	char := preferredChar
	if char == 0 {
		char = brightnessToChar(brightness)
	}

	return brightness, finalColor, char
}

// brightnessChars is the block-character ramp used to map brightness to a glyph.
// Declared at package level to avoid allocating the slice on every call.
var brightnessChars = [...]rune{' ', '░', '▒', '▓', '█'}

// brightnessToChar maps a brightness value in [0,1] to a Unicode block
// character, darkest to brightest.
func brightnessToChar(brightness float64) rune {
	b := clampFloat(brightness, 0.0, 1.0)
	index := int(b*float64(len(brightnessChars)-1) + 0.5) // round to nearest
	if index < 0 {
		index = 0
	}
	if index >= len(brightnessChars) {
		index = len(brightnessChars) - 1
	}
	return brightnessChars[index]
}

// drawFillFace renders a non-glyph face directly onto the destination canvas.
// High-density fill faces go through subcellScene instead; this path handles
// outline/overlay modes only.
func drawFillFace(cvs *canvas.Canvas, points []Vector2D, clr Color, char rune, mode FaceRenderMode) {
	if len(points) < 3 {
		return
	}
	if mode == FaceRenderGlyph {
		drawGlyphFace(cvs, points, clr.ToCellColor(), char)
		return
	}

	cellColor := clr.ToCellColor()

	// Build the cell options once outside the loop — block chars need both
	// fg and bg set for solid fill; symbol chars use fg only.
	// A fixed-size array keeps the backing storage on the stack and avoids a
	// heap allocation for the slice literal on every call.
	var optBuf [2]cell.Option
	optBuf[0] = cell.FgColor(cellColor)
	var opts []cell.Option
	if shouldFillFaceBackground(char) {
		optBuf[1] = cell.BgColor(cellColor)
		opts = optBuf[:]
	} else {
		opts = optBuf[:1]
	}

	// Bounding box of the projected polygon.
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

	// Cache canvas bounds so we don't call Area() on every pixel.
	cvsW := cvs.Area().Dx()
	cvsH := cvs.Area().Dy()

	for x := int(math.Round(minX)); x <= int(math.Round(maxX)); x++ {
		if x < 0 || x >= cvsW {
			continue
		}
		for y := int(math.Round(minY)); y <= int(math.Round(maxY)); y++ {
			if y < 0 || y >= cvsH {
				continue
			}
			if pointInPolygon(float64(x)+0.5, float64(y)+0.5, points) {
				_, _ = cvs.SetCell(image.Point{X: x, Y: y}, char, opts...)
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

// polygonCenter returns the centroid (average vertex position) of a polygon.
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

// shouldFillFaceBackground reports whether char is a block-shading rune that
// needs both foreground and background colored for a solid fill. Symbol-driven
// models use foreground only so internal glyph detail stays visible.
func shouldFillFaceBackground(char rune) bool {
	switch char {
	case ' ', '░', '▒', '▓', '█':
		return true
	default:
		return false
	}
}

// pointInPolygon reports whether (x, y) lies inside polygon using the
// ray-casting (even-odd) algorithm.
func pointInPolygon(x, y float64, polygon []Vector2D) bool {
	n := len(polygon)
	inside := false
	for i := range polygon {
		j := (i + n - 1) % n
		xi, yi := polygon[i].X, polygon[i].Y
		xj, yj := polygon[j].X, polygon[j].Y
		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
	}
	return inside
}

// drawLine draws a line between p1 and p2 using Bresenham's algorithm.
func drawLine(cvs *canvas.Canvas, p1, p2 Vector2D, char rune, opts ...cell.Option) {
	drawLineFiltered(cvs, p1, p2, char, nil, opts...)
}

// drawLineFiltered draws a line between p1 and p2, skipping cells rejected by
// shouldSkip.
func drawLineFiltered(cvs *canvas.Canvas, p1, p2 Vector2D, char rune, shouldSkip func(image.Point) bool, opts ...cell.Option) {
	x1, y1 := int(math.Round(p1.X)), int(math.Round(p1.Y))
	x2, y2 := int(math.Round(p2.X)), int(math.Round(p2.Y))

	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))

	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}

	cvsW := cvs.Area().Dx()
	cvsH := cvs.Area().Dy()
	e := dx - dy

	for {
		if x1 >= 0 && x1 < cvsW && y1 >= 0 && y1 < cvsH {
			p := image.Point{X: x1, Y: y1}
			if shouldSkip == nil || !shouldSkip(p) {
				_, _ = cvs.SetCell(p, char, opts...)
			}
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * e
		if e2 > -dy {
			e -= dy
			x1 += sx
		}
		if e2 < dx {
			e += dx
			y1 += sy
		}
	}
}

// drawAxes draws X (red) and Y (green) reference axes through the canvas centre.
func drawAxes(cvs *canvas.Canvas) {
	width := cvs.Area().Dx()
	height := cvs.Area().Dy()
	centerX := width / 2
	centerY := height / 2

	for x := 0; x < width; x++ {
		_, _ = cvs.SetCell(image.Point{X: x, Y: centerY}, '-', cell.FgColor(cell.ColorRed))
	}
	for y := 0; y < height; y++ {
		_, _ = cvs.SetCell(image.Point{X: centerX, Y: y}, '|', cell.FgColor(cell.ColorGreen))
	}
}
