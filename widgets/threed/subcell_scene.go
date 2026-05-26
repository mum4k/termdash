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

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
)

const (
	// subcellCoverageSamples is the number of samples per axis used when
	// estimating polygon coverage of a single braille dot.  4 gives a 4×4 = 16
	// sample grid, which is 4× more than the previous 2×2 grid and produces
	// visibly smoother diagonal edges.  Interior and exterior dots are
	// short-circuited before the full grid is evaluated (see polygonCoverage),
	// so the average cost is roughly equivalent to the old 2-sample approach.
	subcellCoverageSamples = 4

	subcellSolidCoverage         = 0.5
	subcellEdgeThresholdMin      = 0.18
	subcellEdgeThresholdVariance = 0.18
)

// subcellDither is an ordered 4×2 Bayer-style matrix used in subcellCovered to
// keep partial-coverage edge pixels visible without blurring diagonal edges.
var subcellDither = [braille.RowMult][braille.ColMult]float64{
	{0.0000, 0.5000},
	{0.7500, 0.2500},
	{0.1875, 0.6875},
	{0.9375, 0.4375},
}

// subcellScene stores filled-polygon coverage at braille-pixel resolution
// (2 cols × 4 rows per terminal cell) so the final output preserves more shape
// detail than a plain one-cell-per-sample raster.
type subcellScene struct {
	cellArea image.Rectangle
	width    int
	height   int
	filled   []bool
	colors   []Color
	// depth holds the face depth (average Z) of the polygon that last painted
	// each dot.  It is initialized to +Inf and updated whenever a closer face
	// paints over an existing dot, giving a per-dot depth buffer that corrects
	// any ordering inaccuracies left by the painter's-algorithm polygon sort.
	depth     []float64
	scaledBuf []Vector2D // reused scratch buffer for FillPolygon — avoids per-call alloc
	dirty     bool
	dirtyMinX int
	dirtyMinY int
	dirtyMaxX int
	dirtyMaxY int
}

// newSubcellScene allocates a new high-resolution scene buffer for the given
// terminal cell area.
func newSubcellScene(cellArea image.Rectangle) *subcellScene {
	width := cellArea.Dx() * braille.ColMult
	height := cellArea.Dy() * braille.RowMult
	n := width * height
	depth := make([]float64, n)
	for i := range depth {
		depth[i] = math.Inf(1)
	}
	return &subcellScene{
		cellArea: cellArea,
		width:    width,
		height:   height,
		filled:   make([]bool, n),
		colors:   make([]Color, n),
		depth:    depth,
	}
}

// Clear resets the scene contents while retaining the allocated storage.
func (s *subcellScene) Clear() {
	if !s.dirty {
		return
	}
	for y := s.dirtyMinY; y < s.dirtyMaxY; y++ {
		row := y * s.width
		for x := s.dirtyMinX; x < s.dirtyMaxX; x++ {
			idx := row + x
			s.filled[idx] = false
			s.colors[idx] = Color{}
			s.depth[idx] = math.Inf(1)
		}
	}
	s.dirty = false
	s.dirtyMinX = 0
	s.dirtyMinY = 0
	s.dirtyMaxX = 0
	s.dirtyMaxY = 0
}

// FillPolygon rasterizes the projected polygon into braille subcells.
// faceDepth is the average scene depth of the polygon (smaller = closer to
// camera).  A dot is only painted when faceDepth is closer than the depth
// already stored for that dot, so later-drawn polygons only win if they are
// genuinely in front.
func (s *subcellScene) FillPolygon(points []Vector2D, clr Color, faceDepth float64) {
	s.FillPolygonWithDepths(points, nil, clr, faceDepth)
}

// FillPolygonWithDepths rasterizes the projected polygon into braille subcells
// using per-vertex depths when provided.
func (s *subcellScene) FillPolygonWithDepths(points []Vector2D, depths []float64, clr Color, faceDepth float64) {
	if len(points) < 3 || s.width == 0 || s.height == 0 {
		return
	}

	// Scale the caller's canvas-space points into braille-pixel space.
	// Reuse the scratch buffer to avoid a heap allocation on every call.
	if cap(s.scaledBuf) < len(points) {
		s.scaledBuf = make([]Vector2D, len(points))
	} else {
		s.scaledBuf = s.scaledBuf[:len(points)]
	}
	for i, p := range points {
		s.scaledBuf[i] = Vector2D{
			X: p.X * braille.ColMult,
			Y: p.Y * braille.RowMult,
		}
	}
	scaled := s.scaledBuf

	minX, maxX := scaled[0].X, scaled[0].X
	minY, maxY := scaled[0].Y, scaled[0].Y
	for _, p := range scaled[1:] {
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

	startX := maxInt(0, int(math.Floor(minX)))
	endX := minInt(s.width-1, int(math.Ceil(maxX)))
	startY := maxInt(0, int(math.Floor(minY)))
	endY := minInt(s.height-1, int(math.Ceil(maxY)))

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			coverage := polygonCoverage(float64(x), float64(y), scaled, subcellCoverageSamples)
			if !subcellCovered(coverage, x, y) {
				continue
			}
			idx := y*s.width + x
			dotDepth := polygonDepth(float64(x), float64(y), scaled, depths, faceDepth)
			// Per-dot depth test: only paint if this face is closer than
			// whatever polygon last touched this dot.
			if dotDepth >= s.depth[idx] {
				continue
			}
			s.filled[idx] = true
			s.colors[idx] = clr
			s.depth[idx] = dotDepth
			s.markDirty(x, y)
		}
	}
}

