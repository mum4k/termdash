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
	"fmt"
	"image"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/alignfor"
	"github.com/mum4k/termdash/private/attrrange"
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

// TextChunk is a part of or the full text displayed in the button.
type TextChunk struct {
	text  string
	tOpts *textOptions
}

// NewChunk creates a new text chunk. Each chunk of text can have its own cell options.
func NewChunk(text string, tOpts ...TextOption) *TextChunk {
	return &TextChunk{
		text:  text,
		tOpts: newTextOptions(tOpts...),
	}
}

// Button can be pressed using a mouse click or a configured keyboard key.
//
// Upon each press, the button invokes a callback provided by the user.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Button struct {
	// text in the text label displayed in the button.
	text strings.Builder

	// givenTOpts are text options given for the button's of text.
	givenTOpts []*textOptions
	// tOptsTracker tracks the positions in a text to which the givenTOpts apply.
	tOptsTracker *attrrange.Tracker

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
// The callback function can be nil in which case pressing the button is a
// no-op.
func New(text string, cFn CallbackFn, opts ...Option) (*Button, error) {
	return NewFromChunks([]*TextChunk{NewChunk(text)}, cFn, opts...)
}

// NewFromChunks is like New, but allows specifying write options for
// individual chunks of text displayed in the button.
func NewFromChunks(chunks []*TextChunk, cFn CallbackFn, opts ...Option) (*Button, error) {
	if len(chunks) == 0 {
		return nil, errors.New("at least one text chunk must be specified")
	}

	var (
		text       strings.Builder
		givenTOpts []*textOptions
	)
	tOptsTracker := attrrange.NewTracker()
	for i, tc := range chunks {
		if tc.text == "" {
			return nil, fmt.Errorf("text chunk[%d] is empty, all chunks must contains some text", i)
		}

		pos := text.Len()
		givenTOpts = append(givenTOpts, tc.tOpts)
		tOptsIdx := len(givenTOpts) - 1
		if err := tOptsTracker.Add(pos, pos+len(tc.text), tOptsIdx); err != nil {
			return nil, err
		}
		text.WriteString(tc.text)
	}

	opt := newOptions(text.String())
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}

	for _, tOpts := range givenTOpts {
		tOpts.setDefaultFgColor(opt.textColor)
	}
	return &Button{
		text:         text,
		givenTOpts:   givenTOpts,
		tOptsTracker: tOptsTracker,
		mouseFSM:     button.NewFSM(mouse.ButtonLeft, image.ZR),
		callback:     cFn,
		opts:         opt,
	}, nil
}

// SetCallback replaces the callback function of the button with the one provided.
func (b *Button) SetCallback(cFn CallbackFn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.callback = cFn
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

	sw := b.shadowWidth()
	shadowAr := image.Rect(sw, sw, cvsAr.Dx(), cvsAr.Dy())
	if !b.opts.disableShadow {
		if err := cvs.SetAreaCells(shadowAr, shadowRune, cell.BgColor(b.opts.shadowColor)); err != nil {
			return err
		}
	}

	buttonAr := image.Rect(0, 0, cvsAr.Dx()-sw, cvsAr.Dy()-sw)
	if b.state == button.Down && !b.opts.disableShadow {
		buttonAr = shadowAr
	}

	var fillColor cell.Color
	switch {
	case b.state == button.Down && b.opts.pressedFillColor != nil:
		fillColor = *b.opts.pressedFillColor
	case meta.Focused && b.opts.focusedFillColor != nil:
		fillColor = *b.opts.focusedFillColor
	default:
		fillColor = b.opts.fillColor
	}

	if err := cvs.SetAreaCells(buttonAr, buttonRune, cell.BgColor(fillColor)); err != nil {
		return err
	}
	return b.drawText(cvs, meta, buttonAr)
}

// drawText draws the text inside the button.
func (b *Button) drawText(cvs *canvas.Canvas, meta *widgetapi.Meta, buttonAr image.Rectangle) error {
	pad := b.opts.textHorizontalPadding
	textAr := image.Rect(buttonAr.Min.X+pad, buttonAr.Min.Y, buttonAr.Dx()-pad, buttonAr.Max.Y)
	start, err := alignfor.Text(textAr, b.text.String(), align.HorizontalCenter, align.VerticalMiddle)
	if err != nil {
		return err
	}

	maxCells := buttonAr.Max.X - start.X
	trimmed, err := draw.TrimText(b.text.String(), maxCells, draw.OverrunModeThreeDot)
	if err != nil {
		return err
	}

	optRange, err := b.tOptsTracker.ForPosition(0) // Text options for the current byte.
	if err != nil {
		return err
	}

	cur := start
	for i, r := range trimmed {
		if i >= optRange.High { // Get the next write options.
			or, err := b.tOptsTracker.ForPosition(i)
			if err != nil {
				return err
			}
			optRange = or
		}

		tOpts := b.givenTOpts[optRange.AttrIdx]
		var cellOpts []cell.Option
		switch {
		case b.state == button.Down && len(tOpts.pressedCellOpts) > 0:
			cellOpts = tOpts.pressedCellOpts
		case meta.Focused && len(tOpts.focusedCellOpts) > 0:
			cellOpts = tOpts.focusedCellOpts
		default:
			cellOpts = tOpts.cellOpts
		}
		cells, err := cvs.SetCell(cur, r, cellOpts...)
		if err != nil {
			return err
		}
		cur = image.Point{cur.X + cells, cur.Y}
	}
	return nil
}

// activated asserts whether the keyboard event activated the button.
func (b *Button) keyActivated(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.opts.globalKeys[k.Key] || (b.opts.focusedKeys[k.Key] && meta.Focused) {
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
func (b *Button) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	if b.keyActivated(k, meta) {
		if b.callback != nil {
			// Mutex must be released when calling the callback.
			// Users might call container methods from the callback like the
			// Container.Update, see #205.
			return b.callback()
		}
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
func (b *Button) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	if b.mouseActivated(m) {
		if b.callback != nil {
			// Mutex must be released when calling the callback.
			// Users might call container methods from the callback like the
			// Container.Update, see #205.
			return b.callback()
		}
	}
	return nil
}

// shadowWidth returns the width of the shadow under the button or zero if the
// button shouldn't have any shadow.
func (b *Button) shadowWidth() int {
	if b.opts.disableShadow {
		return 0
	}
	return 1
}

// Options implements widgetapi.Widget.Options.
func (b *Button) Options() widgetapi.Options {
	// No need to lock, as the height and width get fixed when New is called.

	width := b.opts.width + b.shadowWidth() + 2*b.opts.textHorizontalPadding
	height := b.opts.height + b.shadowWidth()

	var keyScope widgetapi.KeyScope
	if len(b.opts.focusedKeys) > 0 || len(b.opts.globalKeys) > 0 {
		keyScope = widgetapi.KeyScopeGlobal
	} else {
		keyScope = widgetapi.KeyScopeNone
	}
	return widgetapi.Options{
		MinimumSize:  image.Point{width, height},
		MaximumSize:  image.Point{width, height},
		WantKeyboard: keyScope,
		WantMouse:    widgetapi.MouseScopeGlobal,
	}
}
