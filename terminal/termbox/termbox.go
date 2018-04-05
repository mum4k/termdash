// Package termbox implements terminal using the nsf/termbox-go library.
package termbox

import (
	"context"
	"image"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/eventqueue"
	"github.com/mum4k/termdash/terminalapi"
	tbx "github.com/nsf/termbox-go"
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

// ColorMode sets the terminal color mode.
func ColorMode(cm terminalapi.ColorMode) Option {
	return option(func(t *Terminal) {
		t.colorMode = cm
	})
}

// Terminal provides input and output to a real terminal.
// Wraps the nsf/termbox-go terminal implementation.
// This isn't thread-safe, because termbox isn't and only one instance is ever
// supported, because termbox uses global state.
// Implements terminalapi.Terminal.
type Terminal struct {
	// events is a queue of input events.
	events *eventqueue.Unbound

	// done gets closed when Close() is called.
	done chan struct{}

	// Options.
	colorMode terminalapi.ColorMode
}

// New returns a new termbox based Terminal.
// Call Close() when the terminal isn't required anymore.
func New(opts ...Option) (*Terminal, error) {
	if err := tbx.Init(); err != nil {
		return nil, err
	}
	tbx.SetInputMode(tbx.InputEsc | tbx.InputMouse)

	t := &Terminal{
		events: eventqueue.New(),
		done:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt.set(t)
	}

	om, err := colorMode(t.colorMode)
	if err != nil {
		return nil, err
	}
	tbx.SetOutputMode(om)

	go t.pollEvents() // Stops when Close() is called.
	return t, nil
}

// Implements terminalapi.Terminal.Size.
func (t *Terminal) Size() image.Point {
	w, h := tbx.Size()
	return image.Point{w, h}
}

// Implements terminalapi.Terminal.Clear.
func (t *Terminal) Clear(opts ...cell.Option) error {
	o := cell.NewOptions(opts...)
	fg, err := cellOptsToFg(o)
	if err != nil {
		return err
	}

	bg, err := cellOptsToBg(o)
	if err != nil {
		return err
	}
	return tbx.Clear(fg, bg)
}

// Implements terminalapi.Terminal.Flush.
func (t *Terminal) Flush() error {
	return tbx.Flush()
}

// Implements terminalapi.Terminal.SetCursor.
func (t *Terminal) SetCursor(p image.Point) {
	tbx.SetCursor(p.X, p.Y)
}

// Implements terminalapi.Terminal.HideCursor.
func (t *Terminal) HideCursor() {
	tbx.HideCursor()
}

// Implements terminalapi.Terminal.SetCell.
func (t *Terminal) SetCell(p image.Point, r rune, opts ...cell.Option) error {
	o := cell.NewOptions(opts...)
	fg, err := cellOptsToFg(o)
	if err != nil {
		return err
	}

	bg, err := cellOptsToBg(o)
	if err != nil {
		return err
	}
	tbx.SetCell(p.X, p.Y, r, fg, bg)
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

		events := toTermdashEvents(tbx.PollEvent())
		for _, ev := range events {
			t.events.Push(ev)
		}
	}
}

// Implements terminalapi.Terminal.Event.
func (t *Terminal) Event(ctx context.Context) terminalapi.Event {
	ev, err := t.events.Pull(ctx)
	if err != nil {
		return terminalapi.NewErrorf("unable to pull the next event: %v", err)
	}
	return ev
}

// Closes the terminal, should be called when the terminal isn't required
// anymore to return the screen to a sane state.
func (t *Terminal) Close() {
	close(t.done)
	tbx.Close()
}
