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
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestYLabels(t *testing.T) {
	const nonZeroDecimals = 2
	tests := []struct {
		desc        string
		min         float64
		max         float64
		graphHeight int
		labelWidth  int
		want        []*Label
		wantErr     bool
	}{
		{
			desc:        "fails when canvas is too small",
			min:         0,
			max:         1,
			graphHeight: 1,
			labelWidth:  4,
			wantErr:     true,
		},
		{
			desc:        "fails when labelWidth is too small",
			min:         0,
			max:         1,
			graphHeight: 2,
			labelWidth:  -1,
			wantErr:     true,
		},
		{
			desc:        "works when there are no data points",
			min:         0,
			max:         0,
			graphHeight: 2,
			labelWidth:  1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
			},
		},
		{
			desc:        "only one label on tall canvas without data points",
			min:         0,
			max:         0,
			graphHeight: 25,
			labelWidth:  1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 24}},
			},
		},
		{
			desc:        "works when min equals max",
			min:         5,
			max:         5,
			graphHeight: 2,
			labelWidth:  1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
				{NewValue(2.88, nonZeroDecimals), image.Point{0, 0}},
			},
		},
		{
			desc:        "only two rows on the canvas, labels min and max",
			min:         0,
			max:         5,
			graphHeight: 2,
			labelWidth:  1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
				{NewValue(2.88, nonZeroDecimals), image.Point{0, 0}},
			},
		},
		{
			desc:        "aligns labels to the right",
			min:         0,
			max:         5,
			graphHeight: 2,
			labelWidth:  5,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{4, 1}},
				{NewValue(2.88, nonZeroDecimals), image.Point{1, 0}},
			},
		},
		{
			desc:        "multiple labels, last on the top",
			min:         0,
			max:         5,
			graphHeight: 9,
			labelWidth:  1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 8}},
				{NewValue(2.4, nonZeroDecimals), image.Point{0, 4}},
				{NewValue(4.8, nonZeroDecimals), image.Point{0, 0}},
			},
		},
		{
			desc:        "multiple labels, last on top-1",
			min:         0,
			max:         5,
			graphHeight: 10,
			labelWidth:  1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 9}},
				{NewValue(2.08, nonZeroDecimals), image.Point{0, 5}},
				{NewValue(4.16, nonZeroDecimals), image.Point{0, 1}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			scale, err := NewYScale(tc.min, tc.max, tc.graphHeight, nonZeroDecimals, YScaleModeAnchored, nil)
			if err != nil {
				t.Fatalf("NewYScale => unexpected error: %v", err)
			}
			t.Logf("scale step: %v", scale.Step.Rounded)
			got, err := yLabels(scale, tc.labelWidth)
			if (err != nil) != tc.wantErr {
				t.Errorf("yLabels => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("yLabels => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestXLabels(t *testing.T) {
	const nonZeroDecimals = 2
	tests := []struct {
		desc             string
		min              int
		max              int
		graphWidth       int
		graphZero        image.Point
		customLabels     map[int]string
		labelOrientation LabelOrientation
		want             []*Label
		wantErr          bool
	}{
		{
			desc:       "only one point",
			min:        0,
			max:        0,
			graphWidth: 1,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
			},
		},
		{
			desc:             "only one point, vertical",
			min:              0,
			max:              0,
			graphWidth:       1,
			graphZero:        image.Point{0, 1},
			labelOrientation: LabelOrientationVertical,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
			},
		},
		{
			desc:       "two points, only one label fits",
			min:        0,
			max:        1,
			graphWidth: 1,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
			},
		},
		{
			desc:       "two points, two labels fit exactly",
			min:        0,
			max:        1,
			graphWidth: 5,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(1, nonZeroDecimals), image.Point{4, 3}},
			},
		},
		{
			desc:             "two points, two labels fit exactly, vertical",
			min:              0,
			max:              1,
			graphWidth:       5,
			graphZero:        image.Point{0, 1},
			labelOrientation: LabelOrientationVertical,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(1, nonZeroDecimals), image.Point{4, 3}},
			},
		},
		{
			desc:       "labels are placed according to graphZero",
			min:        0,
			max:        1,
			graphWidth: 5,
			graphZero:  image.Point{3, 5},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{3, 7}},
				{NewValue(1, nonZeroDecimals), image.Point{7, 7}},
			},
		},
		{
			desc:       "skip to next value exhausts the space completely",
			min:        0,
			max:        10,
			graphWidth: 4,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
			},
		},
		{
			desc:       "second label doesn't fit due to its length",
			min:        0,
			max:        100,
			graphWidth: 5,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
			},
		},
		{
			desc:       "two points, two labels, more space than minSpacing so end label adjusted",
			min:        0,
			max:        1,
			graphWidth: 6,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(1, nonZeroDecimals), image.Point{5, 3}},
			},
		},
		{
			desc:       "at most as many labels as there are points",
			min:        0,
			max:        1,
			graphWidth: 100,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(1, nonZeroDecimals), image.Point{98, 3}},
			},
		},
		{
			desc:       "some labels in the middle",
			min:        0,
			max:        3,
			graphWidth: 100,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(1, nonZeroDecimals), image.Point{31, 3}},
				{NewValue(2, nonZeroDecimals), image.Point{62, 3}},
				{NewValue(3, nonZeroDecimals), image.Point{94, 3}},
			},
		},
		{
			desc:       "some labels in the middle, min isn't zero",
			min:        1,
			max:        4,
			graphWidth: 100,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(1, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(2, nonZeroDecimals), image.Point{31, 3}},
				{NewValue(3, nonZeroDecimals), image.Point{62, 3}},
				{NewValue(4, nonZeroDecimals), image.Point{94, 3}},
			},
		},
		{
			desc:       "custom labels provided",
			min:        0,
			max:        3,
			graphWidth: 100,
			graphZero:  image.Point{0, 1},
			customLabels: map[int]string{
				0: "a",
				1: "b",
				2: "c",
				3: "d",
			},
			want: []*Label{
				{NewTextValue("a"), image.Point{0, 3}},
				{NewTextValue("b"), image.Point{31, 3}},
				{NewTextValue("c"), image.Point{62, 3}},
				{NewTextValue("d"), image.Point{94, 3}},
			},
		},
		{
			desc:       "custom labels provided, min isn't zero",
			min:        1,
			max:        4,
			graphWidth: 100,
			graphZero:  image.Point{0, 1},
			customLabels: map[int]string{
				1: "a",
				2: "b",
				3: "c",
				4: "d",
			},
			want: []*Label{
				{NewTextValue("a"), image.Point{0, 3}},
				{NewTextValue("b"), image.Point{31, 3}},
				{NewTextValue("c"), image.Point{62, 3}},
				{NewTextValue("d"), image.Point{94, 3}},
			},
		},
		{
			desc:       "custom labels provided, but only some fit, regression for #117",
			min:        0,
			max:        7,
			graphWidth: 5,
			graphZero:  image.Point{0, 1},
			customLabels: map[int]string{
				0: "a",
				1: "b",
				2: "c",
				3: "d",
				4: "e",
				5: "f",
				6: "g",
				7: "h",
			},
			want: []*Label{
				{NewTextValue("a"), image.Point{0, 3}},
				{NewTextValue("g"), image.Point{4, 3}},
			},
		},
		{
			desc:       "only some custom labels provided",
			min:        0,
			max:        3,
			graphWidth: 100,
			graphZero:  image.Point{0, 1},
			customLabels: map[int]string{
				0: "a",
				3: "d",
			},
			want: []*Label{
				{NewTextValue("a"), image.Point{0, 3}},
				{NewValue(1, nonZeroDecimals), image.Point{31, 3}},
				{NewValue(2, nonZeroDecimals), image.Point{62, 3}},
				{NewTextValue("d"), image.Point{94, 3}},
			},
		},
		{
			desc:       "not displayed if custom labels don't fit",
			min:        0,
			max:        1,
			graphWidth: 6,
			graphZero:  image.Point{0, 1},
			customLabels: map[int]string{
				0: "a very very long custom label",
			},
			want: []*Label{},
		},
		{
			desc:       "more points than pixels",
			min:        0,
			max:        99,
			graphWidth: 6,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(72, nonZeroDecimals), image.Point{4, 3}},
			},
		},
		{
			desc:       "longer labels, only two fit in horizontal",
			min:        0,
			max:        1000,
			graphWidth: 10,
			graphZero:  image.Point{0, 1},
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(421, nonZeroDecimals), image.Point{4, 3}},
			},
		},
		{
			desc:             "longer labels, multiple fit in vertical",
			min:              0,
			max:              1000,
			graphWidth:       10,
			graphZero:        image.Point{0, 1},
			labelOrientation: LabelOrientationVertical,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 3}},
				{NewValue(421, nonZeroDecimals), image.Point{4, 3}},
				{NewValue(842, nonZeroDecimals), image.Point{8, 3}},
			},
		},
		{
			desc:       "longer custom labels, only one fits in horizontal",
			min:        0,
			max:        1000,
			graphWidth: 10,
			graphZero:  image.Point{0, 1},
			customLabels: map[int]string{
				0:   "zero label",
				421: "this one is even longer",
				841: "this label just keeps on going",
			},
			want: []*Label{
				{NewTextValue("zero label"), image.Point{0, 3}},
			},
		},
		{
			desc:       "longer custom labels, all fit in vertical",
			min:        0,
			max:        1000,
			graphWidth: 10,
			graphZero:  image.Point{0, 1},
			customLabels: map[int]string{
				0:   "zero label",
				421: "this one is even longer",
				842: "this label just keeps on going",
			},
			labelOrientation: LabelOrientationVertical,
			want: []*Label{
				{NewTextValue("zero label"), image.Point{0, 3}},
				{NewTextValue("this one is even longer"), image.Point{4, 3}},
				{NewTextValue("this label just keeps on going"), image.Point{8, 3}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			scale, err := NewXScale(tc.min, tc.max, tc.graphWidth, nonZeroDecimals)
			if err != nil {
				t.Fatalf("NewXScale => unexpected error: %v", err)
			}
			t.Logf("scale step: %v, label orientation: %v", scale.Step.Rounded, tc.labelOrientation)
			got, err := xLabels(scale, tc.graphZero, tc.customLabels, tc.labelOrientation)
			if (err != nil) != tc.wantErr {
				t.Errorf("xLabels => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("xLabels => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestXSpace(t *testing.T) {
	tests := []struct {
		desc          string
		graphZero     image.Point
		graphWidth    int
		sub           int
		wantRemaining int
		wantRelative  image.Point
		wantErr       bool
	}{
		{
			desc:       "fails to subtract when we run out of space",
			graphWidth: 1,
			sub:        2,
			wantErr:    true,
		},
		{
			desc:          "subtracts, graph is zero based",
			graphWidth:    2,
			sub:           1,
			wantRemaining: 1,
			wantRelative:  image.Point{1, 1},
		},
		{
			desc:          "subtracts, graph isn't zero based",
			graphZero:     image.Point{10, 10},
			graphWidth:    2,
			sub:           1,
			wantRemaining: 1,
			wantRelative:  image.Point{1, 11},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			xs := newXSpace(tc.graphZero, tc.graphWidth)
			t.Logf("xSpace: %v", xs)

			err := xs.Sub(tc.sub)
			if (err != nil) != tc.wantErr {
				t.Errorf("xSpace.Sub => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			gotRemaining := xs.Remaining()
			if gotRemaining != tc.wantRemaining {
				t.Errorf("xSpace.Remaining => %v, want %v", gotRemaining, tc.wantRemaining)
			}

			gotRelative := xs.Relative()
			if diff := pretty.Compare(tc.wantRelative, gotRelative); diff != "" {
				t.Errorf("xSpace.Relative => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
