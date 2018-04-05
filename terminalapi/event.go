package terminalapi

import (
	"errors"
	"fmt"
	"image"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
)

// event.go defines events that can be received through the terminal API.

// Event represents an input event.
type Event interface {
	isEvent()
}

// Keyboard is the event used when a key is pressed.
// Implements terminalapi.Event.
type Keyboard struct {
	// Key is the pressed key.
	Key keyboard.Key
}

func (*Keyboard) isEvent() {}

// Resize is the event used when the terminal was resized.
// Implements terminalapi.Event.
type Resize struct {
	// Size is the new size of the terminal.
	Size image.Point
}

func (*Resize) isEvent() {}

// Mouse is the event used when the mouse is moved or a mouse button is
// pressed.
// Implements terminalapi.Event.
type Mouse struct {
	// Position of the mouse on the terminal.
	Position image.Point
	// Button identifies the pressed button if any.
	Button mouse.Button
}

func (*Mouse) isEvent() {}

// Error is an event indicating an error while processing input.
type Error string

// NewError returns a new Error event.
func NewError(e string) *Error {
	err := Error(e)
	return &err
}

// NewErrorf returns a new Error event, arguments are similar to fmt.Sprintf.
func NewErrorf(format string, args ...interface{}) *Error {
	err := Error(fmt.Sprintf(format, args...))
	return &err
}

func (*Error) isEvent() {}

// Error returns the error that occurred.
func (e *Error) Error() error {
	if e == nil || *e == "" {
		return nil
	}
	return errors.New(string(*e))
}

// String implements fmt.Stringer.
func (e Error) String() string {
	return string(e)
}