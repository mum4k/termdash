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

// Package numbers implements numerical calculations required when drawing a line chart.
package numbers

import (
	"math"
)

// RoundToNonZeroPlaces rounds the float up, so that it has at least the provided
// number of non-zero decimal places.
// Returns the rounded float and the number of leading decimal places that
// are zero. Returns the original float when places is zero. Negative places
// are treated as positive, so that -2 == 2.
func RoundToNonZeroPlaces(f float64, places int) (float64, int) {
	if f == 0 {
		return 0, 0
	}

	decOnly := zeroBeforeDecimal(f)
	if decOnly == 0 {
		return f, 0
	}
	nzMult := multToNonZero(decOnly)
	if places == 0 {
		return f, multToPlaces(nzMult)
	}
	plMult := placesToMult(places)

	m := float64(nzMult * plMult)
	return math.Ceil(f*m) / m, multToPlaces(nzMult)
}

// multToNonZero returns multiplier for the float, so that the first decimal
// place is non-zero. The float must not be zero.
func multToNonZero(f float64) int {
	v := f
	if v < 0 {
		v *= -1
	}

	mult := 1
	for v < 0.1 {
		v *= 10
		mult *= 10
	}
	return mult
}

// placesToMult translates the number of decimal places to a multiple of 10.
func placesToMult(places int) int {
	if places < 0 {
		places *= -1
	}

	mult := 1
	for i := 0; i < places; i++ {
		mult *= 10
	}
	return mult
}

// multToPlaces translates the multiple of 10 to a number of decimal places.
func multToPlaces(mult int) int {
	places := 0
	for mult > 1 {
		mult /= 10
		places++
	}
	return places
}

// zeroBeforeDecimal modifies the float so that it only has zero value before
// the decimal point.
func zeroBeforeDecimal(f float64) float64 {
	var sign float64 = 1
	if f < 0 {
		f *= -1
		sign = -1
	}

	floor := math.Floor(f)
	return (f - floor) * sign
}
