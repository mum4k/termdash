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

package numbers

import (
	"fmt"
	"image"
	"math"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestRoundToNonZeroPlaces(t *testing.T) {
	tests := []struct {
		float      float64
		places     int
		wantFloat  float64
		wantPlaces int
	}{
		{0, 0, 0, 0},
		{1.1, 0, 1.1, 0},
		{-1, 1, -1, 0},
		{1, 1, 1, 0},
		{1, 10, 1, 0},
		{1, -1, 1, 0},
		{0.12345, 2, 0.13, 0},
		{0.12345, -2, 0.13, 0},
		{0.12345, 10, 0.12345, 0},
		{0.00012345, 2, 0.00013, 3},
		{0.00012345, 3, 0.000124, 3},
		{0.00012345, 10, 0.00012345, 3},
		{-0.00012345, 10, -0.00012345, 3},
		{1.234567, 2, 1.24, 0},
		{-1.234567, 2, -1.23, 0},
		{1099.0000234567, 3, 1099.0000235, 4},
		{-1099.0000234567, 3, -1099.0000234, 4},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v_%v", tc.float, tc.places), func(t *testing.T) {
			gotFloat, gotPlaces := RoundToNonZeroPlaces(tc.float, tc.places)
			if gotFloat != tc.wantFloat || gotPlaces != tc.wantPlaces {
				t.Errorf("RoundToNonZeroPlaces(%v, %d) => (%v, %v), want (%v, %v)", tc.float, tc.places, gotFloat, gotPlaces, tc.wantFloat, tc.wantPlaces)
			}
		})
	}
}

func TestZeroBeforeDecimal(t *testing.T) {
	tests := []struct {
		float float64
		want  float64
	}{
		{0, 0},
		{-1, 0},
		{1, 0},
		{1.0, 0},
		{1.123, 0.123},
		{-1.123, -0.123},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprint(tc.float), func(t *testing.T) {
			got := zeroBeforeDecimal(tc.float)
			if got != tc.want {
				t.Errorf("zeroBeforeDecimal(%v) => %v, want %v", tc.float, got, tc.want)

			}
		})
	}
}

// Copied from the math package of Go 1.10 for backwards compatibility with Go
// 1.8 where the math.Round function doesn't exist yet.
func tolerance(a, b, e float64) bool {
	// Multiplying by e here can underflow denormal values to zero.
	// Check a==b so that at least if a and b are small and identical
	// we say they match.
	if a == b {
		return true
	}
	d := a - b
	if d < 0 {
		d = -d
	}

	// note: b is correct (expected) value, a is actual value.
	// make error tolerance a fraction of b, not a.
	if b != 0 {
		e = e * b
		if e < 0 {
			e = -e
		}
	}
	return d < e
}
func close(a, b float64) bool      { return tolerance(a, b, 1e-14) }
func veryclose(a, b float64) bool  { return tolerance(a, b, 4e-16) }
func soclose(a, b, e float64) bool { return tolerance(a, b, e) }
func alike(a, b float64) bool {
	switch {
	case math.IsNaN(a) && math.IsNaN(b):
		return true
	case a == b:
		return math.Signbit(a) == math.Signbit(b)
	}
	return false
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		desc    string
		values  []float64
		wantMin float64
		wantMax float64
	}{
		{
			desc: "no values",
		},
		{
			desc:    "all values the same",
			values:  []float64{1.1, 1.1},
			wantMin: 1.1,
			wantMax: 1.1,
		},
		{
			desc:    "all values the same and negative",
			values:  []float64{-1.1, -1.1},
			wantMin: -1.1,
			wantMax: -1.1,
		},
		{
			desc:    "min and max among positive values",
			values:  []float64{1.1, 1.2, 1.3},
			wantMin: 1.1,
			wantMax: 1.3,
		},
		{
			desc:    "min and max among positive and zero values",
			values:  []float64{1.1, 0, 1.3},
			wantMin: 0,
			wantMax: 1.3,
		},
		{
			desc:    "min and max among negative, positive and zero values",
			values:  []float64{1.1, 0, 1.3, -11.3, 22.5},
			wantMin: -11.3,
			wantMax: 22.5,
		},
		{
			desc:    "min and max among negative, positive, zero and NaN values",
			values:  []float64{1.1, 0, 1.3, math.NaN(), -11.3, 22.5},
			wantMin: -11.3,
			wantMax: 22.5,
		},
		{
			desc:    "all NaN values",
			values:  []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN()},
			wantMin: math.NaN(),
			wantMax: math.NaN(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotMin, gotMax := MinMax(tc.values)
			if diff := pretty.Compare(tc.wantMin, gotMin); diff != "" {
				t.Errorf("MinMax => unexpected min, diff (-want, +got):\n %s", diff)
			}
			if diff := pretty.Compare(tc.wantMax, gotMax); diff != "" {
				t.Errorf("MinMax => unexpected max, diff (-want, +got):\n %s", diff)
			}
		})
	}
}

func TestMinMaxInts(t *testing.T) {
	tests := []struct {
		desc    string
		values  []int
		wantMin int
		wantMax int
	}{
		{
			desc: "no values",
		},
		{
			desc:    "all values the same",
			values:  []int{1, 1},
			wantMin: 1,
			wantMax: 1,
		},
		{
			desc:    "all values the same and negative",
			values:  []int{-1, -1},
			wantMin: -1,
			wantMax: -1,
		},
		{
			desc:    "min and max among positive values",
			values:  []int{1, 2, 3},
			wantMin: 1,
			wantMax: 3,
		},
		{
			desc:    "min and max among positive and zero values",
			values:  []int{1, 0, 3},
			wantMin: 0,
			wantMax: 3,
		},
		{
			desc:    "min and max among negative, positive and zero values",
			values:  []int{1, 0, 3, -11, 22},
			wantMin: -11,
			wantMax: 22,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotMin, gotMax := MinMaxInts(tc.values)
			if gotMin != tc.wantMin || gotMax != tc.wantMax {
				t.Errorf("MinMaxInts => (%v, %v), want (%v, %v)", gotMin, gotMax, tc.wantMin, tc.wantMax)
			}
		})
	}
}

