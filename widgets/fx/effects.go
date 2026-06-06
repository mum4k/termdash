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
	"math/rand"
	"sync"
	"time"

	"github.com/mum4k/termdash/private/canvas/buffer"
)

// ── Fade ─────────────────────────────────────────────────────────────────────

// FadeIn returns an Effect that transitions the widget from invisible to fully
// visible over duration d.
//
// The transition is three-step so it works on every color mode and with every
// widget:
//
//	t ∈ [0, 0.30)  → cell is blank
//	t ∈ [0.30, 0.65) → cell is visible but dimmed
//	t ∈ [0.65, 1.0]  → cell is fully bright
func FadeIn(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			if isBlank(src) {
				return src
			}
			t := clampT(elapsed.Seconds() / total.Seconds())
			switch {
			case t < 0.30:
				return nil
			case t < 0.65:
				c := src.Copy()
				c.Opts.Dim = true
				c.Opts.Bold = false
				return c
			default:
				c := src.Copy()
				c.Opts.Dim = false
				return c
			}
		},
	}
}

// FadeOut returns an Effect that transitions the widget from fully visible to
// invisible over duration d.  It is the time-reverse of FadeIn.
func FadeOut(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			if isBlank(src) {
				return src
			}
			t := clampT(elapsed.Seconds() / total.Seconds())
			switch {
			case t > 0.70:
				return nil
			case t > 0.35:
				c := src.Copy()
				c.Opts.Dim = true
				c.Opts.Bold = false
				return c
			default:
				return src
			}
		},
	}
}

// ColorFadeIn returns an Effect that interpolates cell foreground and background
// colors from black to their target values over duration d.
//
// This effect produces the smoothest results when widgets use ColorRGB6 colors
// (termdash ColorMode256).  For named or ColorNumber colors it falls back to
// the same three-step dim transition as FadeIn.
func ColorFadeIn(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			if isBlank(src) {
				return src
			}
			t := clampT(elapsed.Seconds() / total.Seconds())
			c := src.Copy()

			fgLerped := false
			if fg, ok := lerpColorToBlack(src.Opts.FgColor, t); ok {
				c.Opts.FgColor = fg
				fgLerped = true
			}
			if bg, ok := lerpColorToBlack(src.Opts.BgColor, t); ok {
				c.Opts.BgColor = bg
			}

			// Fallback dim-step for non-RGB6 foreground colors.
			if !fgLerped {
				switch {
				case t < 0.30:
					return nil
				case t < 0.65:
					c.Opts.Dim = true
					c.Opts.Bold = false
				default:
					c.Opts.Dim = false
				}
			}
			return c
		},
	}
}

// ColorFadeOut returns an Effect that interpolates colors to black over d.
// It is the time-reverse of ColorFadeIn.
func ColorFadeOut(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			if isBlank(src) {
				return src
			}
			// Invert t so the color drains toward black.
			t := 1.0 - clampT(elapsed.Seconds()/total.Seconds())
			c := src.Copy()

			fgLerped := false
			if fg, ok := lerpColorToBlack(src.Opts.FgColor, t); ok {
				c.Opts.FgColor = fg
				fgLerped = true
			}
			if bg, ok := lerpColorToBlack(src.Opts.BgColor, t); ok {
				c.Opts.BgColor = bg
			}

			if !fgLerped {
				switch {
				case t < 0.30:
					return nil
				case t < 0.65:
					c.Opts.Dim = true
					c.Opts.Bold = false
				default:
					c.Opts.Dim = false
				}
			}
			return c
		},
	}
}

// ── Sweeps ────────────────────────────────────────────────────────────────────

// SweepLeft returns an Effect that reveals the canvas from left to right,
// column by column, over duration d.
func SweepLeft(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			t := clampT(elapsed.Seconds() / total.Seconds())
			reveal := int(math.Round(float64(w) * t))
			if x < reveal {
				return src
			}
			return nil
		},
	}
}

