// Package testsegment provides helpers for tests that use the segment package.
package testsegment

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/draw/segdisp/segment"
)

// MustHV draws the segment or panics.
func MustHV(bc *braille.Canvas, ar image.Rectangle, st segment.Type, opts ...segment.Option) {
	if err := segment.HV(bc, ar, st, opts...); err != nil {
		panic(fmt.Sprintf("segment.HV => unexpected error: %v", err))
	}
}

// MustDiagonal draws the segment or panics.
func MustDiagonal(bc *braille.Canvas, ar image.Rectangle, width int, dt segment.DiagonalType, opts ...segment.DiagonalOption) {
	if err := segment.Diagonal(bc, ar, width, dt, opts...); err != nil {
		panic(fmt.Sprintf("segment.Diagonal => unexpected error: %v", err))
	}
}
