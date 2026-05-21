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

// NewAnimatedSymbolSpinner creates a symbol-driven model for UTF-8 glyphs and
// emoji-like frames.
//
// Plain text characters (codepoints below U+2600) get a lightweight two-face
// billboard so the rune is drawn directly without any CDN fetch.  Emoji and
// symbol characters (U+2600 and above) are extruded from a PNG mask when one
// is available; if no mask is found the function falls back to a braille disc.
func NewAnimatedSymbolSpinner(frame string, step int) *Model {
	if !isEmojiFrame(frame) {
		char := RenderableRune(frame, defaultSpinnerStarFallback)
		return newBillboardModel(char)
	}
	if model := newGlyphMaskModel(frame, step); model != nil {
		return model
	}
	return newBrailleDiscModel(step)
}

// isEmojiFrame reports whether the first rune in frame falls in the emoji /
// symbol range (U+2600 and above).  Characters below that threshold are plain
// text and should be rendered as a direct glyph rather than a CDN-fetched mask.
func isEmojiFrame(frame string) bool {
	for _, r := range frame {
		return r >= 0x2600
	}
	return false
}

// newBillboardModel creates a two-face billboard for the given rune: one face
// with a +Z normal (visible from the front) and one with a -Z normal (visible
// from the back).  Both faces use FaceRenderGlyph so the rune is drawn at the
// face center rather than being rasterised through the braille scene.
func newBillboardModel(char rune) *Model {
	const (
		hw = 1.0 // half-width
		hh = 1.0 // half-height
		dz = 0.1 // half-depth — keeps the billboard thin
	)
	model := NewModel()

	// Front face: counter-clockwise vertices give a +Z outward normal.
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: -hw, Y: -hh, Z: dz},
			{X: hw, Y: -hh, Z: dz},
			{X: hw, Y: hh, Z: dz},
			{X: -hw, Y: hh, Z: dz},
		},
		Char:       char,
		RenderMode: FaceRenderGlyph,
	})

	// Back face: reversed winding gives a -Z outward normal so the glyph is
	// also visible when the billboard has been rotated 180 degrees.
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: -hw, Y: -hh, Z: -dz},
			{X: -hw, Y: hh, Z: -dz},
			{X: hw, Y: hh, Z: -dz},
			{X: hw, Y: -hh, Z: -dz},
		},
		Char:       char,
		RenderMode: FaceRenderGlyph,
	})

	return model
}

// newBrailleDiscModel builds a thin disc (coin) whose faces all use
// FaceRenderFill so they are rasterised through the subcell braille scene
// instead of being written as a literal glyph character.
func newBrailleDiscModel(step int) *Model {
	const (
		radius   = 1.1 // matches star-prism outer radius so scale feels consistent
		depth    = 0.22
		segments = 20
	)

	frontColor, sideColor, backColor := spinnerStarColors(step)

	front := make([]Vector3D, segments)
	back := make([]Vector3D, segments)
	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		x := math.Cos(angle) * radius
		y := math.Sin(angle) * radius
		front[i] = Vector3D{X: x, Y: y, Z: depth}
		back[i] = Vector3D{X: x, Y: y, Z: -depth}
	}

	model := NewModel()

	// Front cap (winding: counter-clockwise when viewed from +Z).
	model.AddFace(Face{
		Vertices:   front,
		Char:       '█',
		RenderMode: FaceRenderFill,
		Color:      frontColor,
		HasColor:   true,
	})

	// Back cap (reversed winding so the normal points toward -Z).
	reversedBack := make([]Vector3D, segments)
	for i, v := range back {
		reversedBack[segments-1-i] = v
	}
	model.AddFace(Face{
		Vertices:   reversedBack,
		Char:       '█',
		RenderMode: FaceRenderFill,
		Color:      backColor,
		HasColor:   true,
	})

	// Side quads connecting front and back rings.
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		model.AddFace(Face{
			Vertices:   []Vector3D{front[i], front[next], back[next], back[i]},
			Char:       '█',
			RenderMode: FaceRenderFill,
			Color:      sideColor,
			HasColor:   true,
		})
	}

	return model
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
