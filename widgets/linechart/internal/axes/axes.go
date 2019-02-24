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
	// The Y coordinate of Start is less than the Y coordinate of End.
	Start image.Point
	// End is the point where the Y axis ends.
	End image.Point

	// Scale is the scale of the Y axis.
	Scale *YScale

	// Labels are the labels for values on the Y axis in an increasing order.
	Labels []*Label
}

// RequiredWidth calculates the minimum width required in order to draw the Y
// axis and its labels when displaying values that have this minimum and
// maximum among all the series.
func RequiredWidth(minVal, maxVal float64) int {
	// This is an estimation only, it is possible that more labels in the
	// middle will be generated and might be wider than this. Such cases are
	// handled on the call to Details when the size of canvas is known.
	return longestLabel([]*Label{
		{Value: NewValue(minVal, nonZeroDecimals)},
		{Value: NewValue(maxVal, nonZeroDecimals)},
	}) + axisWidth
}

// YProperties are the properties of the Y axis.
type YProperties struct {
	// Min is the minimum value on the axis.
	Min float64
	// Max is the maximum value on the axis.
	Max float64
	// ReqXHeight is the height required for the X axis and its labels.
	ReqXHeight int
	// ScaleMode determines how the Y axis scales.
	ScaleMode YScaleMode
}

// NewYDetails retrieves details about the Y axis required to draw it on a
// canvas of the provided area.
func NewYDetails(cvsAr image.Rectangle, yp *YProperties) (*YDetails, error) {
	cvsWidth := cvsAr.Dx()
	cvsHeight := cvsAr.Dy()
	maxWidth := cvsWidth - 1 // Reserve one column for the line chart itself.
	if req := RequiredWidth(yp.Min, yp.Max); maxWidth < req {
		return nil, fmt.Errorf("the available maxWidth %d is smaller than the reported required width %d", maxWidth, req)
	}

	graphHeight := cvsHeight - yp.ReqXHeight
	scale, err := NewYScale(yp.Min, yp.Max, graphHeight, nonZeroDecimals, yp.ScaleMode)
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

	// Properties are the properties that were used on the call to NewXDetails.
	Properties *XProperties
}

// String implements fmt.Stringer.
func (xd *XDetails) String() string {
	return fmt.Sprintf("XDetails{Scale:%v}", xd.Scale)
}

// XProperties are the properties of the X axis.
type XProperties struct {
	// Min is the minimum value on the axis, i.e. the position of the first
	// displayed value from the series.
	Min int
	// Max is the maximum value on the axis, i.e. the position of the last
	// displayed value from the series.
	Max int
	// ReqYWidth is the width required for the Y axis and its labels.
	ReqYWidth int
	// CustomLabels are the desired labels for the X axis, these are preferred
	// if provided.
	CustomLabels map[int]string
	// LO is the desired orientation of labels under the X axis.
	LO LabelOrientation
}

// NewXDetails retrieves details about the X axis required to draw it on a canvas
// of the provided area. The yStart is the point where the Y axis starts.
// The numPoints is the number of points in the largest series that will be
// plotted.
// customLabels are the desired labels for the X axis, these are preferred if
// provided.
func NewXDetails(cvsAr image.Rectangle, xp *XProperties) (*XDetails, error) {
	cvsHeight := cvsAr.Dy()
	maxHeight := cvsHeight - 1 // Reserve one row for the line chart itself.
	reqHeight := RequiredHeight(xp.Max, xp.CustomLabels, xp.LO)
	if maxHeight < reqHeight {
		return nil, fmt.Errorf("the available maxHeight %d is smaller than the reported required height %d", maxHeight, reqHeight)
	}

	// The space between the start of the axis and the end of the canvas.
	graphWidth := cvsAr.Dx() - xp.ReqYWidth - 1
	scale, err := NewXScale(xp.Min, xp.Max, graphWidth, nonZeroDecimals)
	if err != nil {
		return nil, err
	}

	// See how the labels would look like on the entire reqHeight.
	graphZero := image.Point{
		// Reserve one point horizontally for the Y axis.
		xp.ReqYWidth + 1,
		cvsAr.Dy() - reqHeight - 1,
	}
	labels, err := xLabels(scale, graphZero, xp.CustomLabels, xp.LO)
	if err != nil {
		return nil, err
	}

	return &XDetails{
		Start:      image.Point{xp.ReqYWidth, cvsAr.Dy() - reqHeight}, // Space for the labels.
		End:        image.Point{xp.ReqYWidth + graphWidth, cvsAr.Dy() - reqHeight},
		Scale:      scale,
		Labels:     labels,
		Properties: xp,
	}, nil
}

// RequiredHeight calculates the minimum height required in order to draw the X
// axis and its labels.
func RequiredHeight(max int, customLabels map[int]string, lo LabelOrientation) int {
	if lo == LabelOrientationHorizontal {
		// One row for the X axis and one row for its labels flowing
		// horizontally.
		return axisWidth + 1
	}

	labels := []*Label{
		{Value: NewValue(float64(max), nonZeroDecimals)},
	}
	for _, cl := range customLabels {
		labels = append(labels, &Label{
			Value: NewTextValue(cl),
		})
	}
	return longestLabel(labels) + axisWidth
}
