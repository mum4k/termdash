// Copyright 2018 Google Inc.
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

// Package text contains a widget that displays textual data.
package text

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"sync"
	"unicode"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Text displays a block of text.
//
// Each line of the text is either trimmed or wrapped according to the provided
// options. The entire text content is either trimmed or rolled up through the
// canvas according to the provided options.
//
// By default the widget supports scrolling of content with either the keyboard
// or mouse. See the options for the default keys and mouse buttons.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Text struct {
	// buff contains the text to be displayed in the widget.
	buff bytes.Buffer
	// givenWOpts are write options given for the text.
	givenWOpts givenWOpts

	// scroll tracks scrolling the position.
	scroll *scrollTracker

	// lastWidth stores the width of the last canvas the widget drew on.
	// Used to determine if the previous line wrapping was invalidated.
	lastWidth int
	// contentChanged indicates if the text content of the widget changed since
	// the last drawing. Used to determine if the previous line wrapping was
	// invalidated.
	contentChanged bool
	// lines stores the starting locations in bytes of all the lines in the
	// buffer. I.e. positions of newline characters and of any calculated line wraps.
	lines []int

	// mu protects the Text widget.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new text widget.
func New(opts ...Option) *Text {
	opt := newOptions(opts...)
	return &Text{
		givenWOpts: newGivenWOpts(),
		scroll:     newScrollTracker(opt),
		opts:       opt,
	}
}

// Reset resets the widget back to empty content.
func (t *Text) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.buff.Reset()
	t.givenWOpts = newGivenWOpts()
	t.scroll = newScrollTracker(t.opts)
	t.lastWidth = 0
	t.contentChanged = true
	t.lines = nil
}

// Write writes text for the widget to display. Multiple calls append
// additional text. The text cannot control characters (unicode.IsControl) or
// space character (unicode.IsSpace) other than:
//   ' ', '\n'
// Any newline ('\n') characters are interpreted as newlines when displaying
// the text.
func (t *Text) Write(text string, wOpts ...WriteOption) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := validText(text); err != nil {
		return err
	}

	pos := t.buff.Len()
	t.givenWOpts[pos] = newOptsRange(pos, pos+len(text), newWriteOptions(wOpts...))
	if _, err := t.buff.WriteString(text); err != nil {
		return err
	}
	t.contentChanged = true
	return nil
}

// lineTrimNeeded returns true if the text on this line needs to be trimmed.
// I.e. if the Text gadget was configured to trim lines, the current point falls
// outside of the canvas and the current rune doesn't start a new line.
func (t *Text) lineTrimNeeded(cur image.Point, cvs *canvas.Canvas, r rune) bool {
	if cur.X < cvs.Area().Dx() || r == '\n' {
		return false
	}
	return !t.opts.wrapAtRunes
}

// minLinesForMarkers are the minimum amount of lines required on the canvas in
// order to draw the scroll markers ('⇧' and '⇩').
const minLinesForMarkers = 3

// draw draws the text context on the canvas starting at the specified line.
// Argument starts are the starting positions of all the lines in the text.
func (t *Text) draw(text string, cvs *canvas.Canvas, starts []int, fromLine int) error {
	var cur image.Point // Tracks the current drawing position on the canvas.
	lines := len(starts)
	height := cvs.Area().Dy()
	optRange := t.givenWOpts.forPosition(0) // Text options for the current byte.
	startPos := starts[fromLine]
	for i, r := range text {
		if i < startPos {
			continue
		}

		// Draw the scroll up marker on the first line if there is more text
		// above the canvas.
		if cur.Y == 0 && height >= minLinesForMarkers && fromLine > 0 {
			if err := cvs.SetCell(cur, '⇧'); err != nil {
				return err
			}
			cur = image.Point{0, cur.Y + 1} // Move to the next line.
			startPos = starts[fromLine+1]   // Skip one line of text, the marker replaced it.
			continue
		}

		if r == '\n' || wrapNeeded(r, cur.X, cvs.Area().Dx(), t.opts) {
			cur = image.Point{0, cur.Y + 1} // Move to the next line.
		}

		// Draw the scroll down marker on the last line if there is more text
		// below the canvas.
		if cur.Y == height-1 && height >= minLinesForMarkers && height < lines-fromLine {
			if height >= minLinesForMarkers {
				if err := cvs.SetCell(cur, '⇩'); err != nil {
					return err
				}
			}
			break
		}

		if r == '\n' {
			continue // Don't print the newline runes, just interpret them above.
		}

		if t.lineTrimNeeded(cur, cvs, r) {
			// Trim by replacing the last printed rune.
			prev := image.Point{cur.X - 1, cur.Y}
			if prev.In(cvs.Area()) {
				if err := cvs.SetCell(prev, '…'); err != nil {
					return err
				}
			}
		}

		if !cur.In(cvs.Area()) {
			continue // Skip any runes belonging to the trimmed area on a line.
		}

		if i >= optRange.high { // Get the next write options.
			optRange = t.givenWOpts.forPosition(i)
		}
		if err := cvs.SetCell(cur, r, optRange.opts.cellOpts); err != nil {
			return err
		}
		cur = image.Point{cur.X + 1, cur.Y} // Move within the same line.
	}
	return nil
}

// Draw draws the text onto the canvas.
// Implements widgetapi.Widget.Draw.
func (t *Text) Draw(cvs *canvas.Canvas) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	text := t.buff.String()
	width := cvs.Area().Dx()
	if t.contentChanged || t.lastWidth != width {
		// The previous text preprocessing (line wrapping) is invalidated when
		// new text is added or the width of the canvas changed.
		t.lines = findLines(text, width, t.opts)
	}
	t.lastWidth = width

	if len(t.lines) == 0 {
		return nil // Nothing to draw if there's no text.
	}

	height := cvs.Area().Dy()
	fromLine := t.scroll.firstLine(len(t.lines), height)
	if err := t.draw(text, cvs, t.lines, fromLine); err != nil {
		return err
	}
	t.contentChanged = false
	return nil
}

// Implements widgetapi.Widget.Keyboard.
func (t *Text) Keyboard(k *terminalapi.Keyboard) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch {
	case k.Key == t.opts.keyUp:
		t.scroll.upOneLine()
	case k.Key == t.opts.keyDown:
		t.scroll.downOneLine()
	case k.Key == t.opts.keyPgUp:
		t.scroll.upOnePage()
	case k.Key == t.opts.keyPgDown:
		t.scroll.downOnePage()
	}
	return nil
}

// Implements widgetapi.Widget.Mouse.
func (t *Text) Mouse(m *terminalapi.Mouse) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch b := m.Button; {
	case b == t.opts.mouseUpButton:
		t.scroll.upOneLine()
	case b == t.opts.mouseDownButton:
		t.scroll.downOneLine()
	}
	return nil
}

func (t *Text) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:  image.Point{1, 1},
		WantMouse:    !t.opts.disableScrolling,
		WantKeyboard: !t.opts.disableScrolling,
	}
}

// validText validates the provided text.
func validText(text string) error {
	if text == "" {
		return errors.New("the text cannot be empty")
	}

	for _, c := range text {
		if c == ' ' || c == '\n' { // Allowed space and control runes.
			continue
		}
		if unicode.IsControl(c) {
			return fmt.Errorf("the provided text %q cannot contain control characters, found: %q", text, c)
		}
		if unicode.IsSpace(c) {
			return fmt.Errorf("the provided text %q cannot contain space character %q", text, c)
		}
	}
	return nil
}
