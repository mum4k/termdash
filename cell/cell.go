/*
Package cell implements cell options and attributes.

A cell is the smallest point on the terminal.
*/
package cell

import (
	"fmt"
	"image"
)

// Option is used to provide options for cells on a 2-D terminal.
type Option interface {
	// set sets the provided option.
	set(*Options)
}

// Options stores the provided options.
type Options struct {
	FgColor Color
	BgColor Color
}

// NewOptions returns a new Options instance after applying the provided options.
func NewOptions(opts ...Option) *Options {
	o := &Options{}
	for _, opt := range opts {
		opt.set(o)
	}
	return o
}

// Cell represents a single cell on the terminal.
type Cell struct {
	// Rune is the rune stored in the cell.
	Rune rune

	// Opts are the cell options.
	Opts *Options
}

// New returns a new cell.
func New(r rune, opts ...Option) *Cell {
	return &Cell{
		Rune: r,
		Opts: NewOptions(opts...),
	}
}

// Apply applies the provided options to the cell.
func (c *Cell) Apply(opts ...Option) {
	for _, opt := range opts {
		opt.set(c.Opts)
	}
}

// Buffer is a 2-D buffer of cells.
// The axes increase right and down.
type Buffer [][]*Cell

// NewBuffer returns a new Buffer of the provided size.
func NewBuffer(size image.Point) (Buffer, error) {
	if size.X <= 0 {
		return nil, fmt.Errorf("invalid buffer width (size.X): %d, must be a positive number", size.X)
	}
	if size.Y <= 0 {
		return nil, fmt.Errorf("invalid buffer height (size.Y): %d, must be a positive number", size.Y)
	}

	b := make([][]*Cell, size.X)
	for col := range b {
		b[col] = make([]*Cell, size.Y)
		for row := range b[col] {
			b[col][row] = New(0)
		}
	}
	return b, nil
}

// Size returns the size of the buffer.
func (b Buffer) Size() image.Point {
	return image.Point{
		len(b),
		len(b[0]),
	}
}

// Area returns the area that is covered by this buffer.
func (b Buffer) Area() image.Rectangle {
	s := b.Size()
	return image.Rect(0, 0, s.X-1, s.Y-1)
}

// option implements Option.
type option func(*Options)

// set implements Option.set.
func (co option) set(opts *Options) {
	co(opts)
}

// FgColor sets the foreground color of the cell.
func FgColor(color Color) Option {
	return option(func(co *Options) {
		co.FgColor = color
	})
}

// BgColor sets the background color of the cell.
func BgColor(color Color) Option {
	return option(func(co *Options) {
		co.BgColor = color
	})
}
