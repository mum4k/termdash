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

// Package fakewidget implements a fake widget that is useful for testing the
// termdash infrastructure.
package fakewidget

import (
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// outputLines are the number of lines written by this plugin.
const outputLines = 3

const (
	sizeLine = iota
	keyboardLine
	mouseLine
)

// MinimumSize is the minimum size required to draw this widget.
var MinimumSize = image.Point{24, 5}

// Mirror is a fake widget. The fake widget draws a border around its assigned
// canvas and writes the size of its assigned canvas on the first line of the
// canvas. It writes the last received keyboard event onto the second line. It
// writes the last received mouse event onto the third line. If a non-empty
// string is provided via the Text() method, that text will be written right
// after the canvas size on the first line.
//
// The widget requests the same options that are provided to the constructor.
// If the options or canvas size don't allow for the three lines mentioned
// above, the widget skips the ones it has no space for.
//
// This is thread-safe and must not be copied.
// Implements widgetapi.Widget.
type Mirror struct {
	// lines are the three lines that will be drawn on the canvas.
	lines []string

	// text is the text provided by the last call to Text().
	text string

	// mu protects lines.
	mu sync.RWMutex

	// opts options for this widget.
	opts widgetapi.Options
}

// New returns a new fake widget.
// The widget will return the provided options on a call to Options().
func New(opts widgetapi.Options) *Mirror {
	return &Mirror{
		lines: make([]string, outputLines),
		opts:  opts,
	}
}

// Draw draws up to there lines on the canvas, assuming there is space for
// them. Returns an error if the canvas is so small that it cannot even draw a
// 2x2 border on it, or of any of the text lines end up being longer than the
// width of the canvas.
// Draw implements widgetapi.Widget.Draw.
func (mi *Mirror) Draw(cvs *canvas.Canvas) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	if err := cvs.Clear(); err != nil {
		return err
	}
	if err := draw.Border(cvs, cvs.Area()); err != nil {
		return err
	}

	mi.lines[sizeLine] = fmt.Sprintf("%s%s", cvs.Size().String(), mi.text)
	usable := area.ExcludeBorder(cvs.Area())
	start := cvs.Area().Intersect(usable).Min
	for i := 0; i < outputLines; i++ {
		if i >= usable.Dy() {
			break
		}

		if err := draw.Text(cvs, mi.lines[i], start, draw.TextMaxX(usable.Max.X)); err != nil {
			return err
		}
		start = image.Point{start.X, start.Y + 1}
	}

	return nil
}

// Text stores a text that should be displayed right after the canvas size on
// the first line of the output.
func (mi *Mirror) Text(txt string) {
	mi.text = txt
}

// Keyboard draws the received key on the canvas.
// Sending the keyboard.KeyEsc causes this widget to forget the last keyboard
// event and return an error instead.
// Keyboard implements widgetapi.Widget.Keyboard.
func (mi *Mirror) Keyboard(k *terminalapi.Keyboard) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	if k.Key == keyboard.KeyEsc {
		mi.lines[keyboardLine] = ""
		return fmt.Errorf("fakewidget received keyboard event: %v", k)
	}
	mi.lines[keyboardLine] = k.Key.String()
	return nil
}

// Mouse draws the canvas coordinates of the mouse event and the name of the
// received mouse button on the canvas.
// Sending the mouse.ButtonRight causes this widget to forget the last mouse
// event and return an error instead.
// Mouse implements widgetapi.Widget.Mouse.
func (mi *Mirror) Mouse(m *terminalapi.Mouse) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	if m.Button == mouse.ButtonRight {
		mi.lines[mouseLine] = ""
		return fmt.Errorf("fakewidget received mouse event: %v", m)
	}
	mi.lines[mouseLine] = fmt.Sprintf("%v%v", m.Position, m.Button)
	return nil
}

// Options implements widgetapi.Widget.Options.
func (mi *Mirror) Options() widgetapi.Options {
	return mi.opts
}

// Draw draws the content that would be expected after placing the Mirror
// widget onto the provided canvas and forwarding the given events.
func Draw(t terminalapi.Terminal, cvs *canvas.Canvas, opts widgetapi.Options, events ...terminalapi.Event) error {
	mirror := New(opts)
	return DrawWithMirror(mirror, t, cvs, events...)
}

// MustDraw is like Draw, but panics on all errors.
func MustDraw(t terminalapi.Terminal, cvs *canvas.Canvas, opts widgetapi.Options, events ...terminalapi.Event) {
	if err := Draw(t, cvs, opts, events...); err != nil {
		panic(fmt.Sprintf("Draw => %v", err))
	}
}

// DrawWithMirror is like Draw, but uses the provided Mirror instead of creating one.
func DrawWithMirror(mirror *Mirror, t terminalapi.Terminal, cvs *canvas.Canvas, events ...terminalapi.Event) error {
	for _, ev := range events {
		switch e := ev.(type) {
		case *terminalapi.Mouse:
			if mirror.opts.WantMouse == widgetapi.MouseScopeNone {
				continue
			}
			if err := mirror.Mouse(e); err != nil {
				return err
			}
		case *terminalapi.Keyboard:
			if mirror.opts.WantKeyboard == widgetapi.KeyScopeNone {
				continue
			}
			if err := mirror.Keyboard(e); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported event type %T", e)
		}
	}

	if err := mirror.Draw(cvs); err != nil {
		return err
	}
	return cvs.Apply(t)
}

// MustDrawWithMirror is like DrawWithMirror, but panics on all errors.
func MustDrawWithMirror(mirror *Mirror, t terminalapi.Terminal, cvs *canvas.Canvas, events ...terminalapi.Event) {
	if err := DrawWithMirror(mirror, t, cvs, events...); err != nil {
		panic(fmt.Sprintf("DrawWithMirror => %v", err))
	}
}
