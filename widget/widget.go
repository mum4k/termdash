// Package widget defines the API of a widget on the dashboard.
package widget

import (
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
)

// Options contains registration options for a widget.
// This is how the widget indicates its needs to the infrastructure.
type Options struct {
}

// Widget is a single widget on the dashboard.
type Widget interface {
	// Draw executes the widget, when called the widget should draw on the
	// canvas. The widget can assume that the canvas content wasn't modified
	// since the last call, i.e. if the widget doesn't need to change anything in
	// the output, this can be a no-op.
	Draw(canvas *canvas.Canvas) error

	// Redraw is called when the widget must redraw all of its content because
	// the previous canvas was invalidated. The widget must not assume that
	// anything on the canvas remained the same, including its size.
	Redraw(canvas *canvas.Canvas) error

	// Keyboard is called when the widget is focused on the dashboard and a key
	// shortcut the widget registered for was pressed. Only called if the widget
	// registered for keyboard events.
	Keyboard(s *keyboard.Shortcut) error

	// Mouse is called when the widget is focused on the dashboard and a mouse
	// event happens on its canvas. Only called if the widget registered for mouse
	// events.
	Mouse(m *mouse.Button) error

	// Options returns registration options for the widget.
	// This is how the widget indicates to the infrastructure whether it is
	// interested in keyboard or mouse shortcuts, what is its minimum canvas
	// size, etc.
	Options() *Options
}
