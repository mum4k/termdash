// Copyright 2020 Google Inc.
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

package tcell

import (
	"context"
	"fmt"
	"image"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/event/eventqueue"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*Terminal)
}

// option implements Option.
type option func(*Terminal)

// set implements Option.set.
func (o option) set(t *Terminal) {
	o(t)
}

// DefaultColorMode is the default value for the ColorMode option.
const DefaultColorMode = terminalapi.ColorMode256

// ColorMode sets the terminal color mode.
// Defaults to DefaultColorMode.
func ColorMode(cm terminalapi.ColorMode) Option {
	return option(func(t *Terminal) {
		t.colorMode = cm
	})
}

// ClearStyle sets the style to use for tcell when clearing the screen.
// Defaults to ColorDefault for foreground and background.
func ClearStyle(fg, bg cell.Color) Option {
	return option(func(t *Terminal) {
		t.clearStyle = &cell.Options{
			FgColor: fg,
			BgColor: bg,
		}
	})
}

// Terminal provides input and output to a real terminal. Wraps the
// gdamore/tcell terminal implementation. This object is not thread-safe.
// Implements terminalapi.Terminal.
type Terminal struct {
	// events is a queue of input events.
	events *eventqueue.Unbound

	// done gets closed when Close() is called.
	done chan struct{}

	// the tcell terminal window
	screen tcell.Screen

	// Options.
	colorMode  terminalapi.ColorMode
	clearStyle *cell.Options
}

// tcellNewScreen can be overridden from tests.
var tcellNewScreen = tcell.NewScreen

// newTerminal creates the terminal and applies the options.
func newTerminal(opts ...Option) (*Terminal, error) {
	screen, err := tcellNewScreen()
	if err != nil {
		return nil, fmt.Errorf("tcell.NewScreen => %v", err)
	}

	t := &Terminal{
		events:    eventqueue.New(),
		done:      make(chan struct{}),
		colorMode: DefaultColorMode,
		clearStyle: &cell.Options{
			FgColor: cell.ColorDefault,
			BgColor: cell.ColorDefault,
		},
		screen: screen,
	}
	for _, opt := range opts {
		opt.set(t)
	}

	return t, nil
}

// New returns a new tcell based Terminal.
// Call Close() when the terminal isn't required anymore.
func New(opts ...Option) (*Terminal, error) {
	// Enable full character set support for tcell
	encoding.Register()

	t, err := newTerminal(opts...)
	if err != nil {
		return nil, err
	}
	if err = t.screen.Init(); err != nil {
		return nil, err
	}

	clearStyle := cellOptsToStyle(t.clearStyle, t.colorMode)
	t.screen.EnableMouse()
	t.screen.SetStyle(clearStyle)

	go t.pollEvents() // Stops when Close() is called.
	return t, nil
}

// Size implements terminalapi.Terminal.Size.
func (t *Terminal) Size() image.Point {
	w, h := t.screen.Size()
	return image.Point{
		X: w,
		Y: h,
	}
}

// Clear implements terminalapi.Terminal.Clear.
func (t *Terminal) Clear(opts ...cell.Option) error {
	o := cell.NewOptions(opts...)
	st := cellOptsToStyle(o, t.colorMode)
	t.screen.Fill(' ', st)
	return nil
}

// Flush implements terminalapi.Terminal.Flush.
func (t *Terminal) Flush() error {
	t.screen.Show()
	return nil
}

// SetCursor implements terminalapi.Terminal.SetCursor.
func (t *Terminal) SetCursor(p image.Point) {
	t.screen.ShowCursor(p.X, p.Y)
}

// HideCursor implements terminalapi.Terminal.HideCursor.
func (t *Terminal) HideCursor() {
	t.screen.HideCursor()
}

// SetCell implements terminalapi.Terminal.SetCell.
func (t *Terminal) SetCell(p image.Point, r rune, opts ...cell.Option) error {
	o := cell.NewOptions(opts...)
	st := cellOptsToStyle(o, t.colorMode)
	t.screen.SetContent(p.X, p.Y, r, nil, st)
	return nil
}

// pollEvents polls and enqueues the input events.
func (t *Terminal) pollEvents() {
	for {
		select {
		case <-t.done:
			return
		default:
		}

		events := toTermdashEvents(t.screen.PollEvent())
		for _, ev := range events {
			t.events.Push(ev)
		}
	}
}

// Event implements terminalapi.Terminal.Event.
func (t *Terminal) Event(ctx context.Context) terminalapi.Event {
	ev := t.events.Pull(ctx)
	if ev == nil {
		return nil
	}
	return ev
}

// Close closes the terminal, should be called when the terminal isn't required
// anymore to return the screen to a sane state.
// Implements terminalapi.Terminal.Close.
func (t *Terminal) Close() {
	close(t.done)
	t.screen.Fini()
}
