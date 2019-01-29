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

/*
Package segdisp simulates a 16-segment display drawn on a canvas.

Given a canvas, determines the placement and size of the individual
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
     ------- -------
       D1      D2
*/
package segdisp

import (
	"fmt"
	"image"
	"log"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/numbers"
	"github.com/mum4k/termdash/segdisp/segment"
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

	segmentMax // Used for validation.
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*Display)
}

// AllSegments returns all 16 segments in an undefined order.
func AllSegments() []Segment {
	var res []Segment
	for s := range segmentNames {
		res = append(res, s)
	}
	return res
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

// Minimum valid size of a cell canvas in order to draw the segment display.
const (
	// MinCols is the smallest valid amount of columns in a cell area.
	MinCols = 4
	// MinRowPixels is the smallest valid amount of rows in a cell area.
	MinRows = 3
)

// Draw draws the current state of the segment display onto the canvas.
// The canvas must be at least MinCols x MinRows cells, or an error will be
// returned.
// Any options provided to draw overwrite the values provided to New.
func (d *Display) Draw(cvs *canvas.Canvas, opts ...Option) error {
	ar, err := Required(cvs.Area())
	if err != nil {
		return err
	}

	bc, err := braille.New(ar)
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
	}

	bcAr := bc.Area()
	sw := segWidth(bcAr)
	half := bcAr.Dx() / 2
	log.Printf("bcAr:%v, sw:%d, half:%d", bcAr, sw, half)

	a1 := image.Rect(sw-1, 0, half-sw/2, sw)
	a2 := image.Rect(half+sw/2, 0, bcAr.Max.X-1, sw)
	log.Printf("a1:%v", a1)
	log.Printf("a2:%v", a2)
	for _, segAr := range []image.Rectangle{a1, a2} {
		if err := segment.HV(bc, segAr, segment.SegmentTypeHorizontal); err != nil {
			return fmt.Errorf("segment.HV => %v", err)
		}
	}

	// Determine gap width.
	// Determine length of short and long segment.
	return bc.CopyTo(cvs)
}

// Required, when given an area of cells, returns either an area of the same
// size or a smaller area that is required to draw one display.
// Returns a smaller area when the provided area didn;t have the required
// aspect ratio.
// Returns an error if the area is too small to draw a segment display.
func Required(cellArea image.Rectangle) (image.Rectangle, error) {
	ar := area.WithRatio(cellArea, image.Point{MinCols, MinRows})
	if ar.Empty() {
		return image.ZR, fmt.Errorf("cell area %v is to small to draw the segment display, need at least %d x %d cells", cellArea, MinCols, MinRows)
	}
	return ar, nil
}

// segWidth given an area for the display determines the width of individual segments.
func segWidth(ar image.Rectangle) int {
	// widthPerc is the relative width of a segment to the width of the canvas.
	const widthPerc = 10
	return int(numbers.Round(float64(ar.Dx()) * 10 / 100))
}
