// Package testdraw provides helpers for tests that use the draw package.
package testdraw

import (
	"image"
	"log"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
)

// MustBox draws box on the canvas or panics.
func MustBox(c *canvas.Canvas, box image.Rectangle, ls draw.LineStyle, opts ...cell.Option) {
	if err := draw.Box(c, box, ls, opts...); err != nil {
		log.Fatalf("draw.Box => unexpected error: %v", err)
	}
}

// MustText draws the text on the canvas or panics.
func MustText(c *canvas.Canvas, text string, tb draw.TextBounds, opts ...cell.Option) {
	if err := draw.Text(c, text, tb, opts...); err != nil {
		log.Fatalf("draw.Text => unexpected error: %v", err)
	}
}
