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

// Package fx provides a composable, time-driven cell-level effect pipeline for
// termdash widgets.
//
// Wrap any widgetapi.Widget in an EffectWidget and attach a sequence of
// Effects (FadeIn, SweepLeft, Dissolve, …).  Each effect runs for a fixed
// duration, then the next one begins.  Once the sequence is exhausted the
// widget renders normally with no overhead.
//
// Basic usage:
//
//	txt, _ := text.New()
//	fxWidget, _ := fx.New(txt,
//	    fx.FadeIn(400*time.Millisecond),
//	    fx.SweepLeft(300*time.Millisecond),
//	)
//	// Place fxWidget inside a container instead of txt.
//
// Parallel effects:
//
//	fxWidget, _ := fx.New(txt, fx.Parallel(
//	    fx.FadeIn(500*time.Millisecond),
//	    fx.SweepDown(500*time.Millisecond),
//	))
//
// Looping:
//
//	fxWidget, _ := fx.NewLooping(txt, fx.Dissolve(600*time.Millisecond, 42))
package fx

import (
	"image"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/buffer"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// CellFunc is the signature for a cell-level transform applied during animation.
//
//   - elapsed is how long the current effect has been running.
//   - total is the full duration of the current effect.
//   - x, y are the zero-based canvas column and row of the cell.
//   - w, h are the canvas width and height in cells.
//   - src is the cell that the wrapped widget painted; never nil.
//
// Return nil to render the cell as a blank (space with default colors).
// Return src (or a copy of it) to render the original content.
type CellFunc func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell

// Effect describes one step in an animation sequence.
type Effect struct {
	// Duration is how long this step plays before the next one begins.
	Duration time.Duration
	// Fn is called once per canvas cell on every Draw cycle while this
	// effect is the active step.  Must be goroutine-safe.
	Fn CellFunc
}

// Parallel returns a single Effect that applies all of the provided effects
// simultaneously.  The combined duration equals the longest individual duration;
// shorter effects hold their final frame until the combined effect finishes.
func Parallel(effects ...Effect) Effect {
	if len(effects) == 0 {
		return Effect{Duration: 0, Fn: func(_, _ time.Duration, _, _, _, _ int, src *buffer.Cell) *buffer.Cell { return src }}
	}

	// Find the maximum duration so all sub-effects share the same timeline.
	var maxDur time.Duration
	for _, e := range effects {
		if e.Duration > maxDur {
			maxDur = e.Duration
		}
	}

	return Effect{
		Duration: maxDur,
		Fn: func(elapsed, total time.Duration, x, y, w, h int, src *buffer.Cell) *buffer.Cell {
			result := src
			for _, e := range effects {
				// Clamp elapsed to this sub-effect's own duration.
				subElapsed := elapsed
				subTotal := e.Duration
				if subElapsed > subTotal {
					subElapsed = subTotal
				}
				if subTotal == 0 {
					continue
				}
				result = e.Fn(subElapsed, subTotal, x, y, w, h, result)
				if result == nil {
					// If any sub-effect blanks this cell, propagate immediately.
					return nil
				}
			}
			return result
		},
	}
}

// EffectWidget wraps any widgetapi.Widget and plays a sequential pipeline of
// Effects on its rendered output each frame.
//
// All widget methods are fully delegated to the inner widget so keyboard and
// mouse handling, size negotiation, and options work transparently.
//
// Implements widgetapi.Widget.  This object is thread-safe.
type EffectWidget struct {
	mu        sync.Mutex
	inner     widgetapi.Widget
	effects   []Effect
	startTime time.Time
	started   bool
	loop      bool

	// Scratch canvas — reused across frames to avoid per-frame allocations.
	tmp      *canvas.Canvas
	lastSize image.Point
}

// New wraps inner and plays each effect in order.  After all effects finish
// the widget passes through its inner widget's output unmodified.
func New(inner widgetapi.Widget, effects ...Effect) (*EffectWidget, error) {
	return &EffectWidget{
		inner:   inner,
		effects: effects,
	}, nil
}

// NewLooping wraps inner and replays the effect sequence forever.
func NewLooping(inner widgetapi.Widget, effects ...Effect) (*EffectWidget, error) {
	return &EffectWidget{
		inner:   inner,
		effects: effects,
		loop:    true,
	}, nil
}

// Reset restarts the effect sequence from the beginning.
// Safe to call from any goroutine.
func (w *EffectWidget) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.started = false
	w.startTime = time.Time{}
}

