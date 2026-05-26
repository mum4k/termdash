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
	"math"
	"unicode"

	"github.com/mum4k/termdash/private/runewidth"
)

const (
	defaultSpinnerStarFallback   = '✶'
	defaultSpinnerStarOuter      = 1.25
	defaultSpinnerStarInner      = 0.55
	defaultSpinnerStarDepth      = 0.34
	defaultSpinnerStarPoints     = 5
	defaultSpinnerStarRefinement = 5
)

// RenderableRune converts a UTF-8 frame string into a rune that the threed
// renderer can safely use for face filling.
//
// The threed renderer is cell-based, so wide glyphs and multi-symbol strings
// cannot be drawn reliably as face characters. This helper keeps simple
// single-cell symbols as-is, accepts harmless trailing variation selectors or
// combining marks, and falls back otherwise.
func RenderableRune(frame string, fallback rune) rune {
	rs := []rune(frame)
	if len(rs) == 0 {
		return fallback
	}
	if rw := runewidth.RuneWidth(rs[0]); rw != 1 && rw != 2 {
		return fallback
	}
	for _, r := range rs[1:] {
		if runewidth.RuneWidth(r) != 0 && !unicode.IsMark(r) {
			return fallback
		}
	}
	return rs[0]
}

// NewAnimatedSpinnerStarPrism creates a dense star prism model that is sized
// and colored for spinner-driven animation sequences.
//
// The frame argument may be any UTF-8 spinner frame. If the frame cannot be
// rendered safely as a single-cell face character, the model falls back to a
// compatible star glyph.
func NewAnimatedSpinnerStarPrism(frame string, step int) *Model {
	char := RenderableRune(frame, defaultSpinnerStarFallback)

	baseFront := make([]Vector3D, 0, defaultSpinnerStarPoints*2)
	baseBack := make([]Vector3D, 0, defaultSpinnerStarPoints*2)
	for i := 0; i < defaultSpinnerStarPoints*2; i++ {
		angle := (math.Pi / float64(defaultSpinnerStarPoints)) * float64(i)
		radius := defaultSpinnerStarOuter
		if i%2 == 1 {
			radius = defaultSpinnerStarInner
		}
		x := math.Sin(angle) * radius
		y := math.Cos(angle) * radius
		baseFront = append(baseFront, Vector3D{X: x, Y: y, Z: defaultSpinnerStarDepth})
		baseBack = append(baseBack, Vector3D{X: x, Y: y, Z: -defaultSpinnerStarDepth})
	}

	front := densifyClosedPath(baseFront, defaultSpinnerStarRefinement)
	back := densifyClosedPath(baseBack, defaultSpinnerStarRefinement)
	frontColor, sideColor, backColor := spinnerStarColors(step)

	model := NewModel()
	model.AddFace(Face{Vertices: front, Char: char, Color: frontColor, HasColor: true})

	reversedBack := make([]Vector3D, len(back))
	for i := range back {
		reversedBack[i] = back[len(back)-1-i]
	}
	model.AddFace(Face{Vertices: reversedBack, Char: char, Color: backColor, HasColor: true})

	for i := range front {
		next := (i + 1) % len(front)
		model.AddFace(Face{
			Vertices: []Vector3D{front[i], front[next], back[next], back[i]},
			Char:     char,
			Color:    sideColor,
			HasColor: true,
		})
	}

	return model
}

// NewAnimatedSymbolSpinner creates a symbol-driven model for UTF-8 glyphs.
//
// Symbols are rendered directly into a small animated prism. The helper does
// not fetch, embed, or rasterize external artwork.
func NewAnimatedSymbolSpinner(frame string, step int) *Model {
	return NewAnimatedSpinnerStarPrism(frame, step)
}

// spinnerStarColors returns a cool palette that shifts subtly with the frame.
func spinnerStarColors(step int) (front, side, back Color) {
	phase := float64(step%6) / 5.0
	glow := 0.14 * math.Sin(phase*math.Pi)
	return Color{R: 0.84 + glow, G: 0.93 + glow/2, B: 1.0},
		Color{R: 0.46 + glow/3, G: 0.68 + glow/2, B: 0.94},
		Color{R: 0.62 + glow/4, G: 0.82 + glow/3, B: 1.0}
}

// densifyClosedPath refines a closed polygon by interpolating extra points
// along each edge so the projected silhouette reads more smoothly.
func densifyClosedPath(points []Vector3D, segments int) []Vector3D {
	if len(points) == 0 || segments <= 1 {
		cp := make([]Vector3D, len(points))
		copy(cp, points)
		return cp
	}

	refined := make([]Vector3D, 0, len(points)*segments)
	for i := range points {
		start := points[i]
		end := points[(i+1)%len(points)]
		for segment := 0; segment < segments; segment++ {
			t := float64(segment) / float64(segments)
			refined = append(refined, interpolateVector3D(start, end, t))
		}
	}
	return refined
}

// interpolateVector3D returns the point between two vertices at the given
// normalized interpolation factor.
func interpolateVector3D(a, b Vector3D, t float64) Vector3D {
	return Vector3D{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
		Z: a.Z + (b.Z-a.Z)*t,
	}
}
