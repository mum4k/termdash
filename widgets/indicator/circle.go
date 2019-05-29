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

package indicator

// circle.go assists in calculation of points and angles on a circle.

import (
	"image"
)

// midAndRadius given an area of a braille canvas, determines the mid point in
// pixels and radius to draw the largest circle that fits.
// The circle's mid point is always positioned on the {0,1} pixel in the chosen
// cell so that any text inside of it can be visually centered.
func midAndRadius(ar image.Rectangle) (image.Point, int) {
	mid := image.Point{ar.Dx() / 2, ar.Dy() / 2}
	if mid.X%2 != 0 {
		mid.X--
	}
	switch mid.Y % 4 {
	case 0:
		mid.Y++
	case 1:
	case 2:
		mid.Y--
	case 3:
		mid.Y -= 2

	}

	// Calculate radius based on the smaller axis.
	var radius int
	if ar.Dx() < ar.Dy() {
		if mid.X < ar.Dx()/2 {
			radius = mid.X
		} else {
			radius = ar.Dx() - mid.X - 1
		}
	} else {
		if mid.Y < ar.Dy()/2 {
			radius = mid.Y
		} else {
			radius = ar.Dy() - mid.Y - 1
		}
	}
	return mid, radius
}