// Done reports whether all effects have finished playing.
// Always returns false for a looping EffectWidget.
func (w *EffectWidget) Done() bool {
	if w.loop {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.started {
		return false
	}
	elapsed := time.Since(w.startTime)
	return elapsed >= totalDuration(w.effects)
}

// Draw implements widgetapi.Widget.
func (w *EffectWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	size := cvs.Size()

	// (Re)create the scratch canvas whenever the canvas size changes.
	// Also reset the animation so effects replay cleanly after a terminal resize.
	if w.tmp == nil || w.lastSize != size {
		var err error
		w.tmp, err = canvas.New(cvs.Area())
		if err != nil {
			return err
		}
		w.lastSize = size
		w.started = false
		w.startTime = time.Time{}
	}

	// Start the clock on the very first Draw call.
	if !w.started {
		w.startTime = time.Now()
		w.started = true
	}

	// Paint the inner widget onto the scratch canvas.
	if err := w.tmp.Clear(); err != nil {
		return err
	}
	if err := w.inner.Draw(w.tmp, meta); err != nil {
		return err
	}

	// Locate the currently active effect in the sequence.
	elapsed := time.Since(w.startTime)
	if w.loop {
		if td := totalDuration(w.effects); td > 0 {
			elapsed = elapsed % td
		}
	}
	active, effectElapsed := findEffect(w.effects, elapsed)

	if active == nil {
		// All effects done — pass the inner widget's canvas through unchanged.
		return copyCanvas(w.tmp, cvs)
	}

	return applyEffect(active, effectElapsed, w.tmp, cvs)
}

// Keyboard implements widgetapi.Widget.
func (w *EffectWidget) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return w.inner.Keyboard(k, meta)
}

// Mouse implements widgetapi.Widget.
func (w *EffectWidget) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return w.inner.Mouse(m, meta)
}

// Options implements widgetapi.Widget.
func (w *EffectWidget) Options() widgetapi.Options {
	return w.inner.Options()
}

// ── internal helpers ──────────────────────────────────────────────────────────

// totalDuration sums the durations of all effects.
func totalDuration(effects []Effect) time.Duration {
	var d time.Duration
	for _, e := range effects {
		d += e.Duration
	}
	return d
}

// findEffect walks the effect sequence and returns the currently active effect
// and the time elapsed within that effect.  Returns (nil, 0) when all effects
// have finished.
func findEffect(effects []Effect, elapsed time.Duration) (*Effect, time.Duration) {
	offset := time.Duration(0)
	for i := range effects {
		e := &effects[i]
		end := offset + e.Duration
		if elapsed < end {
			return e, elapsed - offset
		}
		offset = end
	}
	return nil, 0
}

// copyCanvas copies every cell from src to dst.
// Both canvases must have the same zero-based Area().
func copyCanvas(src, dst *canvas.Canvas) error {
	ar := src.Area()
	for y := ar.Min.Y; y < ar.Max.Y; y++ {
		for x := ar.Min.X; x < ar.Max.X; x++ {
			p := image.Point{X: x, Y: y}
			c, err := src.Cell(p)
			if err != nil {
				continue
			}
			// *cell.Options implements cell.Option via its Set method.
			_, _ = dst.SetCell(p, c.Rune, c.Opts)
		}
	}
	return nil
}

// applyEffect calls the effect's CellFunc for every cell and writes the result
// to dst.  A nil return from CellFunc renders a blank cell with default colors.
func applyEffect(e *Effect, elapsed time.Duration, src, dst *canvas.Canvas) error {
	ar := src.Area()
	w, h := ar.Dx(), ar.Dy()
	for y := ar.Min.Y; y < ar.Max.Y; y++ {
		for x := ar.Min.X; x < ar.Max.X; x++ {
			p := image.Point{X: x, Y: y}
			srcCell, err := src.Cell(p)
			if err != nil {
				continue
			}

			result := e.Fn(elapsed, e.Duration, x, y, w, h, srcCell)

			if result == nil {
				_, _ = dst.SetCell(p, ' ',
					cell.FgColor(cell.ColorDefault),
					cell.BgColor(cell.ColorDefault),
				)
			} else {
				_, _ = dst.SetCell(p, result.Rune, result.Opts)
			}
		}
	}
	return nil
}
