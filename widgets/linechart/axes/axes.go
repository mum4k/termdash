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

	// axisWidth is width of an axis.
	axisWidth = 1
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

// RequiredWidth calculates the minimum width required in order to draw the Y
// axis and its labels.
func (y *Y) RequiredWidth() int {
	// This is an estimation only, it is possible that more labels in the
	// middle will be generated and might be wider than this. Such cases are
	// handled on the call to Details when the size of canvas is known.
	return longestLabel([]*Label{
		{Value: y.min},
		{Value: y.max},
	}) + axisWidth
}

// Details retrieves details about the Y axis required to draw it on a canvas
// of the provided area.
// The argument reqXHeight is the height required for the X axis and its labels.
func (y *Y) Details(cvsAr image.Rectangle, reqXHeight int, mode YScaleMode) (*YDetails, error) {
	cvsWidth := cvsAr.Dx()
	cvsHeight := cvsAr.Dy()
	maxWidth := cvsWidth - 1 // Reserve one column for the line chart itself.
	if req := y.RequiredWidth(); maxWidth < req {
		return nil, fmt.Errorf("the available maxWidth %d is smaller than the reported required width %d", maxWidth, req)
	}

	graphHeight := cvsHeight - reqXHeight
	scale, err := NewYScale(y.min.Value, y.max.Value, graphHeight, nonZeroDecimals, mode)
	if err != nil {
		return nil, err
	}

	// See how the labels would look like on the entire maxWidth.
	maxLabelWidth := maxWidth - axisWidth
	labels, err := yLabels(scale, maxLabelWidth)
	if err != nil {
		return nil, err
	}

	var width int
	// Determine the largest label, which might be less than maxWidth.
	// Such case would allow us to save more space for the line chart itself.
	widest := longestLabel(labels)
	if widest < maxLabelWidth {
		// Save the space and recalculate the labels, since they need to be realigned.
		l, err := yLabels(scale, widest)
		if err != nil {
			return nil, err
		}
		labels = l
		width = widest + axisWidth // One for the axis itself.
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

// longestLabel returns the width of the widest label.
func longestLabel(labels []*Label) int {
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
func NewXDetails(numPoints int, yStart image.Point, cvsAr image.Rectangle, customLabels map[int]string, lo LabelOrientation) (*XDetails, error) {
	cvsHeight := cvsAr.Dy()
	maxHeight := cvsHeight - 1 // Reserve one row for the line chart itself.
	reqHeight := RequiredHeight(numPoints, customLabels, lo)
	if maxHeight < reqHeight {
		return nil, fmt.Errorf("the available maxHeight %d is smaller than the reported required height %d", maxHeight, reqHeight)
	}

	// The space between the start of the axis and the end of the canvas.
	graphWidth := cvsAr.Dx() - yStart.X - 1
	scale, err := NewXScale(numPoints, graphWidth, nonZeroDecimals)
	if err != nil {
		return nil, err
	}

	// See how the labels would look like on the entire reqHeight.
	graphZero := image.Point{
		// Reserve one point horizontally for the Y axis.
		yStart.X + 1,
		cvsAr.Dy() - reqHeight - 1,
	}
	labels, err := xLabels(scale, graphZero, customLabels, lo)
	if err != nil {
		return nil, err
	}

	return &XDetails{
		Start:  image.Point{yStart.X, cvsAr.Dy() - reqHeight}, // Space for the labels.
		End:    image.Point{yStart.X + graphWidth, cvsAr.Dy() - reqHeight},
		Scale:  scale,
		Labels: labels,
	}, nil
}

// RequiredHeight calculates the minimum height required in order to draw the X
// axis and its labels.
func RequiredHeight(numPoints int, customLabels map[int]string, lo LabelOrientation) int {
	if lo == LabelOrientationHorizontal {
		// One row for the X axis and one row for its labels flowing
		// horizontally.
		return axisWidth + 1
	}

	labels := []*Label{
		{Value: NewValue(float64(numPoints), nonZeroDecimals)},
	}
	for _, cl := range customLabels {
		labels = append(labels, &Label{
			Value: NewTextValue(cl),
		})
	}
	return longestLabel(labels) + axisWidth
}
