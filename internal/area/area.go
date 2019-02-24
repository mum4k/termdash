// Copyright 2018 Google Inc.
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

// Package area provides functions working with image areas.
package area

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/internal/numbers"
)

// Size returns the size of the provided area.
func Size(area image.Rectangle) image.Point {
	return image.Point{
		area.Dx(),
		area.Dy(),
	}
}

// FromSize returns the corresponding area for the provided size.
func FromSize(size image.Point) (image.Rectangle, error) {
	if size.X < 0 || size.Y < 0 {
		return image.Rectangle{}, fmt.Errorf("cannot convert zero or negative size to an area, got: %+v", size)
	}
	return image.Rect(0, 0, size.X, size.Y), nil
}

// HSplit returns two new areas created by splitting the provided area at the
// specified percentage of its width. The percentage must be in the range
// 0 <= heightPerc <= 100.
// Can return zero size areas.
func HSplit(area image.Rectangle, heightPerc int) (top image.Rectangle, bottom image.Rectangle, err error) {
	if min, max := 0, 100; heightPerc < min || heightPerc > max {
		return image.ZR, image.ZR, fmt.Errorf("invalid heightPerc %d, must be in range %d <= heightPerc <= %d", heightPerc, min, max)
	}
	height := area.Dy() * heightPerc / 100
	top = image.Rect(area.Min.X, area.Min.Y, area.Max.X, area.Min.Y+height)
	if top.Dy() == 0 {
		top = image.ZR
	}
	bottom = image.Rect(area.Min.X, area.Min.Y+height, area.Max.X, area.Max.Y)
	if bottom.Dy() == 0 {
		bottom = image.ZR
	}
	return top, bottom, nil
}

// VSplit returns two new areas created by splitting the provided area at the
// specified percentage of its width. The percentage must be in the range
// 0 <= widthPerc <= 100.
// Can return zero size areas.
func VSplit(area image.Rectangle, widthPerc int) (left image.Rectangle, right image.Rectangle, err error) {
	if min, max := 0, 100; widthPerc < min || widthPerc > max {
		return image.ZR, image.ZR, fmt.Errorf("invalid widthPerc %d, must be in range %d <= widthPerc <= %d", widthPerc, min, max)
	}
	width := area.Dx() * widthPerc / 100
	left = image.Rect(area.Min.X, area.Min.Y, area.Min.X+width, area.Max.Y)
	if left.Dx() == 0 {
		left = image.ZR
	}
	right = image.Rect(area.Min.X+width, area.Min.Y, area.Max.X, area.Max.Y)
	if right.Dx() == 0 {
		right = image.ZR
	}
	return left, right, nil
}

// ExcludeBorder returns a new area created by subtracting a border around the
// provided area. Return the zero area if there isn't enough space to exclude
// the border.
func ExcludeBorder(area image.Rectangle) image.Rectangle {
	// If the area dimensions are smaller than this, subtracting a point for the
	// border on each of its sides results in a zero area.
	const minDim = 2
	if area.Dx() < minDim || area.Dy() < minDim {
		return image.ZR
	}
	return image.Rect(
		numbers.Abs(area.Min.X+1),
		numbers.Abs(area.Min.Y+1),
		numbers.Abs(area.Max.X-1),
		numbers.Abs(area.Max.Y-1),
	)
}

// WithRatio returns the largest area that has the requested ratio but is
// either equal or smaller than the provided area. Returns zero area if the
// area or the ratio are zero, or if there is no such area.
func WithRatio(area image.Rectangle, ratio image.Point) image.Rectangle {
	ratio = numbers.SimplifyRatio(ratio)
	if area == image.ZR || ratio == image.ZP {
		return image.ZR
	}

	wFact := area.Dx() / ratio.X
	hFact := area.Dy() / ratio.Y

	var fact int
	if wFact < hFact {
		fact = wFact
	} else {
		fact = hFact
	}
	return image.Rect(
		area.Min.X,
		area.Min.Y,
		ratio.X*fact+area.Min.X,
		ratio.Y*fact+area.Min.Y,
	)
}
