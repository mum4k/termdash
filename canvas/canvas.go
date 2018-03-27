// Package canvas defines the canvas that the widgets draw on.
package canvas

import (
	"image"

	"github.com/mum4k/termdash/cell"
)

// Canvas is where a widget draws its output for display on the terminal.
type Canvas struct{}

// Size returns the size of the 2-D canvas given to the widget.
func (c *Canvas) Size() image.Point {
	return image.Point{0, 0}
}

// Clear clears all the content on the canvas.
func (c *Canvas) Clear() {}

// FlushDesired provides a hint to the infrastructure that the canvas was
// changed and should be flushed to the terminal.
func (c *Canvas) FlushDesired() {}

// SetCell sets the value of the specified cell on the canvas.
// Use the options to specify which attributes to modify, if an attribute
// option isn't specified, the attribute retains its previous value.
func (c *Canvas) SetCell(p image.Point, r rune, opts ...cell.Option) {}
