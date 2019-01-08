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

package axes

import (
	"fmt"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestYScale(t *testing.T) {
	tests := []struct {
		desc            string
		min             float64
		max             float64
		cvsHeight       int
		nonZeroDecimals int
		want            *YScale
	}{
		{
			desc:            "min and max are zero",
			min:             0,
			max:             0,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			want: &YScale{
				Min:           NewValue(0, 2),
				Max:           NewValue(0, 2),
				Step:          NewValue(0, 2),
				CvsHeight:     4,
				brailleHeight: 16,
			},
		},
		{
			desc:            "zero based scale",
			min:             0,
			max:             10,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			want: &YScale{
				Min:           NewValue(0, 2),
				Max:           NewValue(10, 2),
				Step:          NewValue(float64(10)/15, 2),
				CvsHeight:     4,
				brailleHeight: 16,
			},
		},
		{
			desc:            "min is negative",
			min:             -10,
			max:             10,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			want: &YScale{
				Min:           NewValue(-10, 2),
				Max:           NewValue(10, 2),
				Step:          NewValue(float64(20)/15, 2),
				CvsHeight:     4,
				brailleHeight: 16,
			},
		},
		{
			desc:            "greater than one",
			min:             0,
			max:             5,
			cvsHeight:       1,
			nonZeroDecimals: 1,
			want: &YScale{
				Min:           NewValue(0, 1),
				Max:           NewValue(5, 1),
				Step:          NewValue(float64(5)/3, 1),
				CvsHeight:     1,
				brailleHeight: 4,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := NewYScale(tc.min, tc.max, tc.cvsHeight, tc.nonZeroDecimals)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewYScale => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

// pixelToValueTest is a test case for PixelToValue.
type pixelToValueTest struct {
	pixel   int
	want    float64
	wantErr bool
}

// valueToPixelTest is a test case for ValueToPixel.
type valueToPixelTest struct {
	value   float64
	want    int
	wantErr bool
}

// cellLabelTest is a test case for CellLabel.
type cellLabelTest struct {
	cell    int
	want    *Value
	wantErr bool
}

func TestPixelToValueAndViceVersa(t *testing.T) {
	tests := []struct {
		desc              string
		min               float64
		max               float64
		cvsHeight         int
		nonZeroDecimals   int
		pixelToValueTests []pixelToValueTest
		valueToPixelTests []valueToPixelTest
		cellLabelTests    []cellLabelTest
	}{
		{
			desc:            "fails on negative pixel",
			min:             0,
			max:             10,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{-1, 0, true},
			},
		},
		{
			desc:            "fails on pixel out of range",
			min:             0,
			max:             10,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{16, 0, true},
			},
		},
		{
			desc:            "fails on value or cell too small",
			min:             -1,
			max:             0,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			valueToPixelTests: []valueToPixelTest{
				{-2, 0, true},
			},
			cellLabelTests: []cellLabelTest{
				{-1, nil, true},
			},
		},
		{
			desc:            "fails on value or cell too large",
			min:             -1,
			max:             0,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			valueToPixelTests: []valueToPixelTest{
				{1, 0, true},
			},
			cellLabelTests: []cellLabelTest{
				{4, nil, true},
			},
		},
		{
			desc:            "integer scale",
			min:             0,
			max:             6,
			cvsHeight:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{3, 0, false},
				{2, 2, false},
				{1, 4, false},
				{0, 6, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 3, false},
				{0.5, 3, false},
				{1, 2, false},
				{1.5, 2, false},
				{2, 2, false},
				{4, 1, false},
				{6, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(0, 2), false},
			},
		},
		{
			desc:            "integer scale, multi-row canvas",
			min:             0,
			max:             14,
			cvsHeight:       2,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{7, 0, false},
				{6, 2, false},
				{5, 4, false},
				{4, 6, false},
				{3, 8, false},
				{2, 10, false},
				{1, 12, false},
				{0, 14, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 7, false},
				{1, 6, false},
				{4, 5, false},
				{6, 4, false},
				{14, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(8, 2), false},
				{1, NewValue(0, 2), false},
			},
		},
		{
			desc:            "negative integer scale",
			min:             -3,
			max:             3,
			cvsHeight:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{3, -3, false},
				{2, -1, false},
				{1, 1, false},
				{0, 3, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{-3, 3, false},
				{-2.5, 3, false},
				{-2, 2, false},
				{-1.5, 2, false},
				{-1, 2, false},
				{0, 1, false},
				{1, 1, false},
				{3, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(-3, 2), false},
			},
		},
		{
			desc:            "negative integer scale, max is zero",
			min:             -6,
			max:             0,
			cvsHeight:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{3, -6, false},
				{2, -4, false},
				{1, -2, false},
				{0, 0, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{-6, 3, false},
				{-4, 2, false},
				{-2, 1, false},
				{0, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(-6, 2), false},
			},
		},
		{
			desc:            "zero based float scale",
			min:             0,
			max:             0.3,
			cvsHeight:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{3, 0, false},
				{2, 0.1, false},
				{1, 0.2, false},
				{0, 0.3, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 3, false},
				{0.1, 2, false},
				{0.2, 1, false},
				{0.3, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(0, 2), false},
			},
		},
	}

	for _, test := range tests {
		scale := NewYScale(test.min, test.max, test.cvsHeight, test.nonZeroDecimals)
		t.Run(fmt.Sprintf("PixelToValue:%s", test.desc), func(t *testing.T) {
			for _, tc := range test.pixelToValueTests {
				got, err := scale.PixelToValue(tc.pixel)
				if (err != nil) != tc.wantErr {
					t.Errorf("PixelToValue => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					continue
				}
				if got != tc.want {
					t.Errorf("PixelToValue(%v) => %v, want %v", tc.pixel, got, tc.want)
				}
			}
		})

		t.Run(fmt.Sprintf("ValueToPixel:%s", test.desc), func(t *testing.T) {
			for _, tc := range test.valueToPixelTests {
				got, err := scale.ValueToPixel(tc.value)
				if (err != nil) != tc.wantErr {
					t.Errorf("ValueToPixel => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					continue
				}
				if got != tc.want {
					t.Errorf("ValueToPixel(%v) => %v, want %v", tc.value, got, tc.want)
				}
			}
		})

		t.Run(fmt.Sprintf("CellLabel:%s", test.desc), func(t *testing.T) {
			for _, tc := range test.cellLabelTests {
				got, err := scale.CellLabel(tc.cell)
				if (err != nil) != tc.wantErr {
					t.Errorf("CellLabel => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					continue
				}
				if diff := pretty.Compare(tc.want, got); diff != "" {
					t.Errorf("CellLabel(%v) => unexpected diff (-want, +got):\n%s", tc.cell, diff)
				}
			}
		})
	}
}
