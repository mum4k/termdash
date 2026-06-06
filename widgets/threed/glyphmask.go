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

import (
	"image"
	"math"
	"strings"
)

const (
	// defaultSymbolMaskResolution is the pixel resolution used when rendering
	// image masks for braille previews. High resolution preserves detail in the
	// 2-D braille preview.
	defaultSymbolMaskResolution = 256

	// defaultSymbolModelResolution is the pixel resolution used when building
	// 3-D extruded image models. At 40 each mask cell projects to roughly
	// 1.0–1.2 terminal columns at the default zoom, giving fine detail while
	// staying at or above the minimum braille subcell size.
	defaultSymbolModelResolution = 40

	// defaultSymbolDepthScale is the half-depth expressed as a fraction of the
	// total model extent (defaultSymbolExtent).  Using an extent-based value
	// keeps the shape visually thick regardless of how many mask pixels there
	// are.
	defaultSymbolDepthScale = 0.08

	defaultSymbolExtent = 2.6
)

// glyphMask is a compact binary raster used to build extruded symbol meshes.
type glyphMask struct {
	Width  int
	Height int
	Filled []bool
	Colors []Color
}

// At reports whether the given raster cell is filled.
func (m glyphMask) At(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}
	return m.Filled[y*m.Width+x]
}

// ColorAt returns the sampled color for the given raster cell.
func (m glyphMask) ColorAt(x, y int) Color {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return Color{}
	}
	if len(m.Colors) != len(m.Filled) {
		return Color{}
	}
	return m.Colors[y*m.Width+x]
}

// buildGlyphMaskModel converts a filled binary mask into an extruded mesh.
//
// Each filled cell becomes its own front and back face, preserving the
// per-pixel color sampled from the source image rather than averaging detailed
// artwork into a uniform face color.
//
// The side color is fixed (rather than varying with a step parameter) because
// the per-frame variation is imperceptible (≤3% per channel) and fixity allows
// callers to cache the resulting model across frames.
func buildGlyphMaskModel(mask glyphMask) *Model {
	sideColor := Color{R: 0.46, G: 0.68, B: 0.94}
	model := NewModel()

	cellSize := defaultSymbolExtent / math.Max(float64(mask.Width), float64(mask.Height))
	// halfDepth is a fixed fraction of the total model extent so the shape
	// looks visually thick (comparable to the braille-disc fallback) regardless
	// of how many mask pixels are used.
	halfDepth := defaultSymbolExtent * defaultSymbolDepthScale
	left := -float64(mask.Width) * cellSize / 2
	top := float64(mask.Height) * cellSize / 2

	// One front+back face per filled cell so per-pixel color detail is
	// preserved.  Boundary side faces are added separately below.
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			if !mask.At(x, y) {
				continue
			}
			// Use the PNG color directly. Phong shading (ambient 0.5 +
			// diffuse 0.9) already lifts the base color enough for terminal
			// display; pre-multiplying further would clip bright channels and
			// wash out internal features (eyes, outlines, etc.).
			cellColor := mask.ColorAt(x, y)
			addGlyphCellFaces(model, left, top, cellSize, halfDepth, x, y, cellColor)
			addGlyphBoundaryFaces(model, mask, left, top, cellSize, halfDepth, x, y, blendColor(cellColor, sideColor, 0.18))
		}
	}

	if len(model.Faces) == 0 {
		return nil
	}
	return model
}

// addGlyphCellFaces adds one front face and one back face for a single filled
// mask cell.  Using one face per cell (rather than one per row-run) preserves
// the per-pixel color sampled from the source image.
func addGlyphCellFaces(model *Model, left, top, cellSize, halfDepth float64, x, y int, color Color) {
	x0 := left + float64(x)*cellSize
	x1 := x0 + cellSize
	y0 := top - float64(y)*cellSize
	y1 := y0 - cellSize

	// Front face (facing +Z, CCW winding when viewed from +Z).
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: x0, Y: y0, Z: halfDepth},
			{X: x1, Y: y0, Z: halfDepth},
			{X: x1, Y: y1, Z: halfDepth},
			{X: x0, Y: y1, Z: halfDepth},
		},
		Char:       '█',
		RenderMode: FaceRenderFill,
		Color:      color,
		HasColor:   true,
	})

	// Back face (facing -Z, reversed winding).
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: x0, Y: y0, Z: -halfDepth},
			{X: x0, Y: y1, Z: -halfDepth},
			{X: x1, Y: y1, Z: -halfDepth},
			{X: x1, Y: y0, Z: -halfDepth},
		},
		Char:       '█',
		RenderMode: FaceRenderFill,
		Color:      color,
		HasColor:   true,
	})
}

