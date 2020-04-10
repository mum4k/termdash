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

package sixteen

// attributes.go calculates attributes needed when determining placement of
// segments.

import (
	"fmt"
	"image"
	"math"

	"github.com/mum4k/termdash/private/numbers"
	"github.com/mum4k/termdash/private/segdisp"
	"github.com/mum4k/termdash/private/segdisp/segment"
)

// hvSegType maps horizontal and vertical segments to their type.
var hvSegType = map[Segment]segment.Type{
	A1: segment.Horizontal,
	A2: segment.Horizontal,
	B:  segment.Vertical,
	C:  segment.Vertical,
	D1: segment.Horizontal,
	D2: segment.Horizontal,
	E:  segment.Vertical,
	F:  segment.Vertical,
	G1: segment.Horizontal,
	G2: segment.Horizontal,
	J:  segment.Vertical,
	M:  segment.Vertical,
}

// diaSegType maps diagonal segments to their type.
var diaSegType = map[Segment]segment.DiagonalType{
	H: segment.LeftToRight,
	K: segment.RightToLeft,
	N: segment.RightToLeft,
	L: segment.LeftToRight,
}

// Attributes contains attributes needed to draw the segment display.
// Refer to doc/segment_placement.svg for a visual aid and explanation of the
// usage of the square roots.
type Attributes struct {
	// segSize is the width of a vertical or height of a horizontal segment.
	segSize int

	// diaGap is the shortest distance between slopes on two neighboring
	// perpendicular segments.
	diaGap float64

	// segPeakDist is the distance between the peak of the slope on a segment
	// and the point where the slope ends.
	segPeakDist float64

	// diaLeg is the leg of a square whose hypotenuse is the diaGap.
	diaLeg float64

	// peakToPeak is a horizontal or vertical distance between peaks of two
	// segments.
	peakToPeak int

	// shortLen is length of the shorter segment, e.g. D1.
	shortLen int

	// longLen is length of the longer segment, e.g. F.
	longLen int

	// horizLeftX is the X coordinate where the area of the segment horizontally
	// on the left starts, i.e. X coordinate of F and E.
	horizLeftX int
	// horizMidX is the X coordinate where the area of the segment horizontally in
	// the middle starts, i.e. X coordinate of J and M.
	horizMidX int
	// horizRightX is the X coordinate where the area of the segment horizontally
	// on the right starts, i.e. X coordinate of B and C.
	horizRightX int

	// vertCenY is the Y coordinate where the area of the segment vertically
	// in the center starts, i.e. Y coordinate of G1 and G2.
	vertCenY int
	// VertBotY is the Y coordinate where the area of the segment vertically
	// at the bottom starts, i.e. Y coordinate of D1 and D2.
	VertBotY int
}

// NewAttributes calculates attributes needed to place the segments for the
// provided pixel area.
func NewAttributes(bcAr image.Rectangle) *Attributes {
	segSize := segdisp.SegmentSize(bcAr)

	// diaPerc is the size of the diaGap in percentage of the segment's size.
	const diaPerc = 40
	// Ensure there is at least one pixel diagonally between segments so they
	// don't visually blend.
	_, dg := numbers.MinMaxInts([]int{
		int(float64(segSize) * diaPerc / 100),
		1,
	})
	diaGap := float64(dg)

	segLeg := float64(segSize) / math.Sqrt2
	segPeakDist := segLeg / math.Sqrt2

	diaLeg := diaGap / math.Sqrt2
	peakToPeak := diaLeg * 2
	if segSize == 2 {
		// Display that has segment size of two looks more balanced with peak
		// distance of two.
		peakToPeak = 2
	}
	if peakToPeak > 3 && int(peakToPeak)%2 == 0 {
		// Prefer odd distances to create centered look.
		peakToPeak++
	}

	twoSegHypo := 2*segLeg + diaGap
	twoSegLeg := twoSegHypo / math.Sqrt2
	edgeSegGap := twoSegLeg - segPeakDist

	spaces := int(math.Round(2*edgeSegGap + peakToPeak))
	shortLen := (bcAr.Dx()-spaces)/2 - 1
	longLen := (bcAr.Dy()-spaces)/2 - 1

	ptp := int(math.Round(peakToPeak))
	horizLeftX := int(math.Round(edgeSegGap))

	// Refer to doc/segment_placement.svg.
	// Diagram labeled "A mid point".
	offset := int(math.Round(diaLeg - segPeakDist))
	horizMidX := horizLeftX + shortLen + offset
	horizRightX := horizLeftX + shortLen + ptp + shortLen + offset

	vertCenY := horizLeftX + longLen + offset
	vertBotY := horizLeftX + longLen + ptp + longLen + offset

	return &Attributes{
		segSize:     segSize,
		diaGap:      diaGap,
		segPeakDist: segPeakDist,
		diaLeg:      diaLeg,
		peakToPeak:  ptp,
		shortLen:    shortLen,
		longLen:     longLen,

		horizLeftX:  horizLeftX,
		horizMidX:   horizMidX,
		horizRightX: horizRightX,
		vertCenY:    vertCenY,
		VertBotY:    vertBotY,
	}
}

