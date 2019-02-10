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

// Package button implements an interactive widget that can be pressed to
// activate.
package button

import (
	"errors"
	"image"
	"sync"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// CallbackFn is the function called when the button is pressed.
// The callback function must be non-blocking, ideally just storing a value and
// returning, since event processing blocks the redraws.
//
// The callback function must be thread-safe as the mouse or keyboard events
// that press the button are processed in a separate goroutine.
//
// If the function returns an error, the widget will forward it back to the
// termdash infrastructure which causes a panic, unless the user provided a
// termdash.ErrorHandler.
type CallbackFn func() error

// Button can be pressed using a mouse click or a configured keyboard key.
//
// Upon each press, the button invokes a callback provided by the user.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Button struct {
	// text in the text label displayed in the button.
	text string

	// mu protects the widget.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new Button that will display the provided text.
// Each press of the button will invoke the callback function.
func New(text string, cFn CallbackFn, opts ...Option) (*Button, error) {
	opt := newOptions(runewidth.StringWidth(text))
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &Button{
		text: text,
		opts: opt,
	}, nil
}

// Draw draws the Button widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (b *Button) Draw(cvs *canvas.Canvas) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return errors.New("unimplemented")
}

// Keyboard processes keyboard events, acts as a button press on the configured
// Key.
//
// Implements widgetapi.Widget.Keyboard.
func (*Button) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("unimplemented")
}

// Mouse processes mouse events, acts as a button press if both the press and
// the release happen inside the button.
//
// Implements widgetapi.Widget.Mouse.
func (*Button) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the SegmentDisplay widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (b *Button) Options() widgetapi.Options {
	// No need to lock, as the height and width get fixed when New is called.

	height := b.opts.height + 1 // One for the shadow.
	return widgetapi.Options{
		MinimumSize:  image.Point{b.opts.width, height},
		WantKeyboard: true,
		WantMouse:    true,
	}
}
