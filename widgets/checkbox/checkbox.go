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

// Package checkbox implements an interactive checkbox widget.
package checkbox

import (
	"image"
	"strings"
	"sync"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Checkbox lets the user toggle a single boolean value.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Checkbox struct {
	mu sync.Mutex

	label   string
	checked bool

	opts *options
}

// New returns a new Checkbox with the provided label.
func New(label string, opts ...Option) (*Checkbox, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}

	return &Checkbox{
		label:   label,
		checked: opt.checked,
		opts:    opt,
	}, nil
}

// Checked reports whether the checkbox is currently checked.
func (c *Checkbox) Checked() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.checked
}

// SetChecked replaces the current checked state.
func (c *Checkbox) SetChecked(checked bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checked = checked
}

// Toggle flips the current checked state and returns the new value.
func (c *Checkbox) Toggle() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checked = !c.checked
	return c.checked
}

// Draw draws the Checkbox onto the canvas.
// Implements widgetapi.Widget.Draw.
func (c *Checkbox) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	need := image.Point{X: widthFor(c.opts, c.label), Y: 1}
	if need.X > cvs.Area().Dx() || need.Y > cvs.Area().Dy() {
		return draw.ResizeNeeded(cvs)
	}

	cellOpts := c.opts.cellOpts
	if c.checked {
		cellOpts = c.opts.checkedCellOpts
	} else if meta.Focused && len(c.opts.focusedCellOpts) > 0 {
		cellOpts = c.opts.focusedCellOpts
	}

	return draw.Text(cvs, textFor(c.opts, c.label, c.checked), image.Point{}, draw.TextCellOpts(cellOpts...))
}

// Keyboard processes keyboard events for the checkbox.
// Implements widgetapi.Widget.Keyboard.
func (c *Checkbox) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_ = meta
	switch k.Key {
	case keyboard.KeyEnter, ' ':
		return c.trigger()
	default:
		return nil
	}
}

// Mouse processes mouse events for the checkbox.
// Implements widgetapi.Widget.Mouse.
func (c *Checkbox) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	_ = meta
	if m.Button != mouse.ButtonLeft {
		return nil
	}
	if m.Position.Y != 0 || m.Position.X < 0 || m.Position.X >= widthFor(c.opts, c.label) {
		return nil
	}
	return c.trigger()
}

// Options implements widgetapi.Widget.Options.
func (c *Checkbox) Options() widgetapi.Options {
	c.mu.Lock()
	defer c.mu.Unlock()

	return widgetapi.Options{
		MinimumSize:  image.Point{X: widthFor(c.opts, c.label), Y: 1},
		WantKeyboard: widgetapi.KeyScopeFocused,
		WantMouse:    widgetapi.MouseScopeWidget,
	}
}

// Text returns the checkbox's current rendered one-line text.
func (c *Checkbox) Text() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return textFor(c.opts, c.label, c.checked)
}

// TextFor returns the checkbox's rendered text for the provided checked state.
func (c *Checkbox) TextFor(checked bool) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return textFor(c.opts, c.label, checked)
}

// trigger flips the checkbox state and calls the change hook, if configured.
func (c *Checkbox) trigger() error {
	c.mu.Lock()
	c.checked = !c.checked
	checked := c.checked
	callback := c.opts.onChange
	c.mu.Unlock()

	if callback != nil {
		return callback(checked)
	}
	return nil
}

// textFor returns the one-line checkbox text for the provided state.
func textFor(opts *options, label string, checked bool) string {
	box := opts.indicator.Unchecked
	if checked {
		box = opts.indicator.Checked
	}
	if label == "" {
		return box
	}
	return box + strings.Repeat(" ", opts.labelGap) + label
}

// widthFor returns the maximum rendered width across checked states.
func widthFor(opts *options, label string) int {
	width := runewidth.StringWidth(textFor(opts, label, false))
	if checkedWidth := runewidth.StringWidth(textFor(opts, label, true)); checkedWidth > width {
		width = checkedWidth
	}
	return width
}