// hvSegArea returns the area for the specified horizontal or vertical segment.
func (a *Attributes) hvSegArea(s Segment) image.Rectangle {
	var (
		start  image.Point
		length int
	)

	switch s {
	case A1:
		start = image.Point{a.horizLeftX, 0}
		length = a.shortLen

	case A2:
		a1 := a.hvSegArea(A1)
		start = image.Point{a1.Max.X + a.peakToPeak, 0}
		length = a.shortLen

	case F:
		start = image.Point{0, a.horizLeftX}
		length = a.longLen

	case J:
		start = image.Point{a.horizMidX, a.horizLeftX}
		length = a.longLen

	case B:
		start = image.Point{a.horizRightX, a.horizLeftX}
		length = a.longLen

	case G1:
		start = image.Point{a.horizLeftX, a.vertCenY}
		length = a.shortLen

	case G2:
		g1 := a.hvSegArea(G1)
		start = image.Point{g1.Max.X + a.peakToPeak, a.vertCenY}
		length = a.shortLen

	case E:
		f := a.hvSegArea(F)
		start = image.Point{0, f.Max.Y + a.peakToPeak}
		length = a.longLen

	case M:
		j := a.hvSegArea(J)
		start = image.Point{a.horizMidX, j.Max.Y + a.peakToPeak}
		length = a.longLen

	case C:
		b := a.hvSegArea(B)
		start = image.Point{a.horizRightX, b.Max.Y + a.peakToPeak}
		length = a.longLen

	case D1:
		start = image.Point{a.horizLeftX, a.VertBotY}
		length = a.shortLen

	case D2:
		d1 := a.hvSegArea(D1)
		start = image.Point{d1.Max.X + a.peakToPeak, a.VertBotY}
		length = a.shortLen

	default:
		panic(fmt.Sprintf("cannot determine area for unknown horizontal or vertical segment %v(%d)", s, s))
	}

	return a.hvArFromStart(start, s, length)
}

// hvArFromStart given start coordinates of a segment, its length and its type,
// determines its area.
func (a *Attributes) hvArFromStart(start image.Point, s Segment, length int) image.Rectangle {
	st := hvSegType[s]
	switch st {
	case segment.Horizontal:
		return image.Rect(start.X, start.Y, start.X+length, start.Y+a.segSize)
	case segment.Vertical:
		return image.Rect(start.X, start.Y, start.X+a.segSize, start.Y+length)
	default:
		panic(fmt.Sprintf("cannot create area for segment of unknown type %v(%d)", st, st))
	}
}

// diaSegArea returns the area for the specified diagonal segment.
func (a *Attributes) diaSegArea(s Segment) image.Rectangle {
	switch s {
	case H:
		return a.diaBetween(A1, F, J, G1)
	case K:
		return a.diaBetween(A2, B, J, G2)
	case N:
		return a.diaBetween(G1, M, E, D1)
	case L:
		return a.diaBetween(G2, M, C, D2)

	default:
		panic(fmt.Sprintf("cannot determine area for unknown diagonal segment %v(%d)", s, s))
	}
}

// diaBetween given four segments (two horizontal and two vertical) returns the
// area between them for a diagonal segment.
func (a *Attributes) diaBetween(top, left, right, bottom Segment) image.Rectangle {
	topAr := a.hvSegArea(top)
	leftAr := a.hvSegArea(left)
	rightAr := a.hvSegArea(right)
	bottomAr := a.hvSegArea(bottom)

	// hvToDiaGapPerc is the size of gap between horizontal or vertical segment
	// and the diagonal segment between them in percentage of the diaGap.
	const hvToDiaGapPerc = 30
	hvToDiaGap := a.diaGap * hvToDiaGapPerc / 100

	startX := int(math.Round(float64(topAr.Min.X) + a.segPeakDist - a.diaLeg + hvToDiaGap))
	startY := int(math.Round(float64(leftAr.Min.Y) + a.segPeakDist - a.diaLeg + hvToDiaGap))
	endX := int(math.Round(float64(bottomAr.Max.X) - a.segPeakDist + a.diaLeg - hvToDiaGap))
	endY := int(math.Round(float64(rightAr.Max.Y) - a.segPeakDist + a.diaLeg - hvToDiaGap))
	return image.Rect(startX, startY, endX, endY)
}