// markDirty records the smallest subcell bounds touched this frame.
func (s *subcellScene) markDirty(x, y int) {
	if !s.dirty {
		s.dirty = true
		s.dirtyMinX = x
		s.dirtyMinY = y
		s.dirtyMaxX = x + 1
		s.dirtyMaxY = y + 1
		return
	}
	if x < s.dirtyMinX {
		s.dirtyMinX = x
	}
	if y < s.dirtyMinY {
		s.dirtyMinY = y
	}
	if x >= s.dirtyMaxX {
		s.dirtyMaxX = x + 1
	}
	if y >= s.dirtyMaxY {
		s.dirtyMaxY = y + 1
	}
}

// subcellFastInset is the inset from the dot boundary used by the
// polygonCoverage fast-path corner tests.  A small inset avoids landing
// exactly on polygon edges where the ray-casting result can be ambiguous.
const subcellFastInset = 0.05

// polygonCoverage returns the fraction of a single braille dot (the unit
// square at [x, x+1) × [y, y+1)) covered by the polygon.
//
// Fast-path: before running the full N×N sample grid, five cheap point-in-
// polygon tests (center + four near-corners) classify the dot:
//
//   - All 5 inside  → coverage 1.0  (interior dot, skip full grid)
//   - All 5 outside → coverage 0.0  (exterior dot, skip full grid)
//   - Mixed         → run full N×N grid for an accurate edge estimate
//
// For a typical solid convex polygon, ~70-80% of bounding-box dots are
// interior and ~15% are exterior, so the full grid only runs for the ~10-15%
// that lie on the edge.  This makes the net cost of a 4×4 (16-sample) grid
// comparable to the old 2×2 (4-sample) grid while producing 4× better edge
// coverage resolution.
func polygonCoverage(x, y float64, polygon []Vector2D, samples int) float64 {
	// Five-point fast classification: center + four inset corners.
	cx, cy := x+0.5, y+0.5
	lo, hi := subcellFastInset, 1-subcellFastInset
	fastIn := 0
	if pointInPolygon(cx, cy, polygon) {
		fastIn++
	}
	if pointInPolygon(x+lo, y+lo, polygon) {
		fastIn++
	}
	if pointInPolygon(x+hi, y+lo, polygon) {
		fastIn++
	}
	if pointInPolygon(x+lo, y+hi, polygon) {
		fastIn++
	}
	if pointInPolygon(x+hi, y+hi, polygon) {
		fastIn++
	}
	if fastIn == 5 {
		return 1.0 // clearly interior
	}
	if fastIn == 0 {
		return 0.0 // clearly exterior
	}

	// Edge dot: run the full sample grid for an accurate coverage fraction.
	if samples <= 1 {
		// With only one sample we already have the center result above.
		if pointInPolygon(cx, cy, polygon) {
			return 1.0
		}
		return 0.0
	}
	var covered int
	step := 1.0 / float64(samples)
	for sy := 0; sy < samples; sy++ {
		py := y + (float64(sy)+0.5)*step
		for sx := 0; sx < samples; sx++ {
			px := x + (float64(sx)+0.5)*step
			if pointInPolygon(px, py, polygon) {
				covered++
			}
		}
	}
	return float64(covered) / float64(samples*samples)
}

// polygonDepth estimates the polygon depth at a covered braille dot.
func polygonDepth(x, y float64, polygon []Vector2D, depths []float64, fallback float64) float64 {
	if len(depths) != len(polygon) {
		return fallback
	}

	px, py := x+0.5, y+0.5
	if !pointInPolygon(px, py, polygon) {
		var ok bool
		px, py, ok = firstCoveredSample(x, y, polygon, subcellCoverageSamples)
		if !ok {
			return fallback
		}
	}

	for i := 1; i < len(polygon)-1; i++ {
		if depth, ok := triangleDepth(
			Vector2D{X: px, Y: py},
			polygon[0], polygon[i], polygon[i+1],
			depths[0], depths[i], depths[i+1],
		); ok {
			return depth
		}
	}
	return fallback
}

