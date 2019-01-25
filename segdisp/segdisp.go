/*
Package segdisp simulates a 16-segment display drawn on a braille canvas.

Given a braille canvas, determines the placement and size of the individual
segments and exposes API that can turn individual segments on and off.

The following outlines segments in the display and their names.

       A1      A2
     ------- -------
    | \     |     / |
    |  \    |    /  |
  F |   H   J   K   | B
    |    \  |  /    |
    |     \ | /     |
     -G1---- ----G2-
    |     / | \     |
    |    /  |  \    |
  E |   N   M   L   | C
    |  /    |    \  |
    | /     |     \ |
     ------- -------  o
       D1      D2     DP
*/
package segdisp

import (
	"errors"
	"fmt"

	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/cell"
)

// Segment represents a single segment in the display.
type Segment int

// String implements fmt.Stringer()
func (s Segment) String() string {
	if n, ok := segmentNames[s]; ok {
		return n
	}
	return "SegmentUnknown"
}

// segmentNames maps Segment values to human readable names.
var segmentNames = map[Segment]string{
	A1: "A1",
	A2: "A2",
	B:  "B",
	C:  "C",
	D1: "D1",
	D2: "D2",
	E:  "E",
	F:  "F",
	G1: "G1",
	G2: "G2",
	H:  "H",
	J:  "J",
	K:  "K",
	L:  "L",
	M:  "M",
	N:  "M",
	DP: "DP",
}

const (
	segmentUnknown Segment = iota

	A1
	A2
	B
	C
	D1
	D2
	E
	F
	G1
	G2
	H
	J
	K
	L
	M
	N
	DP

	segmentMax // Used for validation.
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*Display)
}

// option implements Option.
type option func(*Display)

// set implements Option.set.
func (o option) set(d *Display) {
	o(d)
}

// CellOpts sets the cell options on the cells that contain the segment display.
func CellOpts(cOpts ...cell.Option) Option {
	return option(func(d *Display) {
		d.cellOpts = cOpts
	})
}

// Display represents the segment display.
// This object is not thread-safe.
type Display struct {
	// segments maps segments to their current status.
	segments map[Segment]bool

	cellOpts []cell.Option
}

// New creates a new segment display.
// Initially all the segments are off.
func New(opts ...Option) *Display {
	d := &Display{
		segments: map[Segment]bool{},
	}

	for _, opt := range opts {
		opt.set(d)
	}
	return d
}

// Clear clears the entire display, turning all segments off.
func (d *Display) Clear(opts ...Option) {
	for _, opt := range opts {
		opt.set(d)
	}

	d.segments = map[Segment]bool{}
}

// SetSegment sets the specified segment on.
// This method is idempotent.
func (d *Display) SetSegment(s Segment) error {
	if s <= segmentUnknown || s >= segmentMax {
		return fmt.Errorf("unknown segment %v", s)
	}
	d.segments[s] = true
	return nil
}

// ClearSegment sets the specified segment off.
// This method is idempotent.
func (d *Display) ClearSegment(s Segment) error {
	if s <= segmentUnknown || s >= segmentMax {
		return fmt.Errorf("unknown segment %v", s)
	}
	d.segments[s] = false
	return nil
}

// ToggleSegment toggles the state of the specified segment, i.e it either sets
// or clears it depending on its current state.
func (d *Display) ToggleSegment(s Segment) error {
	if s <= segmentUnknown || s >= segmentMax {
		return fmt.Errorf("unknown segment %v", s)
	}
	if d.segments[s] {
		d.segments[s] = false
	} else {
		d.segments[s] = true
	}
	return nil
}

// Minimum valid size of braille canvas in order to draw the segment display.
const (
	// MinColPixels is the smallest valid amount of columns in pixels.
	MinColPixels = 4 * braille.ColMult
	// MinRowPixels is the smallest valid amount of rows in pixels.
	MinRowPixels = 3 * braille.RowMult
)

// Draw draws the current state of the segment display onto the canvas.
// The canvas must be at 4x3 cells, or an error will be returned.
// Any options provided to draw overwrite the values provided to New.
func (d *Display) Draw(bc *braille.Canvas, opts ...Option) error {
	if size := bc.Size(); size.X < MinColPixels || size.Y < MinRowPixels {
		return fmt.Errorf("the canvas size %v is too small for the segment display, need at least %d columns and %d rows in pixels", size, MinColPixels, MinRowPixels)
	}

	// Determine line width.
	// Determine gap width.
	// Determine length of short and long segment.
	return errors.New("unimplemented")
}
