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

// Package slider implements an interactive horizontal value slider widget.
package slider

import (
	"image"
	"sync"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Slider lets the user select a numeric value within a fixed range.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Slider struct {
	mu sync.Mutex

	value int
	opts  *options
}

// New returns a new Slider.
func New(opts ...Option) (*Slider, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	s := &Slider{
		value: opt.value,
		opts:  opt,
	}
	s.value = s.clamp(s.value)
	return s, nil
}

// Value returns the current slider value.
func (s *Slider) Value() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}

// SetValue replaces the current value, clamping it to the configured range.
func (s *Slider) SetValue(v int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = s.clamp(v)
}

// Draw draws the Slider onto the canvas.
// Implements widgetapi.Widget.Draw.
func (s *Slider) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	need := image.Point{X: s.opts.width, Y: 1}
	if need.X > cvs.Area().Dx() || need.Y > cvs.Area().Dy() {
		return draw.ResizeNeeded(cvs)
	}

	knob := s.knobIndexLocked(s.value)
	for i := 0; i < s.opts.width; i++ {
		r := s.opts.trackRune
		cellOpts := s.opts.trackCellOpts
		switch {
		case i < knob:
			r = s.opts.fillRune
			cellOpts = s.opts.fillCellOpts
		case i == knob:
			r = s.opts.knobRune
			cellOpts = s.opts.knobCellOpts
			if meta.Focused && len(s.opts.focusedKnobOps) > 0 {
				cellOpts = s.opts.focusedKnobOps
			}
		}
		if _, err := cvs.SetCell(image.Point{X: i, Y: 0}, r, cellOpts...); err != nil {
			return err
		}
	}
	return nil
}

// Keyboard processes keyboard events for the slider.
// Implements widgetapi.Widget.Keyboard.
func (s *Slider) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_ = meta
	switch k.Key {
	case keyboard.KeyArrowLeft:
		return s.changeBy(-s.opts.step)
	case keyboard.KeyArrowRight:
		return s.changeBy(s.opts.step)
	case keyboard.KeyHome:
		return s.setAndNotify(s.opts.min)
	case keyboard.KeyEnd:
		return s.setAndNotify(s.opts.max)
	default:
		return nil
	}
}

// Mouse processes mouse events for the slider.
// Implements widgetapi.Widget.Mouse.
func (s *Slider) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	_ = meta
	if m.Button != mouse.ButtonLeft {
		return nil
	}
	if m.Position.Y != 0 || m.Position.X < 0 || m.Position.X >= s.opts.width {
		return nil
	}
	return s.setAndNotify(s.valueAtX(m.Position.X))
}

// Options implements widgetapi.Widget.Options.
func (s *Slider) Options() widgetapi.Options {
	s.mu.Lock()
	defer s.mu.Unlock()
	return widgetapi.Options{
		MinimumSize:  image.Point{X: s.opts.width, Y: 1},
		WantKeyboard: widgetapi.KeyScopeFocused,
		WantMouse:    widgetapi.MouseScopeWidget,
	}
}

func (s *Slider) changeBy(delta int) error {
	return s.setAndNotify(s.Value() + delta)
}

func (s *Slider) setAndNotify(next int) error {
	s.mu.Lock()
	next = s.clamp(next)
	if next == s.value {
		s.mu.Unlock()
		return nil
	}
	s.value = next
	callback := s.opts.onChange
	s.mu.Unlock()

	if callback != nil {
		return callback(next)
	}
	return nil
}

func (s *Slider) clamp(v int) int {
	if v < s.opts.min {
		return s.opts.min
	}
	if v > s.opts.max {
		return s.opts.max
	}
	return v
}

func (s *Slider) knobIndexLocked(v int) int {
	if s.opts.width <= 1 || s.opts.max == s.opts.min {
		return 0
	}
	return (s.clamp(v) - s.opts.min) * (s.opts.width - 1) / (s.opts.max - s.opts.min)
}

func (s *Slider) valueAtX(x int) int {
	if x <= 0 || s.opts.width <= 1 {
		return s.opts.min
	}
	if x >= s.opts.width-1 {
		return s.opts.max
	}
	return s.opts.min + (x*(s.opts.max-s.opts.min))/(s.opts.width-1)
}