func firstCoveredSample(x, y float64, polygon []Vector2D, samples int) (float64, float64, bool) {
	if samples <= 0 {
		samples = 1
	}
	step := 1.0 / float64(samples)
	for sy := 0; sy < samples; sy++ {
		py := y + (float64(sy)+0.5)*step
		for sx := 0; sx < samples; sx++ {
			px := x + (float64(sx)+0.5)*step
			if pointInPolygon(px, py, polygon) {
				return px, py, true
			}
		}
	}
	return 0, 0, false
}

func triangleDepth(p, a, b, c Vector2D, da, db, dc float64) (float64, bool) {
	v0x, v0y := b.X-a.X, b.Y-a.Y
	v1x, v1y := c.X-a.X, c.Y-a.Y
	v2x, v2y := p.X-a.X, p.Y-a.Y

	d00 := v0x*v0x + v0y*v0y
	d01 := v0x*v1x + v0y*v1y
	d11 := v1x*v1x + v1y*v1y
	d20 := v2x*v0x + v2y*v0y
	d21 := v2x*v1x + v2y*v1y
	denom := d00*d11 - d01*d01
	if math.Abs(denom) < 1e-12 {
		return 0, false
	}

	v := (d11*d20 - d01*d21) / denom
	w := (d00*d21 - d01*d20) / denom
	u := 1 - v - w
	const epsilon = 1e-9
	if u < -epsilon || v < -epsilon || w < -epsilon {
		return 0, false
	}
	return u*da + v*db + w*dc, true
}

// subcellCovered converts fractional coverage into a braille-dot on/off
// decision. The ordered dither keeps partial edge samples visible without
// turning every shallow diagonal into a heavy blur.
func subcellCovered(coverage float64, x, y int) bool {
	if coverage <= 0 {
		return false
	}
	if coverage >= subcellSolidCoverage {
		return true
	}
	dither := subcellDither[y%braille.RowMult][x%braille.ColMult]
	threshold := subcellEdgeThresholdMin + dither*subcellEdgeThresholdVariance
	return coverage >= threshold
}

const unicodeBrailleOffset = 0x2800

// CopyTo converts the subcell scene into braille characters and writes them to
// the destination canvas.
//
// Only terminal cells touched by filled subcells are visited, and braille runes
// are assembled directly instead of allocating a nested braille canvas.
func (s *subcellScene) CopyTo(dst *canvas.Canvas) error {
	if s.width == 0 || s.height == 0 || !s.dirty {
		return nil
	}

	startCellX := s.dirtyMinX / braille.ColMult
	endCellX := minInt(s.cellArea.Dx(), (s.dirtyMaxX+braille.ColMult-1)/braille.ColMult)
	startCellY := s.dirtyMinY / braille.RowMult
	endCellY := minInt(s.cellArea.Dy(), (s.dirtyMaxY+braille.RowMult-1)/braille.RowMult)

	for cellY := startCellY; cellY < endCellY; cellY++ {
		for cellX := startCellX; cellX < endCellX; cellX++ {
			var colorSum Color
			litCount := 0
			var bits byte

			for py := 0; py < braille.RowMult; py++ {
				for px := 0; px < braille.ColMult; px++ {
					x := cellX*braille.ColMult + px
					y := cellY*braille.RowMult + py
					idx := y*s.width + x
					if !s.filled[idx] {
						continue
					}
					colorSum = colorSum.Add(s.colors[idx])
					litCount++
					bits |= brailleDotBit(px, py)
				}
			}

			if litCount == 0 {
				continue
			}

			// Average color across all lit dots in this cell.
			avgColor := colorSum.Multiply(1 / float64(litCount))
			if _, err := dst.SetCell(image.Point{X: cellX, Y: cellY}, rune(unicodeBrailleOffset)|rune(bits), cell.FgColor(avgColor.ToCellColor())); err != nil {
				return err
			}
		}
	}

	return nil
}

// hasFill reports whether p's terminal cell contains any filled subcell.
func (s *subcellScene) hasFill(p image.Point) bool {
	if s == nil || s.width == 0 || s.height == 0 || !p.In(s.cellArea) {
		return false
	}

	baseX := p.X * braille.ColMult
	baseY := p.Y * braille.RowMult
	for py := 0; py < braille.RowMult; py++ {
		for px := 0; px < braille.ColMult; px++ {
			x := baseX + px
			y := baseY + py
			idx := y*s.width + x
			if s.filled[idx] {
				return true
			}
		}
	}
	return false
}

// minInt returns the smaller integer.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