// rasterImageMask downsamples an alpha-backed image into a square binary mask.
func rasterImageMask(img image.Image, resolution int) glyphMask {
	src := img.Bounds()
	srcW, srcH := src.Dx(), src.Dy()
	if srcW <= 0 || srcH <= 0 || resolution <= 0 {
		return glyphMask{}
	}

	scale := math.Max(float64(srcW), float64(srcH)) / float64(resolution)
	if scale < 1 {
		scale = 1
	}
	outW := maxInt(1, int(math.Ceil(float64(srcW)/scale)))
	outH := maxInt(1, int(math.Ceil(float64(srcH)/scale)))
	offsetX := (resolution - outW) / 2
	offsetY := (resolution - outH) / 2

	mask := glyphMask{
		Width:  resolution,
		Height: resolution,
		Filled: make([]bool, resolution*resolution),
		Colors: make([]Color, resolution*resolution),
	}

	for y := 0; y < outH; y++ {
		srcMinY := src.Min.Y + int(math.Floor(float64(y)*scale))
		srcMaxY := src.Min.Y + int(math.Ceil(float64(y+1)*scale))
		if srcMaxY <= srcMinY {
			srcMaxY = srcMinY + 1
		}
		if srcMaxY > src.Max.Y {
			srcMaxY = src.Max.Y
		}

		for x := 0; x < outW; x++ {
			srcMinX := src.Min.X + int(math.Floor(float64(x)*scale))
			srcMaxX := src.Min.X + int(math.Ceil(float64(x+1)*scale))
			if srcMaxX <= srcMinX {
				srcMaxX = srcMinX + 1
			}
			if srcMaxX > src.Max.X {
				srcMaxX = src.Max.X
			}

			var alphaSum uint64
			var redSum uint64
			var greenSum uint64
			var blueSum uint64
			var samples uint64
			var visibleSamples uint64
			for sy := srcMinY; sy < srcMaxY; sy++ {
				for sx := srcMinX; sx < srcMaxX; sx++ {
					r, g, b, a := img.At(sx, sy).RGBA()
					alphaSum += uint64(a)
					samples++
					if a == 0 {
						continue
					}
					// Go's color.RGBA() always returns pre-multiplied values
					// (r_premult = r_actual * a / 65535).  Un-premultiply here
					// so averaging across samples yields the true channel
					// intensities rather than alpha-dimmed values.
					redSum += uint64(r) * 65535 / uint64(a)
					greenSum += uint64(g) * 65535 / uint64(a)
					blueSum += uint64(b) * 65535 / uint64(a)
					visibleSamples++
				}
			}
			if samples == 0 {
				continue
			}

			// Fill the cell if at least one source pixel had meaningful alpha.
			// Using a low threshold (1/64 of max) rather than 1/8 ensures
			// thin features and edge cells are not silently dropped during
			// downsampling.
			avgAlpha := alphaSum / samples
			if avgAlpha < 0x0400 {
				continue
			}
			idx := (offsetY+y)*mask.Width + (offsetX + x)
			mask.Filled[idx] = true
			if visibleSamples > 0 {
				mask.Colors[idx] = Color{
					R: clampFloat(float64(redSum)/float64(visibleSamples)/65535.0, 0, 1),
					G: clampFloat(float64(greenSum)/float64(visibleSamples)/65535.0, 0, 1),
					B: clampFloat(float64(blueSum)/float64(visibleSamples)/65535.0, 0, 1),
				}
			}
		}
	}

	return mask
}

