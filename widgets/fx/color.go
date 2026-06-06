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

package fx

import (
	"math"

	"github.com/mum4k/termdash/cell"
)

// lerpColorToBlack interpolates a cell.Color toward black.
//
// t = 0  →  pure black (ColorRGB6(0,0,0))
// t = 1  →  original color unchanged
//
// Only works for colors in the xterm 6×6×6 cube (xterm indices 16–231, i.e.
// termdash ColorRGB6 values).  Returns (c, false) for named colors, ColorDefault,
// and any other indexed color outside the cube.
func lerpColorToBlack(c cell.Color, t float64) (cell.Color, bool) {
	r, g, b, ok := decodeRGB6(c)
	if !ok {
		return c, false
	}
	r2 := int(math.Round(float64(r) * t))
	g2 := int(math.Round(float64(g) * t))
	b2 := int(math.Round(float64(b) * t))
	return cell.ColorRGB6(r2, g2, b2), true
}

// lerpColorRGB6 linearly interpolates between two ColorRGB6 colors.
//
// t = 0  →  from color
// t = 1  →  to color
//
// Returns (from, false) if either color is not an RGB6 color.
func lerpColorRGB6(from, to cell.Color, t float64) (cell.Color, bool) {
	rf, gf, bf, ok1 := decodeRGB6(from)
	rt, gt, bt, ok2 := decodeRGB6(to)
	if !ok1 || !ok2 {
		return from, false
	}
	r := int(math.Round(float64(rf) + float64(rt-rf)*t))
	g := int(math.Round(float64(gf) + float64(gt-gf)*t))
	b := int(math.Round(float64(bf) + float64(bt-bf)*t))
	return cell.ColorRGB6(clamp6(r), clamp6(g), clamp6(b)), true
}

// decodeRGB6 extracts the r, g, b components [0–5] from a termdash ColorRGB6
// color value.
//
// termdash's Color type is off-by-one from the xterm index (ColorDefault = 0,
// so Color(n+1) corresponds to xterm index n).  The 6×6×6 cube occupies xterm
// indices 16–231.
//
// Returns ok=false for ColorDefault, named/system colors (xterm 0–15), and
// grayscale ramp entries (xterm 232–255).
func decodeRGB6(c cell.Color) (r, g, b int, ok bool) {
	// Convert from termdash's off-by-one representation to xterm index.
	idx := int(c) - 1
	if idx < 16 || idx > 231 {
		return 0, 0, 0, false
	}
	n := idx - 16
	return n / 36, (n / 6) % 6, n % 6, true
}

// IsRGB6 reports whether c is a ColorRGB6 color that supports smooth
// color-interpolation effects.
func IsRGB6(c cell.Color) bool {
	_, _, _, ok := decodeRGB6(c)
	return ok
}

// clamp6 clamps v to the valid RGB6 component range [0, 5].
func clamp6(v int) int {
	if v < 0 {
		return 0
	}
	if v > 5 {
		return 5
	}
	return v
}
