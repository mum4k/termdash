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
Package sixteen simulates a 16-segment display drawn on a canvas.

Given a canvas, determines the placement and size of the individual
segments and exposes API that can turn individual segments on and off or
display characters.

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
package sixteen

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"log"
	"math"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw/segdisp/segment"
	"github.com/mum4k/termdash/numbers"
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
	N:  "N",
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

// characterSegments maps characters that can be displayed on their segments.
var characterSegments = map[rune][]Segment{
	' ': nil,
	'w': {E, N, L, C},
	'W': {F, E, N, L, C, B},
}

// SupportsChars asserts whether the display supports all runes in the
// provided string.
// The display only supports a subset of ASCII runes.
// Returns any unsupported runes found in the string in an unspecified order.
func SupportsChars(s string) (bool, []rune) {
	unsupp := map[rune]bool{}
	for _, r := range s {
		if _, ok := characterSegments[r]; !ok {
			unsupp[r] = true
		}
	}

	var res []rune
	for r := range unsupp {
		res = append(res, r)
	}
	return len(res) == 0, res
}

// Sanitize returns a copy of the string, replacing all unsupported characters
// with a space character.
func Sanitize(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		if _, ok := characterSegments[r]; !ok {
			b.WriteRune(' ')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// AllSegments returns all 16 segments in an undefined order.
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

// ErrUnsupportedCharacter is returned when the provided character cannot be displayed.
var ErrUnsupportedCharacter = errors.New("unsupported character")

// Character sets all the segments that are needed to display the provided character.
// Returns ErrUnsupportedCharacter when the character cannot be displayed.
// Doesn't clear the display of segments set previously.
func (d *Display) SetCharacter(c rune) error {
	seg, ok := characterSegments[c]
	if !ok {
		return ErrUnsupportedCharacter
	}

	for _, s := range seg {
		if err := d.SetSegment(s); err != nil {
			return err
		}
	}
	return nil
}

// Minimum valid size of a cell canvas in order to draw the segment display.
const (
	// MinCols is the smallest valid amount of columns in a cell area.
	MinCols = 6
	// MinRowPixels is the smallest valid amount of rows in a cell area.
	MinRows = 5
)

// aspectRatio is the desired aspect ratio of a single segment display.
var aspectRatio = image.Point{3, 5}

// Draw draws the current state of the segment display onto the canvas.
// The canvas must be at least MinCols x MinRows cells, or an error will be
// returned.
// Any options provided to draw overwrite the values provided to New.
func (d *Display) Draw(cvs *canvas.Canvas, opts ...Option) error {
	for _, o := range opts {
		o.set(d)
	}

	ar, err := Required(cvs.Area())
	if err != nil {
		return err
	}

	bc, err := braille.New(ar)
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
	}

	bcAr := area.WithRatio(bc.Area(), aspectRatio)
	segW := segWidth(bcAr)
	if segW == 4 {
		segW = 5
	}

	// Gap between the edge and the first segment.
	//_, edgeGap := numbers.MinMaxInts([]int{segW/2 + 1 + segW/6, 1})
	// Gap between two horizontal or vertical segments.
	_, diaGap := numbers.MinMaxInts([]int{int(float64(segW) * 0.4), 1})

	segLeg := float64(segW) / math.Sqrt2
	twoSegHypo := 2*segLeg + float64(diaGap)
	twoSegLeg := twoSegHypo / math.Sqrt2
	segPeakDist := segLeg / math.Sqrt2
	edgeSegGap := twoSegLeg - segPeakDist
	diaLeg := (float64(diaGap) / math.Sqrt2)
	peakToPeak := diaLeg * 2
	if segW == 2 {
		peakToPeak = 2
	}

	// Lengths of the short and long segment.
	shortL := (bcAr.Dx()-int(numbers.Round(2*edgeSegGap+peakToPeak)))/2 - 1
	longL := (bcAr.Dy()-int(numbers.Round(2*edgeSegGap+peakToPeak)))/2 - 1

	//log.Printf("dx:%d segW:%d, edgeGap:%d, segGap:%d, shortL:%d, longL:%d, end:%d, mid:%d, midGap:%d segDist:%d", bcAr.Dx(), segW, edgeGap, segGap, shortL, longL, end, mid, midGap, segDist)

	eg := int(numbers.Round(edgeSegGap))
	ptp := int(numbers.Round(peakToPeak))

	a1Ar := image.Rect(eg, 0, eg+shortL, segW)
	a2Ar := image.Rect(a1Ar.Max.X+ptp, 0, a1Ar.Max.X+ptp+shortL, segW)
	fAr := image.Rect(0, eg, segW, eg+longL)

	midStart := a1Ar.Max.X + int(numbers.Round(diaLeg-segPeakDist))

	jAr := image.Rect(midStart, eg, midStart+segW, eg+longL)

	endStart := a2Ar.Max.X + int(numbers.Round(diaLeg-segPeakDist))
	bAr := image.Rect(endStart, eg, endStart+segW, eg+longL)

	cenStart := fAr.Max.Y + int(numbers.Round(diaLeg-segPeakDist))
	g1Ar := image.Rect(eg, cenStart, eg+shortL, cenStart+segW)
	g2Ar := image.Rect(g1Ar.Max.X+ptp, cenStart, g1Ar.Max.X+ptp+shortL, cenStart+segW)

	eAr := image.Rect(0, fAr.Max.Y+ptp, segW, fAr.Max.Y+ptp+longL)
	mAr := image.Rect(midStart, jAr.Max.Y+ptp, midStart+segW, jAr.Max.Y+ptp+longL)
	cAr := image.Rect(endStart, bAr.Max.Y+ptp, endStart+segW, bAr.Max.Y+ptp+longL)

	botStart := eAr.Max.Y + int(numbers.Round(diaLeg-segPeakDist))
	d1Ar := image.Rect(eg, botStart, eg+shortL, botStart+segW)
	d2Ar := image.Rect(d1Ar.Max.X+ptp, botStart, d1Ar.Max.X+ptp+shortL, botStart+segW)

	var sOpts []segment.Option
	if len(d.cellOpts) > 0 {
		sOpts = append(sOpts, segment.CellOpts(d.cellOpts...))
	}
	for _, segArg := range []struct {
		s    Segment
		st   segment.Type
		ar   image.Rectangle
		opts []segment.Option
	}{
		{A1, segment.Horizontal, a1Ar, nil},
		{A2, segment.Horizontal, a2Ar, nil},

		{F, segment.Vertical, fAr, nil},
		{J, segment.Vertical, jAr, []segment.Option{segment.SkipSlopesLTE(2)}},
		{B, segment.Vertical, bAr, []segment.Option{segment.ReverseSlopes()}},

		{G1, segment.Horizontal, g1Ar, []segment.Option{segment.SkipSlopesLTE(2)}},
		{G2, segment.Horizontal, g2Ar, []segment.Option{segment.SkipSlopesLTE(2)}},

		{E, segment.Vertical, eAr, nil},
		{M, segment.Vertical, mAr, []segment.Option{segment.SkipSlopesLTE(2)}},
		{C, segment.Vertical, cAr, []segment.Option{segment.ReverseSlopes()}},

		{D1, segment.Horizontal, d1Ar, []segment.Option{segment.ReverseSlopes()}},
		{D2, segment.Horizontal, d2Ar, []segment.Option{segment.ReverseSlopes()}},
	} {
		if !d.segments[segArg.s] {
			continue
		}
		log.Printf("segment.HV for %v, ar:%v", segArg.s, segArg.ar)
		sOpts := append(sOpts, segArg.opts...)
		if err := segment.HV(bc, segArg.ar, segArg.st, sOpts...); err != nil {
			return fmt.Errorf("failed to draw segment %v, segment.HV => %v", segArg.s, err)
		}
	}

	topLStartX := int(numbers.Round(float64(a1Ar.Min.X) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	topLStartY := int(numbers.Round(float64(fAr.Min.Y) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	topLEndX := int(numbers.Round(float64(g1Ar.Max.X) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	topLEndY := int(numbers.Round(float64(jAr.Max.Y) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	hAr := image.Rect(topLStartX, topLStartY, topLEndX, topLEndY)

	topRStartX := int(numbers.Round(float64(a2Ar.Max.X) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	topRStartY := int(numbers.Round(float64(bAr.Min.Y) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	topREndX := int(numbers.Round(float64(g2Ar.Min.X) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	topREndY := int(numbers.Round(float64(jAr.Max.Y) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	kAr := image.Rect(topRStartX, topRStartY, topREndX, topREndY)

	botLStartX := int(numbers.Round(float64(g1Ar.Max.X) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	botLStartY := int(numbers.Round(float64(mAr.Min.Y) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	botLEndX := int(numbers.Round(float64(d1Ar.Min.X) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	botLEndY := int(numbers.Round(float64(eAr.Max.Y) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	nAr := image.Rect(botLStartX, botLStartY, botLEndX, botLEndY)

	botRStartX := int(numbers.Round(float64(g2Ar.Min.X) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	botRStartY := int(numbers.Round(float64(mAr.Min.Y) + segPeakDist - diaLeg + float64(diaGap)*0.3))
	botREndX := int(numbers.Round(float64(d2Ar.Max.X) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	botREndY := int(numbers.Round(float64(cAr.Max.Y) - segPeakDist + diaLeg - float64(diaGap)*0.3))
	lAr := image.Rect(botRStartX, botRStartY, botREndX, botREndY)

	var dsOpts []segment.DiagonalOption
	if len(d.cellOpts) > 0 {
		dsOpts = append(dsOpts, segment.DiagonalCellOpts(d.cellOpts...))
	}
	for _, segArg := range []struct {
		s  Segment
		dt segment.DiagonalType
		ar image.Rectangle
	}{
		{H, segment.LeftToRight, hAr},
		{K, segment.RightToLeft, kAr},
		{N, segment.RightToLeft, nAr},
		{L, segment.LeftToRight, lAr},
	} {
		if !d.segments[segArg.s] {
			continue
		}
		log.Printf("segment.Diagonal for %v, ar:%v", segArg.s, segArg.ar)
		if err := segment.Diagonal(bc, segArg.ar, segW, segArg.dt, dsOpts...); err != nil {
			return fmt.Errorf("failed to draw segment %v, segment.Diagonal => %v", segArg.s, err)
		}
	}

	return bc.CopyTo(cvs)
}

// Required, when given an area of cells, returns either an area of the same
// size or a smaller area that is required to draw one display.
// Returns a smaller area when the provided area didn't have the required
// aspect ratio.
// Returns an error if the area is too small to draw a segment display, i.e.
// smaller than MinCols x MinRows.
func Required(cellArea image.Rectangle) (image.Rectangle, error) {
	if cols, rows := cellArea.Dx(), cellArea.Dy(); cols < MinCols || rows < MinRows {
		return image.ZR, fmt.Errorf("cell area %v is too small to draw the segment display, has %dx%d cells, need at least %dx%d cells",
			cellArea, cols, rows, MinCols, MinRows)
	}

	bcAr := image.Rect(cellArea.Min.X, cellArea.Min.Y, cellArea.Max.X*braille.ColMult, cellArea.Max.Y*braille.RowMult)
	bcArAdj := area.WithRatio(bcAr, aspectRatio)

	needCols := int(math.Ceil(float64(bcArAdj.Dx()) / braille.ColMult))
	needRows := int(math.Ceil(float64(bcArAdj.Dy()) / braille.RowMult))
	needAr := image.Rect(cellArea.Min.X, cellArea.Min.Y, cellArea.Min.X+needCols, cellArea.Min.Y+needRows)
	return needAr, nil
}

// segWidth given an area for the display determines the width of individual segments.
func segWidth(ar image.Rectangle) int {
	// widthPerc is the relative width of a segment to the width of the canvas.
	const widthPerc = 9
	return int(numbers.Round(float64(ar.Dx()) * widthPerc / 100))
}
