// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package segment provides functions that draw a single segment.
package segment

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/draw"
)

// Type identifies the type of the segment that is drawn.
type Type int

// String implements fmt.Stringer()
func (st Type) String() string {
	if n, ok := segmentTypeNames[st]; ok {
		return n
	}
	return "TypeUnknown"
}

// segmentTypeNames maps Type values to human readable names.
var segmentTypeNames = map[Type]string{
	Horizontal: "Horizontal",
	Vertical:   "Vertical",
}

const (
	segmentTypeUnknown Type = iota

	// Horizontal is a horizontal segment.
	Horizontal
	// Vertical is a vertical segment.
	Vertical

	segmentTypeMax // Used for validation.
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options stores the provided options.
type options struct {
	cellOpts      []cell.Option
	skipSlopesLTE int
	reverseSlopes bool
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

// SkipSlopesLTE if provided instructs HV to not create slopes at the ends of a
// segment if the height of the horizontal or the width of the vertical segment
// is less or equal to the provided value.
func SkipSlopesLTE(v int) Option {
	return option(func(opts *options) {
		opts.skipSlopesLTE = v
	})
}

// ReverseSlopes if provided reverses the order in which slopes are drawn.
// This only has a visible effect when the horizontal segment has height of two
// or the vertical segment has width of two.
// Without this option segments with height / width of two look like this:
//    -   |
//   --- ||
//        |
//
// With this option:
//   ---  |
//    -   ||
//        |
func ReverseSlopes() Option {
	return option(func(opts *options) {
		opts.reverseSlopes = true
	})
}

// validArea validates the provided area.
func validArea(ar image.Rectangle) error {
	if ar.Min.X < 0 || ar.Min.Y < 0 {
		return fmt.Errorf("the start coordinates cannot be negative, got: %v", ar)
	}
	if ar.Max.X < 0 || ar.Max.Y < 0 {
		return fmt.Errorf("the end coordinates cannot be negative, got: %v", ar)
	}
	if ar.Dx() < 1 || ar.Dy() < 1 {
		return fmt.Errorf("the area for the segment must be at least 1x1 pixels, got %vx%v in area:%v", ar.Dx(), ar.Dy(), ar)
	}
	return nil
}

// HV draws a horizontal or a vertical display segment, filling the provided area.
// The segment will have slopes on both of its ends.
func HV(bc *braille.Canvas, ar image.Rectangle, st Type, opts ...Option) error {
	if err := validArea(ar); err != nil {
		return err
	}

	opt := &options{}
	for _, o := range opts {
		o.set(opt)
	}

	var nextLine nextHVLineFn
	var lines int
	switch st {
	case Horizontal:
		lines = ar.Dy()
		nextLine = nextHorizLine

	case Vertical:
		lines = ar.Dx()
		nextLine = nextVertLine

	default:
		return fmt.Errorf("unsupported segment type %v(%d)", st, st)
	}

	for i := 0; i < lines; i++ {
		start, end := nextLine(i, ar, opt)
		if err := draw.BrailleLine(bc, start, end, draw.BrailleLineCellOpts(opt.cellOpts...)); err != nil {
			return err
		}

	}
	return nil
}

// nextHVLineFn is a function that determines the start and end points of a line
// number num in a horizontal or a vertical segment.
type nextHVLineFn func(num int, ar image.Rectangle, opt *options) (image.Point, image.Point)

// nextHorizLine determines the start and end point of individual lines in a
// horizontal segment.
func nextHorizLine(num int, ar image.Rectangle, opt *options) (image.Point, image.Point) {
	// Start and end points of the full row without adjustments for slopes.
	start := image.Point{ar.Min.X, ar.Min.Y + num}
	end := image.Point{ar.Max.X - 1, ar.Min.Y + num}

	height := ar.Dy()
	width := ar.Dx()
	if height <= opt.skipSlopesLTE || height < 2 || width < 3 {
		// No slopes under these dimensions as we don't have the resolution.
		return start, end
	}

	// Don't adjust rows that fall exactly in the middle of the segment height.
	// E.g when height divides oddly, we want the middle row to take the full
	// width:
	//     --
	//    ----
	//     --
	//
	// And when the height divides oddly, we want the two middle rows to take
	// the full width:
	//     --
	//    ----
	//    ----
	//     --
	// We only do this for segments that are at least three rows tall.
	// For smaller segments we still want this behavior:
	//     --
	//    ----
	halfHeight := height / 2
	if height > 2 {
		if num == halfHeight || (height%2 == 0 && num == halfHeight-1) {
			return start, end
		}
	}
	if height == 2 && opt.reverseSlopes {
		return adjustHoriz(start, end, width, num)
	}

	if num < halfHeight {
		adjust := halfHeight - num
		if height%2 == 0 && height > 2 {
			// On evenly divided height, we need one less adjustment on every
			// row above the half, since two rows are taking the full width
			// as shown above.
			adjust--
		}
		return adjustHoriz(start, end, width, adjust)
	}
	adjust := num - halfHeight
	return adjustHoriz(start, end, width, adjust)
}

// nextVertLine determines the start and end point of individual lines in a
// vertical segment.
func nextVertLine(num int, ar image.Rectangle, opt *options) (image.Point, image.Point) {
	// Start and end points of the full column without adjustments for slopes.
	start := image.Point{ar.Min.X + num, ar.Min.Y}
	end := image.Point{ar.Min.X + num, ar.Max.Y - 1}

	height := ar.Dy()
	width := ar.Dx()
	if width <= opt.skipSlopesLTE || height < 3 || width < 2 {
		// No slopes under these dimensions as we don't have the resolution.
		return start, end
	}

	// Don't adjust lines that fall exactly in the middle of the segment height.
	// E.g when width divides oddly, we want the middle line to take the full
	// height:
	//    |
	//   |||
	//   |||
	//    |
	//
	// And when the width divides oddly, we want the two middle columns to take
	// the full height:
	//    ||
	//   ||||
	//   ||||
	//    ||
	//
	// We only do this for segments that are at least three columns wide.
	// For smaller segments we still want this behavior:
	//     |
	//    ||
	//    ||
	//     |
	halfWidth := width / 2
	if width > 2 {
		if num == halfWidth || (width%2 == 0 && num == halfWidth-1) {
			return start, end
		}
	}
	if width == 2 && opt.reverseSlopes {
		return adjustVert(start, end, width, num)
	}

	if num < halfWidth {
		adjust := halfWidth - num
		if width%2 == 0 && width > 2 {
			// On evenly divided width, we need one less adjustment on every
			// column above the half, since two lines are taking the full
			// height as shown above.
			adjust--
		}
		return adjustVert(start, end, height, adjust)
	}
	adjust := num - halfWidth
	return adjustVert(start, end, height, adjust)
}

// adjustHoriz given start and end points that identify a horizontal line,
// returns points that are adjusted towards each other on the line by the
// specified amount.
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
		ne = ns
	}
	return ns, ne
}

