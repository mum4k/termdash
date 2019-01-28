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

// Package axes calculates the required layout and draws the X and Y axes of a line chart.
package axes

import (
	"fmt"
	"image"
)

const (
	// nonZeroDecimals determines the overall precision of values displayed on the
	// graph, it indicates the number of non-zero decimal places the values will be
	// rounded up to.
	nonZeroDecimals = 2

	// yAxisWidth is width of the Y axis.
	yAxisWidth = 1
)

// YDetails contain information about the Y axis that will be drawn onto the
// canvas.
type YDetails struct {
	// Width in character cells of the Y axis and its character labels.
	Width int

	// Start is the point where the Y axis starts.
	// Both coordinates of Start are less than End.
	Start image.Point
	// End is the point where the Y axis ends.
	End image.Point

	// Scale is the scale of the Y axis.
	Scale *YScale

	// Labels are the labels for values on the Y axis in an increasing order.
	Labels []*Label
}

// Y tracks the state of the Y axis throughout the lifetime of a line chart.
// Implements lazy resize of the axis to decrease visual "jumping".
// This object is not thread-safe.
type Y struct {
	// min is the smallest value on the Y axis.
	min *Value
	// max is the largest value on the Y axis.
	max *Value
	// details about the Y axis as it will be drawn.
	details *YDetails
}

// NewY returns a new Y instance.
// The minVal and maxVal represent the minimum and maximum value that will be
// displayed on the line chart among all of the series.
func NewY(minVal, maxVal float64) *Y {
	y := &Y{}
	y.Update(minVal, maxVal)
	return y
}

// Update updates the stored minVal and maxVal.
func (y *Y) Update(minVal, maxVal float64) {
	y.min, y.max = NewValue(minVal, nonZeroDecimals), NewValue(maxVal, nonZeroDecimals)
}

// RequiredWidth calculates the minimum width required in order to draw the Y axis.
func (y *Y) RequiredWidth() int {
	// This is an estimation only, it is possible that more labels in the
	// middle will be generated and might be wider than this. Such cases are
	// handled on the call to Details when the size of canvas is known.
	return widestLabel([]*Label{
		{Value: y.min},
		{Value: y.max},
	}) + yAxisWidth
}

// Details retrieves details about the Y axis required to draw it on a canvas
// of the provided area.
func (y *Y) Details(cvsAr image.Rectangle, mode YScaleMode) (*YDetails, error) {
	cvsWidth := cvsAr.Dx()
	cvsHeight := cvsAr.Dy()
	maxWidth := cvsWidth - 1 // Reserve one row for the line chart itself.
	if req := y.RequiredWidth(); maxWidth < req {
		return nil, fmt.Errorf("the received maxWidth %d is smaller than the reported required width %d", maxWidth, req)
	}

	graphHeight := cvsHeight - 2 // One row for the X axis and one for its labels.
	scale, err := NewYScale(y.min.Value, y.max.Value, graphHeight, nonZeroDecimals, mode)
	if err != nil {
		return nil, err
	}

	// See how the labels would look like on the entire maxWidth.
	maxLabelWidth := maxWidth - yAxisWidth
	labels, err := yLabels(scale, maxLabelWidth)
	if err != nil {
		return nil, err
	}

	var width int
	// Determine the largest label, which might be less than maxWidth.
	// Such case would allow us to save more space for the line chart itself.
	widest := widestLabel(labels)
	if widest < maxLabelWidth {
		// Save the space and recalculate the labels, since they need to be realigned.
		l, err := yLabels(scale, widest)
		if err != nil {
			return nil, err
		}
		labels = l
		width = widest + yAxisWidth // One for the axis itself.
	} else {
		width = maxWidth
	}

	return &YDetails{
		Width:  width,
		Start:  image.Point{width - 1, 0},
		End:    image.Point{width - 1, graphHeight},
		Scale:  scale,
		Labels: labels,
	}, nil
}

// widestLabel returns the width of the widest label.
func widestLabel(labels []*Label) int {
	var widest int
	for _, label := range labels {
		if l := len(label.Value.Text()); l > widest {
			widest = l
		}
	}
	return widest
}

// XDetails contain information about the X axis that will be drawn onto the
// canvas.
type XDetails struct {
	// Start is the point where the X axis starts.
	// Both coordinates of Start are less than End.
	Start image.Point
	// End is the point where the X axis ends.
	End image.Point

	// Scale is the scale of the X axis.
	Scale *XScale

	// Labels are the labels for values on the X axis in an increasing order.
	Labels []*Label
}

// NewXDetails retrieves details about the X axis required to draw it on a canvas
// of the provided area. The yStart is the point where the Y axis starts.
// The numPoints is the number of points in the largest series that will be
// plotted.
// customLabels are the desired labels for the X axis, these are preferred if
// provided.
func NewXDetails(numPoints int, yStart image.Point, cvsAr image.Rectangle, customLabels map[int]string) (*XDetails, error) {
	if min := 3; cvsAr.Dy() < min {
		return nil, fmt.Errorf("the canvas isn't tall enough to accommodate the X axis, its labels and the line chart, got height %d, minimum is %d", cvsAr.Dy(), min)
	}

	// The space between the start of the axis and the end of the canvas.
	graphWidth := cvsAr.Dx() - yStart.X - 1
	scale, err := NewXScale(numPoints, graphWidth, nonZeroDecimals)
	if err != nil {
		return nil, err
	}

	// One point horizontally for the Y axis.
	// Two points vertically, one for the X axis and one for its labels.
	graphZero := image.Point{yStart.X + 1, cvsAr.Dy() - 3}
	labels, err := xLabels(scale, graphZero, customLabels)
	if err != nil {
		return nil, err
	}
	return &XDetails{
		Start:  image.Point{yStart.X, cvsAr.Dy() - 2}, // One row for the labels.
		End:    image.Point{yStart.X + graphWidth, cvsAr.Dy() - 2},
		Scale:  scale,
		Labels: labels,
	}, nil
}
