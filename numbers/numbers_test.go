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
	"math"
	"testing"
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

func alike(a, b float64) bool {
	switch {
	case math.IsNaN(a) && math.IsNaN(b):
		return true
	case a == b:
		return math.Signbit(a) == math.Signbit(b)
	}
	return false
}

var round = []float64{
	5,
	8,
	math.Copysign(0, -1),
	-5,
	10,
	3,
	5,
	3,
	2,
	-9,
}

var vf = []float64{
	4.9790119248836735e+00,
	7.7388724745781045e+00,
	-2.7688005719200159e-01,
	-5.0106036182710749e+00,
	9.6362937071984173e+00,
	2.9263772392439646e+00,
	5.2290834314593066e+00,
	2.7279399104360102e+00,
	1.8253080916808550e+00,
	-8.6859247685756013e+00,
}

var vfroundSC = [][2]float64{
	{0, 0},
	{1.390671161567e-309, 0}, // denormal
	{0.49999999999999994, 0}, // 0.5-epsilon
	{0.5, 1},
	{0.5000000000000001, 1}, // 0.5+epsilon
	{-1.5, -2},
	{-2.5, -3},
	{math.NaN(), math.NaN()},
	{math.Inf(1), math.Inf(1)},
	{2251799813685249.5, 2251799813685250}, // 1 bit fraction
	{2251799813685250.5, 2251799813685251},
	{4503599627370495.5, 4503599627370496}, // 1 bit fraction, rounding to 0 bit fraction
	{4503599627370497, 4503599627370497},   // large integer
}

func TestRound(t *testing.T) {
	for i := 0; i < len(vf); i++ {
		if f := Round(vf[i]); !alike(round[i], f) {
			t.Errorf("Round(%g) = %g, want %g", vf[i], f, round[i])
		}
	}
	for i := 0; i < len(vfroundSC); i++ {
		if f := Round(vfroundSC[i][0]); !alike(vfroundSC[i][1], f) {
			t.Errorf("Round(%g) = %g, want %g", vfroundSC[i][0], f, vfroundSC[i][1])
		}
	}
}
