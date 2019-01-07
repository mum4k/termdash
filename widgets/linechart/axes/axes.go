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
	"image"
)

// nonZeroDecimals determines the overall precision of values displayed on the
// graph, it indicates the number of non-zero decimal places the values will be
// rounded up to.
const nonZeroDecimals = 2

// YDetails contain information about the Y axis that will be drawn onto the
// canvas.
type YDetails struct {
	// Width in character cells of the Y axis and its character labels.
	Width int

	// Scale is the scale of the Y axis.
	Scale *YScale

	// Labels are the labels for values on the Y axis in an increasing order.
	Labels []*Label
}

// Label is one value label on an axis.
type Label struct {
	// Text for the label.
	Text string

	// Position of the label within the canvas.
	Pos image.Point
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
func NewY(minVal, maxVal float64) (*Y, error) {
	y := &Y{}
	if err := y.Update(minVal, maxVal); err != nil {
		return nil, err
	}
	return y, nil
}

// RequiredWidth calculates the minimum width required in order to draw the Y axis.
func (y *Y) RequiredWidth() (int, error) {
	minT, maxT := y.min.Text(), y.max.Text()
	var widest int
	if minW, maxW := len(minT), len(maxT); minW > maxW {
		widest = minW
	} else {
		widest = maxW
	}
	const axisWidth = 1
	return widest + axisWidth, nil
}

// Update updates the stored minVal and maxVal.
func (y *Y) Update(minVal, maxVal float64) error {
	y.min, y.max = NewValue(minVal, nonZeroDecimals), NewValue(maxVal, nonZeroDecimals)
	return nil
}

// Details retrieves details about the Y axis required to draw it on the provided canvas.
// of the provided height.
func (y *Y) Details(cvsHeight int) (*YDetails, error) {
	w, err := y.RequiredWidth()
	if err != nil {
		return nil, err
	}

	const labelSpace = 2 // In cells.
	var labels []*Label

	//scale := y.max.Value / float64(cvsHeight)
	for cell := 0; cell < cvsHeight; cell += 1 + labelSpace {
		//v := float64(cell) / scale
		labels = append(labels, &Label{
			Text: "",
			Pos:  image.Point{0, cvsHeight - cell - 1},

			// TODO: align.
		})
	}

	return &YDetails{
		Width:  w,
		Labels: labels,
	}, nil
}