// SweepRight returns an Effect that reveals the canvas from right to left
// over duration d.
func SweepRight(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			t := clampT(elapsed.Seconds() / total.Seconds())
			revealFrom := w - int(math.Round(float64(w)*t))
			if x >= revealFrom {
				return src
			}
			return nil
		},
	}
}

// SweepDown returns an Effect that reveals the canvas from top to bottom,
// row by row, over duration d.
func SweepDown(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			t := clampT(elapsed.Seconds() / total.Seconds())
			reveal := int(math.Round(float64(h) * t))
			if y < reveal {
				return src
			}
			return nil
		},
	}
}

// SweepUp returns an Effect that reveals the canvas from bottom to top
// over duration d.
func SweepUp(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			t := clampT(elapsed.Seconds() / total.Seconds())
			revealFrom := h - int(math.Round(float64(h)*t))
			if y >= revealFrom {
				return src
			}
			return nil
		},
	}
}

// ── Dissolve ──────────────────────────────────────────────────────────────────

// Dissolve returns an Effect that reveals canvas cells in a pseudo-random order
// over duration d.  seed controls the pattern; the same seed always produces
// the same reveal sequence regardless of canvas contents.
//
// The permutation is computed lazily on the first Draw call and cached for
// subsequent frames.  If the canvas is resized the permutation is recomputed
// automatically.
func Dissolve(d time.Duration, seed int64) Effect {
	type state struct {
		w, h       int
		revealTime []int // revealTime[cellIdx] = step at which the cell appears
	}
	var (
		mu sync.Mutex
		st *state
	)

	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			mu.Lock()
			if st == nil || st.w != w || st.h != h {
				n := w * h
				// rand.Perm builds a random permutation: perm[step] = cellIdx.
				perm := rand.New(rand.NewSource(seed)).Perm(n)
				// Invert to revealTime: revealTime[cellIdx] = step.
				rt := make([]int, n)
				for step, cellIdx := range perm {
					rt[cellIdx] = step
				}
				st = &state{w: w, h: h, revealTime: rt}
			}
			rt := st.revealTime
			mu.Unlock()

			t := clampT(elapsed.Seconds() / total.Seconds())
			threshold := int(math.Round(float64(w*h) * t))
			idx := y*w + x
			if idx >= 0 && idx < len(rt) && rt[idx] < threshold {
				return src
			}
			return nil
		},
	}
}

// ── Glitch ────────────────────────────────────────────────────────────────────

// Glitch returns an Effect that briefly floods the canvas with block-drawing
// noise before settling to the real content.  It peaks at mid-animation and
// decays away by t=1.
//
// seed makes the noise pattern reproducible.  Because noise is derived from a
// fast integer hash (no shared mutable state) the function is allocation-free
// and goroutine-safe.
func Glitch(d time.Duration, seed int64) Effect {
	// Characters used for the noise burst — box-drawing / block elements that
	// look appropriately "corrupted" without being distracting.
	noiseRunes := []rune{
		'░', '▒', '▓', '█',
		'╬', '╪', '╫', '║', '═',
		'╔', '╗', '╚', '╝',
		'┼', '│', '─', '·',
	}
	nNoise := int64(len(noiseRunes))

	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			if isBlank(src) {
				return src
			}
			t := clampT(elapsed.Seconds() / total.Seconds())

			// Noise probability peaks at t=0.5 (sin curve).
			noiseChance := math.Sin(t*math.Pi) * 0.72

			// Deterministic per-cell hash — changes each frame because elapsed
			// is baked in.  No allocations, no shared state.
			h64 := cellHash(seed, x, y, elapsed.Milliseconds())
			roll := float64(h64&0xFFFF) / 0xFFFF

			if roll < noiseChance {
				// Show a noise character using the cell's original colors.
				c := src.Copy()
				c.Rune = noiseRunes[int64(h64>>16)%nNoise]
				return c
			}

			// Before the midpoint, hide cells that aren't replaced by noise.
			if t < 0.45 {
				return nil
			}
			return src
		},
	}
}

