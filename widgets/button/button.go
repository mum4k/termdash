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
	"time"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/alignfor"
	"github.com/mum4k/termdash/private/button"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// CallbackFn is the function called when the button is pressed.
// The callback function must be light-weight, ideally just storing a value and
// returning, since more button presses might occur.
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

	// mouseFSM tracks left mouse clicks.
	mouseFSM *button.FSM
	// state is the current state of the button.
	state button.State

	// keyTriggerTime is the last time the button was pressed using a keyboard
	// key. It is nil if the button was triggered by a mouse event.
	// Used to draw button presses on keyboard events, since termbox doesn't
	// provide us with release events for keys.
	keyTriggerTime *time.Time

	// callback gets called on each button press.
	callback CallbackFn

	// mu protects the widget.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new Button that will display the provided text.
// Each press of the button will invoke the callback function.
func New(text string, cFn CallbackFn, opts ...Option) (*Button, error) {
	if cFn == nil {
		return nil, errors.New("the CallbackFn argument cannot be nil")
	}

	opt := newOptions(text)
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &Button{
		text:     text,
		mouseFSM: button.NewFSM(mouse.ButtonLeft, image.ZR),
		callback: cFn,
		opts:     opt,
	}, nil
}

// Vars to be replaced from tests.
var (
	// Runes to use in cells that contain the button.
	// Changed from tests to provide readable test failures.
	buttonRune = ' '
	// Runes to use in cells that contain the shadow.
	// Changed from tests to provide readable test failures.
	shadowRune = ' '

	// timeSince is a function that calculates duration since some time.
	timeSince = time.Since
)

// Draw draws the Button widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (b *Button) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyTriggerTime != nil {
		since := timeSince(*b.keyTriggerTime)
		if since > b.opts.keyUpDelay {
			b.state = button.Up
		}
	}

	cvsAr := cvs.Area()
	b.mouseFSM.UpdateArea(cvsAr)

	shadowAr := image.Rect(shadowWidth, shadowWidth, cvsAr.Dx(), cvsAr.Dy())
	if err := cvs.SetAreaCells(shadowAr, shadowRune, cell.BgColor(b.opts.shadowColor)); err != nil {
		return err
	}

	var buttonAr image.Rectangle
	if b.state == button.Up {
		buttonAr = image.Rect(0, 0, cvsAr.Dx()-shadowWidth, cvsAr.Dy()-shadowWidth)
	} else {
		buttonAr = shadowAr
	}

	if err := cvs.SetAreaCells(buttonAr, buttonRune, cell.BgColor(b.opts.fillColor)); err != nil {
		return err
	}

	textAr := image.Rect(buttonAr.Min.X+1, buttonAr.Min.Y, buttonAr.Dx()-1, buttonAr.Max.Y)
	start, err := alignfor.Text(textAr, b.text, align.HorizontalCenter, align.VerticalMiddle)
	if err != nil {
		return err
	}
	return draw.Text(cvs, b.text, start,
		draw.TextOverrunMode(draw.OverrunModeThreeDot),
		draw.TextMaxX(buttonAr.Max.X),
		draw.TextCellOpts(cell.FgColor(b.opts.textColor)),
	)
}

// activated asserts whether the keyboard event activated the button.
func (b *Button) keyActivated(k *terminalapi.Keyboard) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if k.Key == b.opts.key {
		b.state = button.Down
		now := time.Now().UTC()
		b.keyTriggerTime = &now
		return true
	}
	return false
}

// Keyboard processes keyboard events, acts as a button press on the configured
// Key.
//
// Implements widgetapi.Widget.Keyboard.
func (b *Button) Keyboard(k *terminalapi.Keyboard) error {
	if b.keyActivated(k) {
		// Mutex must be released when calling the callback.
		// Users might call container methods from the callback like the
		// Container.Update, see #205.
		return b.callback()
	}
	return nil
}

// mouseActivated asserts whether the mouse event activated the button.
func (b *Button) mouseActivated(m *terminalapi.Mouse) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	clicked, state := b.mouseFSM.Event(m)
	b.state = state
	b.keyTriggerTime = nil

	return clicked
}

// Mouse processes mouse events, acts as a button press if both the press and
// the release happen inside the button.
//
// Implements widgetapi.Widget.Mouse.
func (b *Button) Mouse(m *terminalapi.Mouse) error {
	if b.mouseActivated(m) {
		// Mutex must be released when calling the callback.
		// Users might call container methods from the callback like the
		// Container.Update, see #205.
		return b.callback()
	}
	return nil
}

// shadowWidth is the width of the shadow under the button in cell.
const shadowWidth = 1

// Options implements widgetapi.Widget.Options.
func (b *Button) Options() widgetapi.Options {
	// No need to lock, as the height and width get fixed when New is called.

	width := b.opts.width + shadowWidth
	height := b.opts.height + shadowWidth
	return widgetapi.Options{
		MinimumSize:  image.Point{width, height},
		MaximumSize:  image.Point{width, height},
		WantKeyboard: b.opts.keyScope,
		WantMouse:    widgetapi.MouseScopeGlobal,
	}
}
