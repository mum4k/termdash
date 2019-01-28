// Package segment provides functions that draw a single segment.
package segment

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
)

// SegmentType identifies the type of the segment that is drawn.
type SegmentType int

// String implements fmt.Stringer()
func (st SegmentType) String() string {
	if n, ok := segmentTypeNames[st]; ok {
		return n
	}
	return "SegmentTypeUnknown"
}

// segmentTypeNames maps SegmentType values to human readable names.
var segmentTypeNames = map[SegmentType]string{
	SegmentTypeHorizontal: "SegmentTypeHorizontal",
	SegmentTypeVertical:   "SegmentTypeVertical",
	SegmentTypeDiagonal:   "SegmentTypeDiagonal",
}

const (
	segmentTypeUnknown SegmentType = iota

	SegmentTypeHorizontal
	SegmentTypeVertical
	SegmentTypeDiagonal

	segmentTypeMax // Used for validation.
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options stores the provided options.
type options struct {
	cellOpts []cell.Option
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// CellOpts sets options on the cells that contain the segment.
// Cell options on a braille canvas can only be set on the entire cell, not per
// pixel.
func CellOpts(cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.cellOpts = cOpts
	})
}

// Draw draws a segment filling the provided area.
func Draw(bc *braille.Canvas, ar image.Rectangle, st SegmentType, opts ...Option) error {
	if ar.Min.X < 0 || ar.Min.Y < 0 {
		return fmt.Errorf("the start coordinates cannot be negative, got: %v", ar)
	}
	if ar.Max.X < 0 || ar.Max.Y < 0 {
		return fmt.Errorf("the end coordinates cannot be negative, got: %v", ar)
	}
	if ar.Dx() < 1 || ar.Dy() < 1 {
		return fmt.Errorf("the area for the segment must be at least 1x1 pixels, got %vx%v in area:%v", ar.Dx(), ar.Dy(), ar)
	}

	opt := &options{}
	for _, o := range opts {
		o.set(opt)
	}

	var nextLine nextLineFn
	var width int
	switch st {
	case SegmentTypeHorizontal:
		width = ar.Dy()
		nextLine = nextHorizLine

	default:
		return fmt.Errorf("unsupported segment type %v(%d)", st, st)
	}

	for i := 0; i < width; i++ {
		start, end := nextLine(i, ar)
		if err := draw.BrailleLine(bc, start, end, draw.BrailleLineCellOpts(opt.cellOpts...)); err != nil {
			return err
		}

	}
	return nil
}

// nextLine is a function that determines the start and end point of line number num in the segment.
type nextLineFn func(num int, ar image.Rectangle) (image.Point, image.Point)

// nextHorizLine determines the start and end point of individual lines in horizontal segment.
func nextHorizLine(num int, ar image.Rectangle) (image.Point, image.Point) {
	// Start and end points of the full line without adjustments for slopes.
	start := image.Point{ar.Min.X, ar.Min.Y + num}
	end := image.Point{ar.Max.X - 1, ar.Min.Y + num}

	height := ar.Dy()
	width := ar.Dx()
	if height < 2 || width < 3 {
		// No slopes under these dimensions as we don't have the resolution.
		return start, end
	}

	// Don't adjust lines that fall exactly in the middle of the segment height.
	// E.g when height divides oddly, we want the middle line to take the full width:
	//     --
	//    ----
	//     --
	//
	// And when the height divides oddly, we want the two middle lines to take
	// the full width:
	//     --
	//    ----
	//    ----
	//     --
	// We only do this for segments that are at least three lines tall.
	// For smaller segments we still want this behavior:
	//     --
	//    ----
	halfHeight := height / 2
	if height > 2 {
		if num == halfHeight || (height%2 == 0 && num == halfHeight-1) {
			return start, end
		}
	}

	if num < halfHeight {
		adjust := halfHeight - num
		if height%2 == 0 && height > 2 {
			// On evenly divided height, we need one less adjustment on every
			// line above the half, since two lines are taking the full width
			// as shown above.
			adjust--
		}
		return adjustHoriz(start, end, width, adjust)
	}
	adjust := num - halfHeight
	return adjustHoriz(start, end, width, adjust)
}

// adjustHoriz given start and end points that identify a horizontal line,
// returns points that are adjusted by the specified amount.
// I.e. the start is moved to the right and the end is moved to the left.
// The points won't be allowed to cross each other.
// The segWidth is the full width of the segment we are drawing.
func adjustHoriz(start, end image.Point, segWidth int, adjust int) (image.Point, image.Point) {
	ns := start.Add(image.Point{adjust, 0})
	ne := end.Sub(image.Point{adjust, 0})

	if ns.X <= ne.X {
		return ns, ne
	}

	halfWidth := segWidth / 2
	if segWidth%2 == 0 {
		// The width of the segment divides evenly, place start and end next to each other.
		// E.g: 0 1 2  3  4 5
		//      - - ns ne - -
		ns = image.Point{halfWidth - 1, start.Y}
		ne = image.Point{halfWidth, end.Y}
	} else {
		// The width of the segment divides oddly, place both start and end on the mid point.
		// E.g: 0 1  2   3 4
		//      - - nsne - -
		ns = image.Point{halfWidth, start.Y}
		ne = image.Point{halfWidth, end.Y}
	}
	return ns, ne
}
