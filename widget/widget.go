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
// Implementations must be thread safe.
type Widget interface {
	// When the infrastructure calls Draw(), the widget must block on the call
	// until it finishes drawing onto the provided canvas. When given the
	// canvas, the widget must first determine its size by calling
	// Canvas.Size(), then limit all its drawing to this area.
	//
	// The widget must not assume that the size of the canvas or its content
	// remains the same between calls.
	Draw(canvas *canvas.Canvas) error

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
