// Package faketerm is a fake implementation of the terminal for the use in tests.
package faketerm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"log"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/cell"
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

// Terminal is a fake terminal.
// This implementation is thread-safe.
type Terminal struct {
	// buffer holds the terminal cells.
	buffer cell.Buffer
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
		log.Fatalf("New => unexpected error: %v", err)
	}
	return ft
}

// BackBuffer returns the back buffer of the fake terminal.
func (t *Terminal) BackBuffer() cell.Buffer {
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
			if r == 0 {
				r = ' '
			}
			b.WriteRune(r)
		}
		b.WriteRune('\n')
	}
	return b.String()
}

// Implements terminalapi.Terminal.Size.
func (t *Terminal) Size() image.Point {
	return t.buffer.Size()
}

// Area returns the area of the fake terminal.
func (t *Terminal) Area() image.Rectangle {
	s := t.Size()
	return image.Rect(0, 0, s.X, s.Y)
}

// Implements terminalapi.Terminal.Clear.
func (t *Terminal) Clear(opts ...cell.Option) error {
	b, err := cell.NewBuffer(t.buffer.Size())
	if err != nil {
		return err
	}
	t.buffer = b
	return nil
}

// Implements terminalapi.Terminal.Flush.
func (t *Terminal) Flush() error {
	return errors.New("unimplemented")
}

// Implements terminalapi.Terminal.SetCursor.
func (t *Terminal) SetCursor(p image.Point) {
	log.Fatal("unimplemented")
}

// Implements terminalapi.Terminal.HideCursor.
func (t *Terminal) HideCursor() {
	log.Fatal("unimplemented")
}

// Implements terminalapi.Terminal.SetCell.
func (t *Terminal) SetCell(p image.Point, r rune, opts ...cell.Option) error {
	ar, err := area.FromSize(t.buffer.Size())
	if err != nil {
		return err
	}
	if !p.In(ar) {
		return fmt.Errorf("cell at point %+v falls out of the terminal area %+v", p, ar)
	}

	cell := t.buffer[p.X][p.Y]
	cell.Rune = r
	cell.Apply(opts...)
	return nil
}

// Implements terminalapi.Terminal.Event.
func (t *Terminal) Event(ctx context.Context) terminalapi.Event {
	log.Fatal("unimplemented")
	return nil
}

// Closes the terminal. This is a no-op on the fake terminal.
func (t *Terminal) Close() {}