// trimGlyphMask crops away fully empty rows and columns around the active
// symbol so the resulting model is centered tightly around the glyph.
func trimGlyphMask(mask glyphMask) (glyphMask, bool) {
	if mask.Width == 0 || mask.Height == 0 || len(mask.Filled) != mask.Width*mask.Height {
		return glyphMask{}, false
	}

	minX, minY := mask.Width, mask.Height
	maxX, maxY := -1, -1
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			if !mask.At(x, y) {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if maxX < minX || maxY < minY {
		return glyphMask{}, false
	}

	width := maxX - minX + 1
	height := maxY - minY + 1
	trimmed := glyphMask{
		Width:  width,
		Height: height,
		Filled: make([]bool, width*height),
		Colors: make([]Color, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			trimmed.Filled[idx] = mask.At(minX+x, minY+y)
			trimmed.Colors[idx] = mask.ColorAt(minX+x, minY+y)
		}
	}
	return trimmed, true
}

// addGlyphBoundaryFaces adds the exposed side quads around one filled mask
// cell. Only boundaries touching empty space produce geometry.
func addGlyphBoundaryFaces(model *Model, mask glyphMask, left, top, cellSize, halfDepth float64, x, y int, color Color) {
	x0 := left + float64(x)*cellSize
	x1 := x0 + cellSize
	y0 := top - float64(y)*cellSize
	y1 := y0 - cellSize

	if !mask.At(x-1, y) {
		addGlyphSideFace(model,
			Vector3D{X: x0, Y: y0, Z: halfDepth},
			Vector3D{X: x0, Y: y1, Z: halfDepth},
			Vector3D{X: x0, Y: y1, Z: -halfDepth},
			Vector3D{X: x0, Y: y0, Z: -halfDepth},
			color,
		)
	}
	if !mask.At(x+1, y) {
		addGlyphSideFace(model,
			Vector3D{X: x1, Y: y0, Z: halfDepth},
			Vector3D{X: x1, Y: y0, Z: -halfDepth},
			Vector3D{X: x1, Y: y1, Z: -halfDepth},
			Vector3D{X: x1, Y: y1, Z: halfDepth},
			color,
		)
	}
	if !mask.At(x, y-1) {
		addGlyphSideFace(model,
			Vector3D{X: x0, Y: y0, Z: halfDepth},
			Vector3D{X: x0, Y: y0, Z: -halfDepth},
			Vector3D{X: x1, Y: y0, Z: -halfDepth},
			Vector3D{X: x1, Y: y0, Z: halfDepth},
			color,
		)
	}
	if !mask.At(x, y+1) {
		addGlyphSideFace(model,
			Vector3D{X: x0, Y: y1, Z: halfDepth},
			Vector3D{X: x1, Y: y1, Z: halfDepth},
			Vector3D{X: x1, Y: y1, Z: -halfDepth},
			Vector3D{X: x0, Y: y1, Z: -halfDepth},
			color,
		)
	}
}

// addGlyphSideFace appends one extruded side quad to the model.
func addGlyphSideFace(model *Model, a, b, c, d Vector3D, color Color) {
	model.AddFace(Face{
		Vertices:   []Vector3D{a, b, c, d},
		Char:       '█',
		RenderMode: FaceRenderFill,
		Color:      color,
		HasColor:   true,
	})
}

// maskToBrailleLines encodes a glyphMask into cols×rows braille Unicode lines.
//
// Each braille character represents a 2×4 pixel cell.  The mask is sampled
// with integer scaling so there is no floating-point drift across cells.
func maskToBrailleLines(mask glyphMask, cols, rows int) []string {
	if cols <= 0 || rows <= 0 || mask.Width == 0 || mask.Height == 0 {
		return nil
	}

	// Each braille character covers 2 columns × 4 rows of pixels.
	pixW := cols * 2
	pixH := rows * 4

	lines := make([]string, rows)
	for row := 0; row < rows; row++ {
		var sb strings.Builder
		for col := 0; col < cols; col++ {
			var bits byte
			// Sample the 2×4 sub-pixels for this braille cell and map each set
			// pixel to its corresponding braille dot bit position.
			// Integer scaling avoids floating-point rounding drift across cells.
			for py := 0; py < 4; py++ {
				for px := 0; px < 2; px++ {
					mx := (col*2 + px) * mask.Width / pixW
					my := (row*4 + py) * mask.Height / pixH
					if mask.At(mx, my) {
						bits |= brailleDotBit(px, py)
					}
				}
			}
			sb.WriteRune(rune(0x2800 + int(bits)))
		}
		lines[row] = sb.String()
	}
	return lines
}

// brailleDotBits maps [col][row] → the Unicode braille bitmask for that dot.
//
// Standard braille dot layout (dots 1–8, Unicode U+2800 base):
//
//	col 0  col 1
//	dot 1  dot 4   row 0   → bits 0, 3
//	dot 2  dot 5   row 1   → bits 1, 4
//	dot 3  dot 6   row 2   → bits 2, 5
//	dot 7  dot 8   row 3   → bits 6, 7
var brailleDotBits = [2][4]byte{
	{1 << 0, 1 << 1, 1 << 2, 1 << 6}, // col 0
	{1 << 3, 1 << 4, 1 << 5, 1 << 7}, // col 1
}

// brailleDotBit returns the bitmask for the braille dot at pixel column px
// (0–1) and pixel row py (0–3). Returns 0 for out-of-range inputs.
func brailleDotBit(px, py int) byte {
	if px < 0 || px > 1 || py < 0 || py > 3 {
		return 0
	}
	return brailleDotBits[px][py]
}

// maxInt returns the larger integer.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// blendColor mixes a base color with an accent color by the given weight.
func blendColor(base, accent Color, weight float64) Color {
	w := clampFloat(weight, 0, 1)
	return Color{
		R: clampFloat(base.R*(1-w)+accent.R*w, 0, 1),
		G: clampFloat(base.G*(1-w)+accent.G*w, 0, 1),
		B: clampFloat(base.B*(1-w)+accent.B*w, 0, 1),
	}
}
