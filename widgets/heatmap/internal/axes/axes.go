// Copyright 2021 Google Inc.
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

// Package axes calculates the required layout and draws the X and Y axes of a heat map.
package axes

import (
	"image"

	"github.com/mum4k/termdash/private/runewidth"
)

// AxisWidth is width of an axis.
const AxisWidth = 1

// YDetails contain information about the Y axis
// that will NOT be drawn onto the canvas, but will take up space.
type YDetails struct {
	// Width in character cells of the Y axis and its character labels.
	Width int

	// Start is the point where the Y axis starts.
	// The Y coordinate of Start is less than the Y coordinate of End.
	Start image.Point

	// End is the point where the Y axis ends.
	End image.Point

	// Labels are the labels for values on the Y axis in an increasing order.
	Labels []*Label
}

// NewYDetails retrieves details about the Y axis required
// to draw it on a canvas of the provided area.
func NewYDetails(labels []string) (*YDetails, error) {
	// See how the labels would look like on the entire maxWidth.
	maxLabelWidth := LongestString(labels)
	ls, err := yLabels(maxLabelWidth, labels)
	if err != nil {
		return nil, err
	}

	width := maxLabelWidth + AxisWidth
	graphHeight := len(labels)

	return &YDetails{
		Width:  width,
		Start:  image.Point{X: width - 1, Y: 0},
		End:    image.Point{X: width - 1, Y: graphHeight},
		Labels: ls,
	}, nil
}

// LongestString returns the length of the longest string in the string array.
func LongestString(strings []string) int {
	var widest int
	for _, s := range strings {
		if l := runewidth.StringWidth(s); l > widest {
			widest = l
		}
	}
	return widest
}

// XDetails contain information about the X axis
// that will NOT be drawn onto the canvas.
type XDetails struct {
	// Start is the point where the X axis starts.
	// Both coordinates of Start are less than End.
	Start image.Point
	// End is the point where the X axis ends.
	End image.Point

	// Labels are the labels for values on the X axis in an increasing order.
	Labels []*Label
}

// NewXDetails retrieves details about the X axis required to draw it on a canvas
// of the provided area.
// The yEnd is the point where the Y axis ends.
func NewXDetails(yEnd image.Point, labels []string, cellWidth int) (*XDetails, error) {
	// The space between the start of the axis and the end of the canvas.
	// graphWidth := cvsAr.Dx() - yEnd.X - 1
	graphWidth := len(labels) * cellWidth

	ls, err := xLabels(yEnd, graphWidth, labels, cellWidth)
	if err != nil {
		return nil, err
	}

	return &XDetails{
		Start:  image.Point{X: yEnd.X, Y: yEnd.Y - 1},
		End:    image.Point{X: yEnd.X + graphWidth, Y: yEnd.Y - 1},
		Labels: ls,
	}, nil
}