// adjustVert given start and end points that identify a vertical line,
// returns points that are adjusted towards each other on the line by the
// specified amount.
// I.e. the start is moved down and the end is moved up.
// The points won't be allowed to cross each other.
// The segHeight is the full height of the segment we are drawing.
func adjustVert(start, end image.Point, segHeight int, adjust int) (image.Point, image.Point) {
	adjStart, adjEnd := adjustHoriz(swapCoord(start), swapCoord(end), segHeight, adjust)
	return swapCoord(adjStart), swapCoord(adjEnd)
}

// swapCoord returns a point with its X and Y coordinates swapped.
func swapCoord(p image.Point) image.Point {
	return image.Point{p.Y, p.X}
}

// DiagonalType determines the type of diagonal segment.
type DiagonalType int

// String implements fmt.Stringer()
func (dt DiagonalType) String() string {
	if n, ok := diagonalTypeNames[dt]; ok {
		return n
	}
	return "DiagonalTypeUnknown"
}

// diagonalTypeNames maps DiagonalType values to human readable names.
var diagonalTypeNames = map[DiagonalType]string{
	LeftToRight: "LeftToRight",
	RightToLeft: "RightToLeft",
}

const (
	diagonalTypeUnknown DiagonalType = iota
	// LeftToRight is a diagonal segment from top left to bottom right.
	LeftToRight
	// RightToLeft is a diagonal segment from top right to bottom left.
	RightToLeft

	diagonalTypeMax // Used for validation.
)

