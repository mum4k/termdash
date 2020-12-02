// Copyright 2020 Google Inc.
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
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/private/alignfor"
	"image"
)

// Label is one text label on an axis.
type Label struct {
	// Label content.
	Text string

	// Position of the label within the canvas.
	Pos image.Point
}

// yLabels returns labels that should be placed next to the cells.
// The labelWidth is the width of the area from the left-most side of the
// canvas until the Y axis (not including the Y axis). This is the area where
// the labels will be placed and aligned.
// Labels are returned with Y coordinates in ascending order.
// Y coordinates grow down.
func yLabels(graphHeight, labelWidth int, labels []string) ([]*Label, error) {
	if min := 2; graphHeight < min {
		return nil, fmt.Errorf("cannot place labels on a canvas with height %d, minimum is %d", graphHeight, min)
	}
	if min := 0; labelWidth < min {
		return nil, fmt.Errorf("cannot place labels in label area width %d, minimum is %d", labelWidth, min)
	}

	var ret []*Label
	for row, l := range labels {
		label, err := rowLabel(row, l, labelWidth)
		if err != nil {
			return nil, err
		}

		ret = append(ret, label)
	}

	return ret, nil
}

// rowLabel returns one label for the specified row.
// The row is the Y coordinate of the row, Y coordinates grow down.
func rowLabel(row int, label string, labelWidth int) (*Label, error) {
	// The area available for the label
	ar := image.Rect(0, row, labelWidth, row+1)

	pos, err := alignfor.Text(ar, label, align.HorizontalRight, align.VerticalMiddle)
	if err != nil {
		return nil, fmt.Errorf("unable to align the label value: %v", err)
	}

	return &Label{
		Text: label,
		Pos:  pos,
	}, nil
}

// xLabels returns labels that should be placed under the cells.
// Labels are returned with X coordinates in ascending order.
// X coordinates grow right.
func xLabels(yEnd image.Point, graphWidth int, labels []string, cellWidth int) ([]*Label, error) {
	var ret []*Label
	ls := LongestString(labels)

	if ls > cellWidth {
		length, index := paddedLabelLength(graphWidth, ls, cellWidth)
		for x := yEnd.X + 1; x <= graphWidth && index < len(labels); x += length {
			ar := image.Rect(x, yEnd.Y, x+length, yEnd.Y+1)
			pos, err := alignfor.Text(ar, labels[index], align.HorizontalCenter, align.VerticalMiddle)
			if err != nil {
				return nil, fmt.Errorf("unable to align the label value: %v", err)
			}

			l := &Label{
				Text: labels[index],
				Pos:  pos,
			}
			index += length / cellWidth
			ret = append(ret, l)
		}
	} else {
		for x := 0; x < len(labels); x++ {
			startX := yEnd.X + 1 + x*cellWidth
			ar := image.Rect(startX, yEnd.Y, startX+cellWidth, yEnd.Y+1)
			pos, err := alignfor.Text(ar, labels[x], align.HorizontalCenter, align.VerticalMiddle)
			if err != nil {
				return nil, fmt.Errorf("unable to align the label value: %v", err)
			}

			l := &Label{
				Text: labels[x],
				Pos:  pos,
			}
			ret = append(ret, l)
		}
	}

	return ret, nil
}

// paddedLabelLength calculates the length of the padded X label and
// the column index corresponding to the label.
// For example, the longest X label's length is 5, like '12:34', and the cell's width is 3.
// So in order to better display, every three columns of cells will display a X label,
// the X label belongs to the middle column of the three columns,
// and the padded length is 3*3 (cellWidth multiplies the number of columns), which is 9.
func paddedLabelLength(graphWidth, longest, cellWidth int) (l, index int) {
	l, index = 0, 0
	for i := longest/cellWidth + 1; i < graphWidth/cellWidth; i++ {
		if (i*cellWidth-longest)%2 == 0 {
			l = i * cellWidth
			index = i / 2
			break
		}
	}
	return
}
