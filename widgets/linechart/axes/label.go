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
func yLabels(scale *YScale, labelWidth int) ([]*Label, error) {
	if min := 2; scale.CvsHeight < min {
		return nil, fmt.Errorf("cannot place labels on a canvas with height %d, minimum is %d", scale.CvsHeight, min)
	}
	if min := 1; labelWidth < min {
		return nil, fmt.Errorf("cannot place labels in label area width %d, minimum is %d", labelWidth, min)
	}

	var labels []*Label
	const labelSpacing = 4
	for y := scale.CvsHeight - 1; y >= 0; y -= labelSpacing {
		label, err := rowLabel(scale, y, labelWidth)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	// Always place at least two labels, first and last.
	if len(labels) < 2 {
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
	if err != nil {
		return nil, fmt.Errorf("unable to align label value for row %d: %v", y, err)
	}

	pos, err := align.Text(ar, v.Text(), align.HorizontalRight, align.VerticalMiddle)
	return &Label{
		Value: v,
		Pos:   pos,
	}, nil
}
