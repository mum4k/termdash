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

// Package axes calculates the required layout and draws the X and Y axes of a heat map.
package axes

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/private/runewidth"
)

// axisWidth is width of an axis.
const axisWidth = 1

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

// RequiredWidth calculates the minimum width required
// in order to draw the Y axis and its labels.
// The parameter ls is the longest string in yLabels.
func RequiredWidth(ls string) int {
	return runewidth.StringWidth(ls) + axisWidth
}

// NewYDetails retrieves details about the Y axis required
// to draw it on a canvas of the provided area.
func NewYDetails(labels []string) (*YDetails, error) {
	if len(labels) == 0 {
		return &YDetails{}, nil
	}

	longest := ""
	longestWidth := 0
	for _, label := range labels {
		if width := runewidth.StringWidth(label); width > longestWidth {
			longest = label
			longestWidth = width
		}
	}

	width := RequiredWidth(longest)
	yAxisX := width - axisWidth
	lbs, err := yLabels(len(labels), yAxisX, labels)
	if err != nil {
		return nil, err
	}

	return &YDetails{
		Width:  width,
		Start:  image.Point{X: yAxisX, Y: 0},
		End:    image.Point{X: yAxisX, Y: len(labels) - 1},
		Labels: lbs,
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
func NewXDetails(cvsAr image.Rectangle, yEnd image.Point, labels []string, cellWidth int) (*XDetails, error) {
	if yEnd.Y < 0 {
		return nil, fmt.Errorf("invalid yEnd %v, want non-negative Y coordinate", yEnd)
	}
	if cellWidth <= 0 {
		return nil, fmt.Errorf("invalid cellWidth %d, want cellWidth > 0", cellWidth)
	}
	if len(labels) == 0 {
		return &XDetails{
			Start: image.Point{X: yEnd.X + 1, Y: yEnd.Y + 1},
			End:   image.Point{X: yEnd.X + 1, Y: yEnd.Y + 1},
		}, nil
	}

	graphWidth := len(labels) * cellWidth
	lbs, err := xLabels(yEnd, graphWidth, labels, cellWidth)
	if err != nil {
		return nil, err
	}

	start := image.Point{X: yEnd.X + 1, Y: yEnd.Y + 1}
	if start.X < 0 {
		start.X = 0
	}
	end := image.Point{X: start.X + graphWidth - 1, Y: start.Y}
	if start.X > cvsAr.Max.X {
		start.X = cvsAr.Max.X
	}
	if end.X > cvsAr.Max.X-1 {
		end.X = cvsAr.Max.X - 1
	}

	return &XDetails{
		Start:  start,
		End:    end,
		Labels: lbs,
	}, nil
}
