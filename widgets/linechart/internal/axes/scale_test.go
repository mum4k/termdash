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
func mustNewYScale(min, max float64, graphHeight, nonZeroDecimals int, mode YScaleMode, valueFormatter func(float64) string) *YScale {
	s, err := NewYScale(min, max, graphHeight, nonZeroDecimals, mode, valueFormatter)
	if err != nil {
		panic(err)
	}
	return s
}

// mustNewXScale returns a new XScale or panics.
func mustNewXScale(min, max int, graphWidth, nonZeroDecimals int) *XScale {
	s, err := NewXScale(min, max, graphWidth, nonZeroDecimals)
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
		graphHeight       int
		nonZeroDecimals   int
		mode              YScaleMode
		pixelToValueTests []pixelToValueTest
		valueToPixelTests []valueToPixelTest
		cellLabelTests    []cellLabelTest
		wantErr           bool
	}{
		{
			desc:            "fails when max is less than min",
			min:             0,
			max:             -1,
			graphHeight:     4,
			nonZeroDecimals: 2,
			wantErr:         true,
		},
		{
			desc:            "fails when canvas height too small",
			min:             0,
			max:             1,
			graphHeight:     0,
			nonZeroDecimals: 2,
			wantErr:         true,
		},
		{
			desc:            "fails on negative pixel",
			min:             0,
			max:             10,
			graphHeight:     4,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{-1, 0, true},
			},
		},
		{
			desc:            "fails on pixel out of range",
			min:             0,
			max:             10,
			graphHeight:     4,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{16, 0, true},
			},
		},
		{
			desc:            "fails on value or cell too small",
			min:             -1,
			max:             0,
			graphHeight:     4,
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
			graphHeight:     4,
			nonZeroDecimals: 2,
			valueToPixelTests: []valueToPixelTest{
				{1, 0, true},
			},
			cellLabelTests: []cellLabelTest{
				{4, nil, true},
			},
		},
		{
			desc:            "fails on an unsupported Y scale mode",
			min:             0,
			max:             0,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleMode(-1),
			wantErr:         true,
		},
		{
			desc:            "works without data points",
			min:             0,
			max:             0,
			graphHeight:     1,
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
			desc:            "min and max are non-zero positive and equal, scale is anchored",
			min:             6,
			max:             6,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAnchored,
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
			desc:            "min and max are non-zero positive and equal, scale is adaptive",
			min:             6,
			max:             6,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAdaptive,
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
			desc:            "min and max are non-zero negative and equal, scale is anchored",
			min:             -6,
			max:             -6,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAnchored,
			pixelToValueTests: []pixelToValueTest{
				{3, -6, false},
				{2, -4, false},
				{1, -2, false},
				{0, 0, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, false},
				{0.5, 0, false},
				{-1, 0, false},
				{-1.5, 1, false},
				{-2, 1, false},
				{-4, 2, false},
				{-6, 3, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(-6, 2), false},
			},
		},
		{
			desc:            "min and max are non-zero negative and equal, scale is adaptive",
			min:             -6,
			max:             -6,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAdaptive,
			pixelToValueTests: []pixelToValueTest{
				{3, -6, false},
				{2, -4, false},
				{1, -2, false},
				{0, 0, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, false},
				{0.5, 0, false},
				{-1, 0, false},
				{-1.5, 1, false},
				{-2, 1, false},
				{-4, 2, false},
				{-6, 3, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(-6, 2), false},
			},
		},
		{
			desc:            "min is non-zero positive, not equal to max, scale is anchored",
			min:             1,
			max:             7,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAnchored,
			pixelToValueTests: []pixelToValueTest{
				{3, 0, false},
				{2, 2.34, false},
				{1, 4.68, false},
				{0, 7, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 3, false},
				{0.5, 3, false},
				{1, 3, false},
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
			desc:            "min is non-zero positive, not equal to max, scale is adaptive",
			min:             1,
			max:             7,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAdaptive,
			pixelToValueTests: []pixelToValueTest{
				{3, 1, false},
				{2, 3, false},
				{1, 5, false},
				{0, 7, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 3, true},
				{0.5, 3, false},
				{1, 3, false},
				{1.5, 3, false},
				{2, 2, false},
				{4, 1, false},
				{6, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(1, 2), false},
			},
		},
		{
			desc:            "integer scale",
			min:             0,
			max:             6,
			graphHeight:     1,
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
			graphHeight:     2,
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
			graphHeight:     1,
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
			graphHeight:     1,
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
			desc:            "negative integer scale, max is also negative, scale is anchored",
			min:             -6,
			max:             -1,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAnchored,
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
				{-1, 0, false},
				{0, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(-6, 2), false},
			},
		},
		{
			desc:            "negative integer scale, max is also negative, scale is adaptive",
			min:             -7,
			max:             -1,
			graphHeight:     1,
			nonZeroDecimals: 2,
			mode:            YScaleModeAdaptive,
			pixelToValueTests: []pixelToValueTest{
				{3, -7, false},
				{2, -5, false},
				{1, -3, false},
				{0, -1, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{-7, 3, false},
				{-4, 1, false},
				{-2, 0, false},
				{-1, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(-7, 2), false},
			},
		},
		{
			desc:            "anchored based float scale",
			min:             0,
			max:             0.3,
			graphHeight:     1,
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
		{
			desc:            "requested value is negative, rounded isn't",
			min:             -0.19866933079506122,
			max:             0.19866933079506122,
			graphHeight:     28,
			nonZeroDecimals: 2,
			valueToPixelTests: []valueToPixelTest{
				{-0.19866933079506122, 111, false},
			},
			pixelToValueTests: []pixelToValueTest{
				{111, -0.19, false},
			},
		},
		{
			desc:            "regression for #92, positive values only, scale is anchored",
			min:             1600,
			max:             1900,
			graphHeight:     4,
			nonZeroDecimals: 2,
			mode:            YScaleModeAnchored,
			pixelToValueTests: []pixelToValueTest{
				{15, 0, false},
				{14, 126.67, false},
				{2, 1646.71, false},
				{1, 1773.38, false},
				{0, 1900, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 15, false},
				{126, 14, false},
				{1600, 2, false},
				{1800, 1, false},
				{1900, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{3, NewValue(0, 2), false},
				{2, NewValue(506.68, 2), false},
				{1, NewValue(1013.36, 2), false},
				{0, NewValue(1520.04, 2), false},
			},
		},
		{
			desc:            "regression for #92, positive values only, scale is adaptive",
			min:             1600,
			max:             1900,
			graphHeight:     4,
			nonZeroDecimals: 2,
			mode:            YScaleModeAdaptive,
			pixelToValueTests: []pixelToValueTest{
				{15, 1600, false},
				{14, 1620, false},
				{13, 1640, false},
				{12, 1660, false},
				{11, 1680, false},
				{10, 1700, false},
				{9, 1720, false},
				{8, 1740, false},
				{7, 1760, false},
				{6, 1780, false},
				{5, 1800, false},
				{4, 1820, false},
				{3, 1840, false},
				{2, 1860, false},
				{1, 1880, false},
				{0, 1900, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{1590, 0, true},
				{1600, 15, false},
				{1620, 14, false},
				{1640, 13, false},
				{1660, 12, false},
				{1680, 11, false},
				{1700, 10, false},
				{1720, 9, false},
				{1740, 8, false},
				{1760, 7, false},
				{1780, 6, false},
				{1800, 5, false},
				{1820, 4, false},
				{1840, 3, false},
				{1860, 2, false},
				{1880, 1, false},
				{1900, 0, false},
				{1910, 0, true},
			},
			cellLabelTests: []cellLabelTest{
				{3, NewValue(1600, 2), false},
				{2, NewValue(1680, 2), false},
				{1, NewValue(1760, 2), false},
				{0, NewValue(1840, 2), false},
			},
		},
		{
			desc:            "regression for #92, negative values only, scale is anchored",
			min:             -1900,
			max:             -1600,
			graphHeight:     4,
			nonZeroDecimals: 2,
			mode:            YScaleModeAnchored,
			pixelToValueTests: []pixelToValueTest{
				{15, -1900, false},
				{14, -1773.33, false},
				{5, -633.3, false},
				{0, 0, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{-1900, 15, false},
				{-1800, 14, false},
				{-633.3, 5, false},
				{0, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{3, NewValue(-1900, 2), false},
				{2, NewValue(-1393.32, 2), false},
				{1, NewValue(-886.64, 2), false},
			},
		},
		{
			desc:            "regression for #92, negative values only, scale is adaptiove",
			min:             -1900,
			max:             -1600,
			graphHeight:     4,
			nonZeroDecimals: 2,
			mode:            YScaleModeAdaptive,
			pixelToValueTests: []pixelToValueTest{
				{15, -1900, false},
				{14, -1880, false},
				{13, -1860, false},
				{12, -1840, false},
				{11, -1820, false},
				{10, -1800, false},
				{9, -1780, false},
				{8, -1760, false},
				{7, -1740, false},
				{6, -1720, false},
				{5, -1700, false},
				{4, -1680, false},
				{3, -1660, false},
				{2, -1640, false},
				{1, -1620, false},
				{0, -1600, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{-1910, 15, true},
				{-1900, 15, false},
				{-1880, 14, false},
				{-1860, 13, false},
				{-1840, 12, false},
				{-1820, 11, false},
				{-1800, 10, false},
				{-1780, 9, false},
				{-1760, 8, false},
				{-1740, 7, false},
				{-1720, 6, false},
				{-1700, 5, false},
				{-1680, 4, false},
				{-1660, 3, false},
				{-1640, 2, false},
				{-1620, 1, false},
				{-1600, 0, false},
				{-1590, 0, true},
			},
			cellLabelTests: []cellLabelTest{
				{3, NewValue(-1900, 2), false},
				{2, NewValue(-1820, 2), false},
				{1, NewValue(-1740, 2), false},
				{0, NewValue(-1660, 2), false},
			},
		},
		{
			desc:            "regression for #92, negative and positive values, scale is adaptive",
			min:             -100,
			max:             200,
			graphHeight:     4,
			nonZeroDecimals: 2,
			mode:            YScaleModeAdaptive,
			pixelToValueTests: []pixelToValueTest{
				{15, -100, false},
				{14, -80, false},
				{13, -60, false},
				{12, -40, false},
				{11, -20, false},
				{10, 0, false},
				{9, 20, false},
				{8, 40, false},
				{7, 60, false},
				{6, 80, false},
				{5, 100, false},
				{4, 120, false},
				{3, 140, false},
				{2, 160, false},
				{1, 180, false},
				{0, 200, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{-100, 15, false},
				{-80, 14, false},
				{-60, 13, false},
				{-40, 12, false},
				{-20, 11, false},
				{0, 10, false},
				{20, 9, false},
				{40, 8, false},
				{60, 7, false},
				{80, 6, false},
				{100, 5, false},
				{120, 4, false},
				{140, 3, false},
				{160, 2, false},
				{180, 1, false},
				{200, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{3, NewValue(-100, 2), false},
				{2, NewValue(-20, 2), false},
				{1, NewValue(60, 2), false},
				{0, NewValue(140, 2), false},
			},
		},
		{
			desc:            "regression for #92, negative and positive values, scale is anchored",
			min:             -100,
			max:             200,
			graphHeight:     4,
			nonZeroDecimals: 2,
			mode:            YScaleModeAnchored,
			pixelToValueTests: []pixelToValueTest{
				{15, -100, false},
				{14, -80, false},
				{13, -60, false},
				{12, -40, false},
				{11, -20, false},
				{10, 0, false},
				{9, 20, false},
				{8, 40, false},
				{7, 60, false},
				{6, 80, false},
				{5, 100, false},
				{4, 120, false},
				{3, 140, false},
				{2, 160, false},
				{1, 180, false},
				{0, 200, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{-100, 15, false},
				{-80, 14, false},
				{-60, 13, false},
				{-40, 12, false},
				{-20, 11, false},
				{0, 10, false},
				{20, 9, false},
				{40, 8, false},
				{60, 7, false},
				{80, 6, false},
				{100, 5, false},
				{120, 4, false},
				{140, 3, false},
				{160, 2, false},
				{180, 1, false},
				{200, 0, false},
			},
			cellLabelTests: []cellLabelTest{
				{3, NewValue(-100, 2), false},
				{2, NewValue(-20, 2), false},
				{1, NewValue(60, 2), false},
				{0, NewValue(140, 2), false},
			},
		},
	}

	for _, test := range tests {
		scale, err := NewYScale(test.min, test.max, test.graphHeight, test.nonZeroDecimals, test.mode, nil)
		if (err != nil) != test.wantErr {
			t.Errorf("NewYScale => unexpected error: %v, wantErr: %v", err, test.wantErr)
		}
		if err != nil {
			continue
		}
		t.Log(fmt.Sprintf("scale:%v", scale))

		t.Run(fmt.Sprintf("PixelToValue:%s", test.desc), func(t *testing.T) {
			for _, tc := range test.pixelToValueTests {
				got, err := scale.PixelToValue(tc.pixel)
				if (err != nil) != tc.wantErr {
					t.Errorf("PixelToValue(%v) => unexpected error: %v, wantErr: %v", tc.pixel, err, tc.wantErr)
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
					t.Errorf("ValueToPixel(%v) => unexpected error: %v, wantErr: %v", tc.value, err, tc.wantErr)
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
					t.Errorf("CellLabel(%v) => unexpected error: %v, wantErr: %v", tc.cell, err, tc.wantErr)
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
		min               int
		max               int
		graphWidth        int
		nonZeroDecimals   int
		pixelToValueTests []pixelToValueTest
		valueToPixelTests []valueToPixelTest
		valueToCellTests  []valueToCellTest
		cellLabelTests    []cellLabelTest
		wantErr           bool
	}{
		{
			desc:       "fails when min is negative",
			min:        -1,
			graphWidth: 1,
			wantErr:    true,
		},
		{
			desc:       "fails when max is negative",
			max:        -1,
			graphWidth: 1,
			wantErr:    true,
		},
		{
			desc:       "fails when min > max",
			min:        1,
			max:        0,
			graphWidth: 1,
			wantErr:    true,
		},
		{
			desc:       "fails when graphWidth zero",
			min:        0,
			max:        0,
			graphWidth: 0,
			wantErr:    true,
		},
		{
			desc:            "fails on negative pixel",
			min:             0,
			max:             0,
			graphWidth:      1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{-1, 0, true},
			},
		},
		{
			desc:            "fails on pixel out of range",
			min:             0,
			max:             0,
			graphWidth:      1,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{2, 0, true},
			},
		},
		{
			desc:            "fails on value or cell too small",
			min:             0,
			max:             0,
			graphWidth:      1,
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
			min:             0,
			max:             0,
			graphWidth:      1,
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
			min:             0,
			max:             0,
			graphWidth:      1,
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
			min:             0,
			max:             5,
			graphWidth:      3,
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
			desc:            "integer scale, min isn't zero",
			min:             1,
			max:             6,
			graphWidth:      3,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{0, 1, false},
				{1, 2, false},
				{2, 3, false},
				{3, 4, false},
				{4, 5, false},
				{5, 6, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, true},
				{1, 0, false},
				{2, 1, false},
				{3, 2, false},
				{4, 3, false},
				{5, 4, false},
				{6, 5, false},
			},
			valueToCellTests: []valueToCellTest{
				{0, 0, true},
				{1, 0, false},
				{2, 0, false},
				{3, 1, false},
				{4, 1, false},
				{5, 2, false},
				{6, 2, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(1, 2), false},
				{1, NewValue(3, 2), false},
				{2, NewValue(5, 2), false},
			},
		},
		{
			desc:            "float scale, multiple points per pixel",
			min:             0,
			max:             11,
			graphWidth:      3,
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
			desc:            "float scale, multiple points per pixel, min isn't zero",
			min:             1,
			max:             12,
			graphWidth:      3,
			nonZeroDecimals: 2,
			pixelToValueTests: []pixelToValueTest{
				{0, 1, false},
				{1, 3.21, false},
				{2, 5.42, false},
				{3, 7.63, false},
				{4, 9.84, false},
				{5, 12, false},
			},
			valueToPixelTests: []valueToPixelTest{
				{0, 0, true},
				{1, 0, false},
				{2, 0, false},
				{3, 1, false},
				{4, 1, false},
				{5, 2, false},
				{6, 2, false},
				{7, 3, false},
				{8, 3, false},
				{9, 4, false},
				{10, 4, false},
				{11, 5, false},
			},
			valueToCellTests: []valueToCellTest{
				{0, 0, true},
				{1, 0, false},
				{2, 0, false},
				{3, 0, false},
				{4, 0, false},
				{5, 1, false},
				{6, 1, false},
				{7, 1, false},
				{8, 1, false},
				{9, 2, false},
				{10, 2, false},
				{11, 2, false},
			},
			cellLabelTests: []cellLabelTest{
				{0, NewValue(1, 2), false},
				{1, NewValue(5, 2), false},
				{2, NewValue(10, 2), false},
			},
		},
		{
			desc:            "float scale, multiple pixels per point",
			min:             0,
			max:             1,
			graphWidth:      5,
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
		scale, err := NewXScale(test.min, test.max, test.graphWidth, test.nonZeroDecimals)
		if (err != nil) != test.wantErr {
			t.Errorf("NewXScale => unexpected error: %v, wantErr: %v", err, test.wantErr)
		}
		if err != nil {
			continue
		}
		t.Log(fmt.Sprintf("scale:%v", scale))

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
