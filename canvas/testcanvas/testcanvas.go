// Package testcanvas provides helpers for tests that use the canvas package.
package testcanvas

import (
	"image"
	"log"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
)

// MustNew returns a new canvas or panics.
func MustNew(area image.Rectangle) *canvas.Canvas {
	cvs, err := canvas.New(area)
	if err != nil {
		log.Fatalf("canvas.New => unexpected error: %v", err)
	}
	return cvs
}

// MustApply applies the canvas on the terminal or panics.
func MustApply(c *canvas.Canvas, t *faketerm.Terminal) {
	if err := c.Apply(t); err != nil {
		log.Fatalf("canvas.Apply => unexpected error: %v", err)
	}
}

// MustSetCell sets the cell value or panics.
func MustSetCell(c *canvas.Canvas, p image.Point, r rune, opts ...cell.Option) {
	if err := c.SetCell(p, r, opts...); err != nil {
		log.Fatalf("canvas.SetCell => unexpected error: %v", err)
	}
}
