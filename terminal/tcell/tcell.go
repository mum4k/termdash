package tcell

import (
	"context"
	"image"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/event/eventqueue"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// Terminal provides input and output to a real terminal. Wraps the
// gdamore/tcell terminal implementation. This object is not thread-safe.
// Implements terminalapi.Terminal.
type Terminal struct {
	// events is a queue of input events.
	events *eventqueue.Unbound

	// done gets closed when Close() is called.
	done chan struct{}

	screen tcell.Screen
}

// New returns a new tcell based Terminal.
// Call Close() when the terminal isn't required anymore.
func New() (*Terminal, error) {
	// Enable full character set support for tcell
	encoding.Register()

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	t := &Terminal{
		events: eventqueue.New(),
		done:   make(chan struct{}),
		screen: screen,
	}

	if err = t.screen.Init(); err != nil {
		return nil, err
	}

	defaultStyle := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack)

	t.screen.EnableMouse()
	t.screen.SetStyle(defaultStyle)

	go t.pollEvents()
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
	st := cellOptsToStyle(o)
	w, h := t.screen.Size()
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			t.screen.SetContent(col, row, ' ', nil, st)
		}
	}
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
	st := cellOptsToStyle(o)
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
