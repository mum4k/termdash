// Package terminalapi defines the API of all terminal implementations.
package terminalapi

import (
	"context"
	"image"

	"github.com/mum4k/termdash/cell"
)

// Terminal abstracts an implementation of a 2-D terminal.
// A terminal consists of a number of cells.
type Terminal interface {
	// Size returns the terminal width and height in cells.
	Size() image.Point

	// Clear clears the content of the internal back buffer, resetting all cells
	// to their default content and attributes.
	Clear() error
	// Flush flushes the internal back buffer to the terminal.
	Flush() error

	// SetCursor sets the position of the cursor.
	SetCursor(p image.Point)
	// HideCursos hides the cursor.
	HideCursor()

	// SetCell sets the value of the specified cell to the provided rune.
	// Use the options to specify which attributes to modify, if an attribute
	// option isn't specified, the attribute retains its previous value.
	SetCell(p image.Point, r rune, opts ...cell.Option) error

	// Event waits for the next event and returns it.
	// This call blocks until the next event or cancellation of the context.
	Event(ctx context.Context) Event
}
