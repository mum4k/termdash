// Package canvas defines the canvas that the widgets draw on.
package canvas

import (
	"errors"
	"fmt"
	"image"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminalapi"
)

// Canvas is where a widget draws its output for display on the terminal.
type Canvas struct {
	// area is the area the buffer was created for.
	area image.Rectangle

	// buffer is where the drawing happens.
	buffer cell.Buffer
}

// New returns a new Canvas with a buffer for the provided area.
func New(area image.Rectangle) (*Canvas, error) {
	if area.Min.X < 0 || area.Min.Y < 0 || area.Max.X < 0 || area.Max.Y < 0 {
		return nil, fmt.Errorf("area cannot start or end on the negative axis, got: %+v", area)
	}
	size := image.Point{
		area.Dx() + 1,
		area.Dy() + 1,
	}
	b, err := cell.NewBuffer(size)
	if err != nil {
		return nil, err
	}
	return &Canvas{
		area:   area,
		buffer: b,
	}, nil
}

// Size returns the size of the 2-D canvas.
func (c *Canvas) Size() image.Point {
	return c.buffer.Size()
}

// Clear clears all the content on the canvas.
func (c *Canvas) Clear() error {
	b, err := cell.NewBuffer(c.Size())
	if err != nil {
		return err
	}
	c.buffer = b
	return nil
}

// SetCell sets the value of the specified cell on the canvas.
// Use the options to specify which attributes to modify, if an attribute
// option isn't specified, the attribute retains its previous value.
func (c *Canvas) SetCell(p image.Point, r rune, opts ...cell.Option) error {
	if area := c.buffer.Area(); !p.In(area) {
		return fmt.Errorf("cell at point %+v falls out of the canvas area %+v", p, area)
	}

	cell := c.buffer[p.X][p.Y]
	cell.Rune = r
	cell.Apply(opts...)
	return nil
}

// CopyTo copies the content of the canvas onto the provided terminal.
// Guarantees to stay within limits of the area the canvas was created with.
func (c *Canvas) Apply(t terminalapi.Terminal) error {
	return errors.New("unimplemented")
}
