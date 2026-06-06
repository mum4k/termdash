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
	"image"

	"github.com/mum4k/termdash/private/runewidth"
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
	if graphHeight < 0 {
		return nil, fmt.Errorf("invalid graphHeight %d, want graphHeight >= 0", graphHeight)
	}
	if labelWidth < 0 {
		return nil, fmt.Errorf("invalid labelWidth %d, want labelWidth >= 0", labelWidth)
	}
	if len(labels) != graphHeight {
		return nil, fmt.Errorf("invalid label count %d, want %d", len(labels), graphHeight)
	}

	out := make([]*Label, 0, len(labels))
	for row, label := range labels {
		l, err := rowLabel(row, label, labelWidth)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// rowLabel returns one label for the specified row.
// The row is the Y coordinate of the row, Y coordinates grow down.
func rowLabel(row int, label string, labelWidth int) (*Label, error) {
	if row < 0 {
		return nil, fmt.Errorf("invalid row %d, want row >= 0", row)
	}
	if labelWidth < 0 {
		return nil, fmt.Errorf("invalid labelWidth %d, want labelWidth >= 0", labelWidth)
	}
	width := runewidth.StringWidth(label)
	if width > labelWidth {
		return nil, fmt.Errorf("label %q width %d exceeds label width %d", label, width, labelWidth)
	}
	return &Label{
		Text: label,
		Pos:  image.Point{X: labelWidth - width, Y: row},
	}, nil
}

// xLabels returns labels that should be placed under the cells.
// Labels are returned with X coordinates in ascending order.
// X coordinates grow right.
func xLabels(yEnd image.Point, graphWidth int, labels []string, cellWidth int) ([]*Label, error) {
	if yEnd.Y < 0 {
		return nil, fmt.Errorf("invalid yEnd %v, want non-negative Y coordinate", yEnd)
	}
	if graphWidth < 0 {
		return nil, fmt.Errorf("invalid graphWidth %d, want graphWidth >= 0", graphWidth)
	}
	if cellWidth <= 0 {
		return nil, fmt.Errorf("invalid cellWidth %d, want cellWidth > 0", cellWidth)
	}
	if len(labels) == 0 || graphWidth == 0 {
		return nil, nil
	}

	longest := 0
	for _, label := range labels {
		if width := runewidth.StringWidth(label); width > longest {
			longest = width
		}
	}

	padded, index := paddedLabelLength(graphWidth, longest, cellWidth)
	groupCells := 1
	if padded > 0 {
		groupCells = padded / cellWidth
	}
	if groupCells < 1 {
		groupCells = 1
	}

	out := []*Label{}
	for i, label := range labels {
		if i%groupCells != index {
			continue
		}
		width := runewidth.StringWidth(label)
		groupStart := i * cellWidth
		if groupCells > 1 {
			groupStart = (i - index) * cellWidth
			if groupStart < 0 {
				groupStart = 0
			}
		}
		posX := yEnd.X + 1 + groupStart + maxInt(0, (padded-width)/2)
		if posX+width > yEnd.X+1+graphWidth {
			posX = yEnd.X + 1 + graphWidth - width
		}
		if posX < yEnd.X+1 {
			posX = yEnd.X + 1
		}
		out = append(out, &Label{
			Text: label,
			Pos:  image.Point{X: posX, Y: yEnd.Y + 1},
		})
	}
	return out, nil
}

// paddedLabelLength calculates the length of the padded X label and
// the column index corresponding to the label.
// For example, the longest X label's length is 5, like '12:34', and the cell's width is 3.
// So in order to better display, every three columns of cells will display a X label,
// the X label belongs to the middle column of the three columns,
// and the padded length is 3*3 (cellWidth multiplies the number of columns), which is 9.
func paddedLabelLength(graphWidth, longest, cellWidth int) (l, index int) {
	if graphWidth <= 0 || longest <= 0 || cellWidth <= 0 {
		return 0, 0
	}
	groupCells := (longest + cellWidth - 1) / cellWidth
	if groupCells < 1 {
		groupCells = 1
	}
	padded := groupCells * cellWidth
	if padded > graphWidth {
		padded = graphWidth
		groupCells = maxInt(1, padded/cellWidth)
	}
	return padded, groupCells / 2
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
