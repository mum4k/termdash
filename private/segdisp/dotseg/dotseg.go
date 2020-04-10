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
Package dotseg simulates a segment display that can draw dots.

Given a canvas, determines the placement and size of the individual
segments and exposes API that can turn individual segments on and off or
display dot characters.

The following outlines segments in the display and their names.

     ---------------
    |               |
    |               |
    |               |
    |       o D1    |
    |               |
    |               |
    |               |
    |       o D2    |
    |               |
    |               |
    |       o D3    |
     ---------------
*/
package dotseg

import (
	"fmt"
	"strings"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/segdisp"
	"github.com/mum4k/termdash/private/segdisp/segment"
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
	D1: "D1",
	D2: "D2",
	D3: "D3",
}

const (
	segmentUnknown Segment = iota

	// D1 is a segment, see the diagram above.
	D1
	// D2 is a segment, see the diagram above.
	D2
	// D3 is a segment, see the diagram above.
	D3

	segmentMax // Used for validation.
)

// characterSegments maps characters that can be displayed on their segments.
var characterSegments = map[rune][]Segment{
	':': {D1, D2},
	'.': {D3},
}

// SupportedChars returns all characters this display supports.
func SupportedChars() string {
	var b strings.Builder
	for r := range characterSegments {
		b.WriteRune(r)
	}
	return b.String()
}

// AllSegments returns all segments in an undefined order.
func AllSegments() []Segment {
	var res []Segment
	for s := range segmentNames {
		res = append(res, s)
	}
	return res
}

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
		return fmt.Errorf("unknown segment %v(%d)", s, s)
	}
	d.segments[s] = true
	return nil
}

// ClearSegment sets the specified segment off.
// This method is idempotent.
func (d *Display) ClearSegment(s Segment) error {
	if s <= segmentUnknown || s >= segmentMax {
		return fmt.Errorf("unknown segment %v(%d)", s, s)
	}
	d.segments[s] = false
	return nil
}

// ToggleSegment toggles the state of the specified segment, i.e it either sets
// or clears it depending on its current state.
func (d *Display) ToggleSegment(s Segment) error {
	if s <= segmentUnknown || s >= segmentMax {
		return fmt.Errorf("unknown segment %v(%d)", s, s)
	}
	if d.segments[s] {
		d.segments[s] = false
	} else {
		d.segments[s] = true
	}
	return nil
}

// SetCharacter sets all the segments that are needed to display the provided
// character.
// The display only supports characters returned by SupportedsChars().
// Doesn't clear the display of segments set previously.
func (d *Display) SetCharacter(c rune) error {
	seg, ok := characterSegments[c]
	if !ok {
		return fmt.Errorf("display doesn't support character %q rune(%v)", c, c)
	}

	for _, s := range seg {
		if err := d.SetSegment(s); err != nil {
			return err
		}
	}
	return nil
}

// Draw draws the current state of the segment display onto the canvas.
// The canvas must be at least MinCols x MinRows cells, or an error will be
// returned.
// Any options provided to draw overwrite the values provided to New.
func (d *Display) Draw(cvs *canvas.Canvas, opts ...Option) error {
	for _, o := range opts {
		o.set(d)
	}

	bc, bcAr, err := segdisp.ToBraille(cvs)
	if err != nil {
		return err
	}

	attr := newAttributes(bcAr)
	for seg, isSet := range d.segments {
		if !isSet {
			continue
		}

		ar, err := attr.segArea(seg)
		if err != nil {
			return err
		}
		if err := segment.HV(bc, ar, segment.Vertical, segment.CellOpts(d.cellOpts...)); err != nil {
			return err
		}
	}
	return bc.CopyTo(cvs)
}
