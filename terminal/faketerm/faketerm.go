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

// Package faketerm is a fake implementation of the terminal for the use in tests.
package faketerm

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/event/eventqueue"
	"github.com/mum4k/termdash/terminalapi"
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

// WithEventQueue provides a queue of events.
// One event will be consumed from the queue each time Event() is called. If
// not provided, Event() returns an error on each call.
func WithEventQueue(eq *eventqueue.Unbound) Option {
	return option(func(t *Terminal) {
		t.events = eq
	})
}

// Terminal is a fake terminal.
// This implementation is thread-safe.
type Terminal struct {
	// buffer holds the terminal cells.
	buffer cell.Buffer

	// events is a queue of input events.
	events *eventqueue.Unbound

	// mu protects the buffer.
	mu sync.Mutex
}

// New returns a new fake Terminal.
func New(size image.Point, opts ...Option) (*Terminal, error) {
	b, err := cell.NewBuffer(size)
	if err != nil {
		return nil, err
	}

	t := &Terminal{
		buffer: b,
	}
	for _, opt := range opts {
		opt.set(t)
	}
	return t, nil
}

// MustNew is like New, but panics on all errors.
func MustNew(size image.Point, opts ...Option) *Terminal {
	ft, err := New(size, opts...)
	if err != nil {
		panic(fmt.Sprintf("New => unexpected error: %v", err))
	}
	return ft
}

// Resize resizes the terminal to the provided size.
// This also clears the internal buffer.
func (t *Terminal) Resize(size image.Point) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	b, err := cell.NewBuffer(size)
	if err != nil {
		return err
	}

	t.buffer = b
	return nil
}

// BackBuffer returns the back buffer of the fake terminal.
func (t *Terminal) BackBuffer() cell.Buffer {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.buffer
}

// String prints out the buffer into a string.
// This includes the cell runes only, cell options are ignored.
// Implements fmt.Stringer.
func (t *Terminal) String() string {
	size := t.Size()
	var b bytes.Buffer
	for row := 0; row < size.Y; row++ {
		for col := 0; col < size.X; col++ {
			r := t.buffer[col][row].Rune
			p := image.Point{col, row}
			partial, err := t.buffer.IsPartial(p)
			if err != nil {
				panic(fmt.Errorf("unable to determine if point %v is a partial rune: %v", p, err))
			}
			if r == 0 && !partial {
				r = ' '
			}
			b.WriteRune(r)
		}
		b.WriteRune('\n')
	}
	return b.String()
}

// Size implements terminalapi.Terminal.Size.
func (t *Terminal) Size() image.Point {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.buffer.Size()
}

// Area returns the area of the fake terminal.
func (t *Terminal) Area() image.Rectangle {
	s := t.Size()
	return image.Rect(0, 0, s.X, s.Y)
}

// Clear implements terminalapi.Terminal.Clear.
func (t *Terminal) Clear(opts ...cell.Option) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	b, err := cell.NewBuffer(t.buffer.Size())
	if err != nil {
		return err
	}
	t.buffer = b
	return nil
}

// Flush implements terminalapi.Terminal.Flush.
func (t *Terminal) Flush() error {
	return nil // nowhere to flush to.
}

// SetCursor implements terminalapi.Terminal.SetCursor.
func (t *Terminal) SetCursor(p image.Point) {
	log.Fatal("unimplemented")
}

// HideCursor implements terminalapi.Terminal.HideCursor.
func (t *Terminal) HideCursor() {
	log.Fatal("unimplemented")
}

// SetCell implements terminalapi.Terminal.SetCell.
func (t *Terminal) SetCell(p image.Point, r rune, opts ...cell.Option) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := t.buffer.SetCell(p, r, opts...); err != nil {
		return err
	}
	return nil
}

// Event implements terminalapi.Terminal.Event.
func (t *Terminal) Event(ctx context.Context) terminalapi.Event {
	if t.events == nil {
		return terminalapi.NewErrorf("no event queue provided, use the WithEventQueue option when creating the fake terminal")
	}

	ev, err := t.events.Pull(ctx)
	if err != nil {
		return terminalapi.NewErrorf("unable to pull the next event: %v", err)
	}

	if res, ok := ev.(*terminalapi.Resize); ok {
		t.Resize(res.Size)
	}
	return ev
}

// Close closes the terminal. This is a no-op on the fake terminal.
func (t *Terminal) Close() {}