func TestDegreesToRadiansAndViceVersa(t *testing.T) {
	tests := []struct {
		degrees int
		want    float64
	}{
		{0, 0},
		{1, 0.017453292519943295},
		{-1, -0.017453292519943295},
		{15, 0.2617993877991494},
		{90, 1.5707963267948966},
		{180, 3.141592653589793},
		{270, 4.71238898038469},
		{360, 6.283185307179586},
		{361, 0.017453292519943295},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("degrees %d", tc.degrees), func(t *testing.T) {
			got := DegreesToRadians(tc.degrees)
			if !veryclose(got, tc.want) {
				t.Errorf("DegreesToRadians(%v) => %v, want %v", tc.degrees, got, tc.want)
			}
		})
	}
}

func TestRadiansToDegrees(t *testing.T) {
	tests := []struct {
		radians float64
		want    int
	}{
		{0, 0},
		{0.017453292519943295, 1},
		{-0.017453292519943295, 359},
		{-1.5707963267948966, 270},
		{0.2617993877991494, 15},
		{1.5707963267948966, 90},
		{3.141592653589793, 180},
		{4.71238898038469, 270},
		{6.283185307179586, 360},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("radians %v", tc.radians), func(t *testing.T) {
			got := RadiansToDegrees(tc.radians)
			if got != tc.want {
				t.Errorf("RadiansToDegrees(%v) => %v, want %v", tc.radians, got, tc.want)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{-1, 1},
		{-2, 2},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.input), func(t *testing.T) {
			got := Abs(tc.input)
			if got != tc.want {
				t.Errorf("Abs(%d) => %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestFindGCF(t *testing.T) {
	tests := []struct {
		a    int
		b    int
		want int
	}{
		{0, 0, 0},
		{0, 1, 0},
		{1, 0, 0},
		{1, 1, 1},
		{2, 2, 2},
		{50, 35, 5},
		{16, 88, 8},
		{-16, 88, 8},
		{16, -88, 8},
		{-16, -88, 8},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("findGCF(%d,%d)", tc.a, tc.b), func(t *testing.T) {
			if got := findGCF(tc.a, tc.b); got != tc.want {
				t.Errorf("findGCF(%d,%d) => got %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestSimplifyRatio(t *testing.T) {
	tests := []struct {
		desc  string
		ratio image.Point
		want  image.Point
	}{
		{
			desc:  "zero ratio",
			ratio: image.Point{0, 0},
			want:  image.Point{0, 0},
		},
		{
			desc:  "already simplified",
			ratio: image.Point{1, 3},
			want:  image.Point{1, 3},
		},
		{
			desc:  "already simplified and X is negative",
			ratio: image.Point{-1, 3},
			want:  image.Point{-1, 3},
		},
		{
			desc:  "already simplified and Y is negative",
			ratio: image.Point{1, -3},
			want:  image.Point{1, -3},
		},
		{
			desc:  "already simplified and both are negative",
			ratio: image.Point{-1, -3},
			want:  image.Point{-1, -3},
		},
		{
			desc:  "simplifies positive ratio",
			ratio: image.Point{27, 42},
			want:  image.Point{9, 14},
		},
		{
			desc:  "simplifies negative ratio",
			ratio: image.Point{-30, 50},
			want:  image.Point{-3, 5},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := SimplifyRatio(tc.ratio)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("SimplifyRatio => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSplitByRatio(t *testing.T) {
	tests := []struct {
		desc   string
		number int
		ratio  image.Point
		want   image.Point
	}{
		{
			desc:   "zero numerator",
			number: 10,
			ratio:  image.Point{0, 2},
			want:   image.ZP,
		},
		{
			desc:   "zero denominator",
			number: 10,
			ratio:  image.Point{2, 0},
			want:   image.ZP,
		},
		{
			desc:   "zero number",
			number: 0,
			ratio:  image.Point{1, 2},
			want:   image.ZP,
		},
		{
			desc:   "equal ratio",
			number: 2,
			ratio:  image.Point{2, 2},
			want:   image.Point{1, 1},
		},
		{
			desc:   "unequal ratio",
			number: 15,
			ratio:  image.Point{1, 2},
			want:   image.Point{5, 10},
		},
		{
			desc:   "large ratio",
			number: 19,
			ratio:  image.Point{78, 121},
			want:   image.Point{7, 12},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := SplitByRatio(tc.number, tc.ratio)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("SplitByRatio => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSumInts(t *testing.T) {
	tests := []struct {
		desc   string
		values []int
		want   int
	}{
		{
			desc: "empty list",
		},
		{
			desc:   "all values are zero",
			values: []int{0, 0, 0},
		},
		{
			desc:   "positive values",
			values: []int{1, 2, 3},
			want:   6,
		},
		{
			desc:   "negative values",
			values: []int{-1, -2, -3},
			want:   -6,
		},
		{
			desc:   "positive and negative values",
			values: []int{1, -2, 3},
			want:   2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := SumInts(tc.values)
			if got != tc.want {
				t.Errorf("SumInts(%v) => %v, want %v", tc.values, got, tc.want)
			}
		})
	}
}