// ── Wipe (diagonal) ───────────────────────────────────────────────────────────

// WipeDiagonal returns an Effect that reveals the canvas along a diagonal
// sweep from the top-left corner to the bottom-right over duration d.
// The reveal frontier is a line perpendicular to the main diagonal.
func WipeDiagonal(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			t := clampT(elapsed.Seconds() / total.Seconds())
			// Normalise each axis to [0,1] and use their sum as the reveal metric.
			nx := float64(x) / math.Max(float64(w-1), 1)
			ny := float64(y) / math.Max(float64(h-1), 1)
			// The frontier advances from 0 to 2 (max sum of nx+ny).
			if nx+ny <= t*2 {
				return src
			}
			return nil
		},
	}
}

// ── Scan Line ─────────────────────────────────────────────────────────────────

// ScanLine returns an Effect that draws a bright horizontal scan-line that
// sweeps from top to bottom, illuminating cells it passes over and leaving them
// fully revealed behind it.
func ScanLine(d time.Duration) Effect {
	return Effect{
		Duration: d,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			t := clampT(elapsed.Seconds() / total.Seconds())
			scanRow := int(math.Round(float64(h-1) * t))

			switch {
			case y < scanRow:
				// Already revealed.
				return src
			case y == scanRow:
				// On the scan line: bold highlight.
				c := src.Copy()
				c.Opts.Bold = true
				c.Opts.Dim = false
				return c
			default:
				// Not yet revealed.
				return nil
			}
		},
	}
}

// ── Scramble ──────────────────────────────────────────────────────────────────

// Scramble returns an Effect that permanently replaces all non-blank cells
// with animated block-drawing noise, hiding the underlying widget content.
//
// The noise pattern changes every 80 ms so the scramble appears to animate.
// Use NewLooping to keep the effect running indefinitely:
//
//	fx.NewLooping(inner, fx.Scramble(42))
//
// seed makes the noise pattern reproducible across runs.
func Scramble(seed int64) Effect {
	noiseRunes := []rune{
		'░', '▒', '▓', '█',
		'╬', '╪', '╫', '║', '═',
		'╔', '╗', '╚', '╝',
		'┼', '│', '─', '·',
		'•', '◆', '◇', '○', '●',
	}
	nNoise := int64(len(noiseRunes))

	return Effect{
		Duration: time.Second, // short; pair with NewLooping for continuous scramble
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			if isBlank(src) {
				return src
			}
			// Change the noise pattern every 80 ms for an animated scramble.
			frame := elapsed.Milliseconds() / 80
			h64 := cellHash(seed, x, y, frame)
			c := src.Copy()
			c.Rune = noiseRunes[int64(h64>>16)%nNoise]
			c.Opts.Dim = true
			c.Opts.Bold = false
			return c
		},
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// isBlank reports whether a cell carries no visible glyph.
func isBlank(c *buffer.Cell) bool {
	return c.Rune == 0 || c.Rune == ' '
}

// clampT restricts t to [0, 1].
func clampT(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// cellHash returns a deterministic pseudo-random uint64 derived from the given
// seed, cell position, and frame timestamp.  Uses the Murmur3-inspired finalizer
// mix so nearby (seed, x, y, ms) values produce uncorrelated outputs.
func cellHash(seed int64, x, y int, frameMS int64) uint64 {
	v := uint64(seed) ^
		uint64(x)*0x9e3779b97f4a7c15 ^
		uint64(y)*0x6c62272e07bb0142 ^
		uint64(frameMS)*0x517cc1b727220a95

	v ^= v >> 33
	v *= 0xff51afd7ed558ccd
	v ^= v >> 33
	v *= 0xc4ceb9fe1a85ec53
	v ^= v >> 33
	return v
}
