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

package axes

// label.go contains code that calculates the positions of labels on the axes.

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/internal/alignfor"
)

// LabelOrientation represents the orientation of text labels.
type LabelOrientation int

// String implements fmt.Stringer()
func (lo LabelOrientation) String() string {
	if n, ok := labelOrientationNames[lo]; ok {
		return n
	}
	return "LabelOrientationUnknown"
}

// labelOrientationNames maps LabelOrientation values to human readable names.
var labelOrientationNames = map[LabelOrientation]string{
	LabelOrientationHorizontal: "LabelOrientationHorizontal",
	LabelOrientationVertical:   "LabelOrientationVertical",
}

const (
	// LabelOrientationHorizontal is the default label orientation where text
	// flows horizontally.
	LabelOrientationHorizontal LabelOrientation = iota

	// LabelOrientationVertical is an orientation where text flows vertically.
	LabelOrientationVertical
)

// Label is one value label on an axis.
type Label struct {
	// Value if the value to be displayed.
	Value *Value

	// Position of the label within the canvas.
	Pos image.Point
}

// yLabels returns labels that should be placed next to the Y axis.
// The labelWidth is the width of the area from the left-most side of the
// canvas until the Y axis (not including the Y axis). This is the area where
// the labels will be placed and aligned.
// Labels are returned in an increasing value order.
// Label value is not trimmed to the provided labelWidth, the label width is
// only used to align the labels. Alignment is done with the assumption that
// longer labels will be trimmed.
func yLabels(scale *YScale, labelWidth int) ([]*Label, error) {
	if min := 2; scale.GraphHeight < min {
		return nil, fmt.Errorf("cannot place labels on a canvas with height %d, minimum is %d", scale.GraphHeight, min)
	}
	if min := 1; labelWidth < min {
		return nil, fmt.Errorf("cannot place labels in label area width %d, minimum is %d", labelWidth, min)
	}

	var labels []*Label
	const labelSpacing = 4
	seen := map[string]bool{}
	for y := scale.GraphHeight - 1; y >= 0; y -= labelSpacing {
		label, err := rowLabel(scale, y, labelWidth)
		if err != nil {
			return nil, err
		}
		if !seen[label.Value.Text()] {
			labels = append(labels, label)
			seen[label.Value.Text()] = true
		}
	}

	// If we have data, place at least two labels, first and last.
	haveData := scale.Min.Rounded != 0 || scale.Max.Rounded != 0
	if len(labels) < 2 && haveData {
		const maxRow = 0
		label, err := rowLabel(scale, maxRow, labelWidth)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	return labels, nil
}

// rowLabelArea determines the area available for labels on the specified row.
// The row is the Y coordinate of the row, Y coordinates grow down.
func rowLabelArea(row int, labelWidth int) image.Rectangle {
	return image.Rect(0, row, labelWidth, row+1)
}

// rowLabel returns label for the specified row.
func rowLabel(scale *YScale, y int, labelWidth int) (*Label, error) {
	v, err := scale.CellLabel(y)
	if err != nil {
		return nil, fmt.Errorf("unable to determine label value for row %d: %v", y, err)
	}

	ar := rowLabelArea(y, labelWidth)
	pos, err := alignfor.Text(ar, v.Text(), align.HorizontalRight, align.VerticalMiddle)
	if err != nil {
		return nil, fmt.Errorf("unable to align the label value: %v", err)
	}
	return &Label{
		Value: v,
		Pos:   pos,
	}, nil
}

// xSpace represents an available space among the X axis.
type xSpace struct {
	// min is the current relative coordinate.
	// These are zero based, i.e. not adjusted to axisStart.
	cur int
	// max is the maximum relative coordinate.
	// These are zero based, i.e. not adjusted to axisStart.
	// The xSpace instance contains points 0 <= x < max
	max int

	// graphZero is the (0, 0) point on the graph.
	graphZero image.Point
}

// newXSpace returns a new xSpace instance initialized for the provided width.
func newXSpace(graphZero image.Point, graphWidth int) *xSpace {
	return &xSpace{
		cur:       0,
		max:       graphWidth,
		graphZero: graphZero,
	}
}

// Implements fmt.Stringer.
func (xs *xSpace) String() string {
	return fmt.Sprintf("xSpace(size:%d)-cur:%v-max:%v", xs.Remaining(), image.Point{xs.cur, xs.graphZero.Y}, image.Point{xs.max, xs.graphZero.Y})
}

// Remaining returns the remaining size on the X axis.
func (xs *xSpace) Remaining() int {
	return xs.max - xs.cur
}

// Relative returns the relative coordinate within the space, these are zero
// based.
func (xs *xSpace) Relative() image.Point {
	return image.Point{xs.cur, xs.graphZero.Y + 1}
}

// LabelPos returns the absolute coordinate on the canvas where a label should
// be placed. The is the coordinate that represents the current relative
// coordinate of the space.
func (xs *xSpace) LabelPos() image.Point {
	return image.Point{xs.cur + xs.graphZero.X, xs.graphZero.Y + 2} // First down is the axis, second the label.
}

// Sub subtracts the specified size from the beginning of the available
// space.
func (xs *xSpace) Sub(size int) error {
	if xs.Remaining() < size {
		return fmt.Errorf("unable to subtract %d from the start, not enough size in %v", size, xs)
	}
	xs.cur += size
	return nil
}

// xLabels returns labels that should be placed under the X axis.
// The graphZero is the (0, 0) point of the graph area on the canvas.
// Labels are returned in an increasing value order.
// Returned labels shouldn't be trimmed, their count is adjusted so that they
// fit under the width of the axis.
// The customLabels map value positions in the series to the desired custom
// label. These are preferred if present.
func xLabels(scale *XScale, graphZero image.Point, customLabels map[int]string, lo LabelOrientation) ([]*Label, error) {
	space := newXSpace(graphZero, scale.GraphWidth)
	const minSpacing = 3
	var res []*Label

	next := int(scale.Min.Value)
	for haveLabels := 0; haveLabels <= int(scale.Max.Value); haveLabels = len(res) {
		label, err := colLabel(scale, space, customLabels, lo)
		if err != nil {
			return nil, err
		}
		if label == nil {
			break
		}
		res = append(res, label)

		next++
		if next > int(scale.Max.Value) {
			break
		}
		nextCell, err := scale.ValueToCell(next)
		if err != nil {
			return nil, err
		}

		skip := nextCell - space.Relative().X
		if skip < minSpacing {
			skip = minSpacing
		}

		if space.Remaining() <= skip {
			break
		}
		if err := space.Sub(skip); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// colLabel returns a label placed at the beginning of the space.
// The space is adjusted according to how much space was taken by the label.
// Returns nil, nil if the label doesn't fit in the space.
func colLabel(scale *XScale, space *xSpace, customLabels map[int]string, lo LabelOrientation) (*Label, error) {
	pos := space.Relative()
	label, err := scale.CellLabel(pos.X)
	if err != nil {
		return nil, fmt.Errorf("unable to determine label value for column %d: %v", pos.X, err)
	}

	if custom, ok := customLabels[int(label.Value)]; ok {
		label = NewTextValue(custom)
	}

	var labelLen int
	switch lo {
	case LabelOrientationHorizontal:
		labelLen = len(label.Text())
	case LabelOrientationVertical:
		labelLen = 1
	}
	if labelLen > space.Remaining() {
		return nil, nil
	}

	abs := space.LabelPos()
	if err := space.Sub(labelLen); err != nil {
		return nil, err
	}

	return &Label{
		Value: label,
		Pos:   abs,
	}, nil
}