// nextDiagLineFn is a function that determines the start and end points of a line
// number num in a diagonal segment.
// Points start and end define the first diagonal exactly in the middle.
// Points prevStart and prevEnd define line num-1.
type nextDiagLineFn func(num int, start, end, prevStart, prevEnd image.Point) (image.Point, image.Point)

// DiagonalOption is used to provide options.
type DiagonalOption interface {
	// set sets the provided option.
	set(*diagonalOptions)
}

// diagonalOptions stores the provided diagonal options.
type diagonalOptions struct {
	cellOpts []cell.Option
}

// diagonalOption implements DiagonalOption.
type diagonalOption func(*diagonalOptions)

// set implements DiagonalOption.set.
func (o diagonalOption) set(opts *diagonalOptions) {
	o(opts)
}

// DiagonalCellOpts sets options on the cells that contain the diagonal
// segment.
// Cell options on a braille canvas can only be set on the entire cell, not per
// pixel.
func DiagonalCellOpts(cOpts ...cell.Option) DiagonalOption {
	return diagonalOption(func(opts *diagonalOptions) {
		opts.cellOpts = cOpts
	})
}

// Diagonal draws a diagonal segment of the specified width filling the area.
func Diagonal(bc *braille.Canvas, ar image.Rectangle, width int, dt DiagonalType, opts ...DiagonalOption) error {
	if err := validArea(ar); err != nil {
		return err
	}
	if min := 1; width < min {
		return fmt.Errorf("invalid width %d, must be width >= %d", width, min)
	}
	opt := &diagonalOptions{}
	for _, o := range opts {
		o.set(opt)
	}

	var start, end image.Point
	var nextFn nextDiagLineFn
	switch dt {
	case LeftToRight:
		start = ar.Min
		end = image.Point{ar.Max.X - 1, ar.Max.Y - 1}
		nextFn = nextLRLine

	case RightToLeft:
		start = image.Point{ar.Max.X - 1, ar.Min.Y}
		end = image.Point{ar.Min.X, ar.Max.Y - 1}
		nextFn = nextRLLine

	default:
		return fmt.Errorf("unsupported diagonal type %v(%d)", dt, dt)
	}

	if err := draw.BrailleLine(bc, start, end, draw.BrailleLineCellOpts(opt.cellOpts...)); err != nil {
		return err
	}

	ns := start
	ne := end
	for i := 1; i < width; i++ {
		ns, ne = nextFn(i, start, end, ns, ne)

		if !ns.In(ar) || !ne.In(ar) {
			return fmt.Errorf("cannot draw diagonal segment of width %d in area %v, the area isn't large enough for line %v-%v", width, ar, ns, ne)
		}

		if err := draw.BrailleLine(bc, ns, ne, draw.BrailleLineCellOpts(opt.cellOpts...)); err != nil {
			return err
		}
	}
	return nil
}

// nextLRLine is a function that determines the start and end points of the
// next line of a left-to-right diagonal segment.
func nextLRLine(num int, start, end, prevStart, prevEnd image.Point) (image.Point, image.Point) {
	dist := num / 2
	if num%2 != 0 {
		// Every odd line is placed above the mid diagonal.
		ns := image.Point{start.X + dist + 1, start.Y}
		ne := image.Point{end.X, end.Y - dist - 1}
		return ns, ne
	}

	// Every even line is placed under the mid diagonal.
	ns := image.Point{start.X, start.Y + dist}
	ne := image.Point{end.X - dist, end.Y}
	return ns, ne
}

// nextRLLine is a function that determines the start and end points of the
// next line of a right-to-left diagonal segment.
func nextRLLine(num int, start, end, prevStart, prevEnd image.Point) (image.Point, image.Point) {
	dist := num / 2
	if num%2 != 0 {
		// Every odd line is placed above the mid diagonal.
		ns := image.Point{start.X - dist - 1, start.Y}
		ne := image.Point{end.X, end.Y - dist - 1}
		return ns, ne
	}

	// Every even line is placed under the mid diagonal.
	ns := image.Point{start.X, start.Y + dist}
	ne := image.Point{end.X + dist, end.Y}
	return ns, ne
}
