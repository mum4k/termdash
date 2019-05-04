// Copyright 2019 Google Inc.
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

// Package table is a widget that displays text a table consisting of rows and
// columns.
package table

import (
	"errors"
	"image"
	"sync"

	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Table displays text content aligned to rows and columns.
//
// Sizes of columns and rows are automatically sized based on the content or
// specified by the caller. Each table cell supports text alignment, trimming
// or wrapping. Table data can be sorted according to values in individual
// columns.
//
// The entire table can be scrolled if it has more content that fits into the
// container. The table content can be interacted with using keyboard or mouse.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Table struct {
	// mu protects the widget.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new Table.
func New(opts ...Option) (*Table, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &Table{}, nil
}

// Reset resets the table to an empty content.
func (t *Table) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
}

// Write writes the provided content and replaces any content written
// previously.
// The caller is allowed to modify the content after Write, changes to content
// will become visible.
func (t *Table) Write(c *Content) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return errors.New("unimplemented")
}

// Draw draws the Table widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (t *Table) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return nil
}

// Keyboard processes keyboard events.
// The Table widget uses keyboard events to highlight selected rows or columns.
// Implements widgetapi.Widget.Keyboard.
func (t *Table) Keyboard(k *terminalapi.Keyboard) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return errors.New("unimplemented")
}

// Mouse processes mouse events.
// Mouse events are used to scroll the content and highlight rows or columns.
// Implements widgetapi.Widget.Mouse.
func (t *Table) Mouse(m *terminalapi.Mouse) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return errors.New("unimplemented")
}

// Options implements widgetapi.Widget.Options.
func (t *Table) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:  image.Point{1, 1},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
}
