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

// Package textinput implements a widget that accepts text input.
package textinput

import (
	"sync"

	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// FilterFn if provided can be used to filter runes that are allowed in the
// text input field. Any rune for which this function returns false will be
// rejected.
type FilterFn func(rune) bool

// SubmitFn if provided is called when the user submits the content of the text
// input field, the argument text contains all the text in the field.
// Submitting the input field clears its content.
//
// The callback function must be thread-safe as the keyboard event that
// triggers the submission comes from a separate goroutine.
type SubmitFn func(text string) error

// TextInput accepts text input from the user.
//
// Displays an input field where the user can edit text and an optional label.
//
// The text can be submitted by pressing enter or read at any time by calling
// Read.
//
// Implements widgetapi.Widget. This object is thread-safe.
type TextInput struct {
	// mu protects the widget.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new TextInput.
func New(opts ...Option) (*TextInput, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &TextInput{
		opts: opt,
	}, nil
}

// Vars to be replaced from tests.
var (
	// Runes to use in cells that contain are reserved for the text input
	// field if no text is present.
	// Changed from tests to provide readable test failures.
	textFieldRune = ' '
)

// Draw draws the TextInput widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (ti *TextInput) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	return nil
}

// Keyboard processes keyboard events.
// Implements widgetapi.Widget.Keyboard.
func (ti *TextInput) Keyboard(k *terminalapi.Keyboard) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	return nil
}

// Mouse processes mouse events.
// Implements widgetapi.Widget.Mouse.
func (ti *TextInput) Mouse(m *terminalapi.Mouse) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	return nil
}

// Options implements widgetapi.Widget.Options.
func (ti *TextInput) Options() widgetapi.Options {
	return widgetapi.Options{}
}
