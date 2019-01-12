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

// mustNewYScale returns a new YScale or panics.
func mustNewYScale(min, max float64, cvsHeight, nonZeroDecimals int) *YScale {
	s, err := NewYScale(min, max, cvsHeight, nonZeroDecimals)
	if err != nil {
		panic(err)
	}
	return s
}

// mustNewXScale returns a new XScale or panics.
func mustNewXScale(numPoints int, axisWidth, nonZeroDecimals int) *XScale {
	s, err := NewXScale(numPoints, axisWidth, nonZeroDecimals)
	if err != nil {
		panic(err)
	}
	return s
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

// valueToCellTest is a test case for ValueToCell.
type valueToCellTest struct {
	value   int
	want    int
	wantErr bool
}

// cellLabelTest is a test case for CellLabel.
type cellLabelTest struct {
	cell    int
	want    *Value
	wantErr bool
}

func TestYScale(t *testing.T) {
	tests := []struct {
		desc              string
		min               float64
		max               float64
		cvsHeight         int
		nonZeroDecimals   int
		pixelToValueTests []pixelToValueTest
		valueToPixelTests []valueToPixelTest
		cellLabelTests    []cellLabelTest
		wantErr           bool
	}{
		{
			desc:            "fails when max is less than min",
			min:             0,
			max:             -1,
			cvsHeight:       4,
			nonZeroDecimals: 2,
			wantErr:         true,
		},
		{
			desc:            "fails when canvas height too small",
			min:             0,
			max:             1,
			cvsHeight:       0,
			nonZeroDecimals: 2,
			wantErr:         true,
		},
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
			desc:            "works without data points",
			min:             0,
			max:             0,
			cvsHeight:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{0, 0, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(0, 2), false},
			},
		},
		{
			desc:            "min and max are non-zero positive and equal, scale is zero based",
			min:             6,
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
			desc:            "min is non-zero positive, not equal to max, scale is zero based",
			min:             1,
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
			desc:            "negative integer scale, max is also negative, scale has max of zero",
			min:             -6,
			max:             -1,
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
		scale, err := NewYScale(test.min, test.max, test.cvsHeight, test.nonZeroDecimals)
		if (err != nil) != test.wantErr {
			t.Errorf("NewYScale => unexpected error: %v, wantErr: %v", err, test.wantErr)
		}
		if err != nil {
			continue
		}

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

func TestXScale(t *testing.T) {
	tests := []struct {
		desc              string
		numPoints         int
		axisWidth         int
		nonZeroDecimals   int
		pixelToValueTests []pixelToValueTest
		valueToPixelTests []valueToPixelTest
		valueToCellTests  []valueToCellTest
		cellLabelTests    []cellLabelTest
		wantErr           bool
	}{
		{
			desc:      "fails when numPoints negative",
			numPoints: -1,
			axisWidth: 1,
			wantErr:   true,
		},
		{
			desc:      "fails when axisWidth zero",
			numPoints: 1,
			axisWidth: 0,
			wantErr:   true,
		},
		{
			desc:            "fails on negative pixel",
			numPoints:       1,
			axisWidth:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{-1, 0, true},
			},
		},
		{
			desc:            "fails on pixel out of range",
			numPoints:       1,
			axisWidth:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{2, 0, true},
			},
		},
		{
			desc:            "fails on value or cell too small",
			numPoints:       1,
			axisWidth:       1,
			nonZeroDecimals: 2,
			valueToPixelTests: []valueToPixelTest{
				{-1, 0, true},
			},
			valueToCellTests: []valueToCellTest{
				{-1, 0, true},
			},
			cellLabelTests: []cellLabelTest{
				{-1, nil, true},
			},
		},
		{
			desc:            "fails on value or cell too large",
			numPoints:       1,
			axisWidth:       1,
			nonZeroDecimals: 2,
			valueToPixelTests: []valueToPixelTest{
				{1, 0, true},
			},
			valueToCellTests: []valueToCellTest{
				{1, 0, true},
			},
			cellLabelTests: []cellLabelTest{
				{2, nil, true},
			},
		},
		{
			desc:            "works without data points",
			numPoints:       0,
			axisWidth:       1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{0, 0, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(0, 2), false},
			},
		},
		{
			desc:            "integer scale, all points fit",
			numPoints:       6,
			axisWidth:       3,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{0, 0, false},
				{1, 1, false},
				{2, 2, false},
				{3, 3, false},
				{4, 4, false},
				{5, 5, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, false},
				{1, 1, false},
				{2, 2, false},
				{3, 3, false},
				{4, 4, false},
				{5, 5, false},
			},
			valueToCellTests: []valueToCellTest{
				{0, 0, false},
				{1, 0, false},
				{2, 1, false},
				{3, 1, false},
				{4, 2, false},
				{5, 2, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(0, 2), false},
				{1, NewValue(2, 2), false},
				{2, NewValue(4, 2), false},
			},
		},
		{
			desc:            "float scale, multiple points per pixel",
			numPoints:       12,
			axisWidth:       3,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{0, 0, false},
				{1, 2.21, false},
				{2, 4.42, false},
				{3, 6.63, false},
				{4, 8.84, false},
				{5, 11, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, false},
				{1, 0, false},
				{2, 1, false},
				{3, 1, false},
				{4, 2, false},
				{5, 2, false},
				{6, 3, false},
				{7, 3, false},
				{8, 4, false},
				{9, 4, false},
				{10, 5, false},
				{11, 5, false},
			},
			valueToCellTests: []valueToCellTest{
				{0, 0, false},
				{1, 0, false},
				{2, 0, false},
				{3, 0, false},
				{4, 1, false},
				{5, 1, false},
				{6, 1, false},
				{7, 1, false},
				{8, 2, false},
				{9, 2, false},
				{10, 2, false},
				{11, 2, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(0, 2), false},
				{1, NewValue(4, 2), false},
				{2, NewValue(9, 2), false},
			},
		},
		{
			desc:            "float scale, multiple pixels per point",
			numPoints:       2,
			axisWidth:       5,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{0, 0, false},
				{1, 0.12, false},
				{2, 0.24, false},
				{3, 0.36, false},
				{4, 0.48, false},
				{5, 0.6, false},
				{6, 0.72, false},
				{7, 0.84, false},
				{8, 0.96, false},
				{9, 1, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, false},
				{1, 8, false},
			},
			valueToCellTests: []valueToCellTest{
				{0, 0, false},
				{1, 4, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(0, 2), false},
				{1, NewValue(0, 2), false},
				{2, NewValue(0, 2), false},
				{3, NewValue(1, 2), false},
				{4, NewValue(1, 2), false},
			},
		},
	}

	for _, test := range tests {
		scale, err := NewXScale(test.numPoints, test.axisWidth, test.nonZeroDecimals)
		if (err != nil) != test.wantErr {
			t.Errorf("NewXScale => unexpected error: %v, wantErr: %v", err, test.wantErr)
		}
		if err != nil {
			continue
		}

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
				got, err := scale.ValueToPixel(int(tc.value))
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

		t.Run(fmt.Sprintf("ValueToCell:%s", test.desc), func(t *testing.T) {
			for _, tc := range test.valueToCellTests {
				got, err := scale.ValueToCell(tc.value)
				if (err != nil) != tc.wantErr {
					t.Errorf("ValueToCell => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					continue
				}
				if got != tc.want {
					t.Errorf("ValueToCell(%v) => %v, want %v", tc.value, got, tc.want)
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
