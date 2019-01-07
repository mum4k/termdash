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
