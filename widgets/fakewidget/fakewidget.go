// Package fakewidget implements a fake widget that is useful for testing the
// termdash infrastructure.
package fakewidget

import (
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/draw"
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

// Mirror is a fake widget. The fake widget draws a box around its assigned
// canvas and writes the size of its assigned canvas on the first line of the
// canvas. It writes the last received keyboard event onto the second line.
// It writes the last received mouse event onto the third line.
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
	if err := draw.Box(cvs, cvs.Area(), draw.LineStyleLight); err != nil {
		return err
	}

	mi.lines[sizeLine] = cvs.Size().String()
	usable := area.ExcludeBorder(cvs.Area())
	start := cvs.Area().Intersect(usable).Min
	for i := 0; i < outputLines; i++ {
		if i >= usable.Dy() {
			break
		}

		tb := draw.TextBounds{
			Start:   start,
			MaxX:    usable.Max.X,
			Overrun: draw.OverrunModeStrict,
		}
		if err := draw.Text(cvs, mi.lines[i], tb); err != nil {
			return err
		}
		start = image.Point{start.X, start.Y + 1}
	}

	return nil
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
