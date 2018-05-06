package draw

// rectangle.go draws a rectangle.

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
)

// RectangleOption is used to provide options to the Rectangle function.
type RectangleOption interface {
	// set sets the provided option.
	set(*rectOptions)
}

// rectOptions stores the provided options.
type rectOptions struct {
	cellOpts []cell.Option
	char     rune
}

// rectOption implements RectangleOption.
type rectOption func(rOpts *rectOptions)

// set implements Option.set.
func (ro rectOption) set(rOpts *rectOptions) {
	ro(rOpts)
}

// RectCellOpts sets options on the cells that create the rectangle.
func RectCellOpts(opts ...cell.Option) RectangleOption {
	return rectOption(func(rOpts *rectOptions) {
		rOpts.cellOpts = append(rOpts.cellOpts, opts...)
	})
}

// RectangleOption { is the default for the RectChar option.
const DefaultRectChar = ' '

// RectChar sets the character used in each of the cells of the rectangle.
func RectChar(c rune) RectangleOption {
	return rectOption(func(rOpts *rectOptions) {
		rOpts.char = c
	})
}

// Rectangle draws a filled rectangle on the canvas.
func Rectangle(c *canvas.Canvas, r image.Rectangle, opts ...RectangleOption) error {
	opt := &rectOptions{}
	for _, o := range opts {
		o.set(opt)
	}

	if ar := c.Area(); !r.In(ar) {
		return fmt.Errorf("the requested rectangle %v doesn't fit the canvas area %v", r, ar)
	}

	if r.Dx() < 1 || r.Dy() < 1 {
		return fmt.Errorf("the rectangle must be at least 1x1 cell, got %v", r)
	}

	for col := r.Min.X; col < r.Max.X; col++ {
		for row := r.Min.Y; row < r.Max.Y; row++ {
			if err := c.SetCell(image.Point{col, row}, opt.char, opt.cellOpts...); err != nil {
				return err
			}
		}
	}
	return nil
}
