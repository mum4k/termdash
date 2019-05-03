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
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

var (
	testValueFormatter = func(float64) string { return "test" }
)

type updateY struct {
	minVal float64
	maxVal float64
}

func TestY(t *testing.T) {
	tests := []struct {
		desc      string
		yp        *YProperties
		cvsAr     image.Rectangle
		wantWidth int
		want      *YDetails
		wantErr   bool
	}{
		{
			desc: "fails on canvas too small",
			yp: &YProperties{
				Min:        0,
				Max:        3,
				ReqXHeight: 2,
			},
			cvsAr:     image.Rect(0, 0, 3, 2),
			wantWidth: 2,
			wantErr:   true,
		},
		{
			desc: "fails on cvsWidth less than required width",
			yp: &YProperties{
				Min:        0,
				Max:        3,
				ReqXHeight: 2,
			},
			cvsAr:     image.Rect(0, 0, 2, 4),
			wantWidth: 2,
			wantErr:   true,
		},
		{
			desc: "fails when max is less than min",
			yp: &YProperties{
				Min:        0,
				Max:        -1,
				ReqXHeight: 2,
			},
			cvsAr:     image.Rect(0, 0, 4, 4),
			wantWidth: 3,
			wantErr:   true,
		},
		{
			desc: "cvsWidth equals required width",
			yp: &YProperties{
				Min:        0,
				Max:        3,
				ReqXHeight: 2,
			},
			cvsAr:     image.Rect(0, 0, 3, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 2,
				Start: image.Point{1, 0},
				End:   image.Point{1, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals, YScaleModeAnchored, nil),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc: "success for anchored scale",
			yp: &YProperties{
				Min:        1,
				Max:        3,
				ReqXHeight: 2,
				ScaleMode:  YScaleModeAnchored,
			},
			cvsAr:     image.Rect(0, 0, 3, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 2,
				Start: image.Point{1, 0},
				End:   image.Point{1, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals, YScaleModeAnchored, nil),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc: "accommodates X scale that needs more height",
			yp: &YProperties{
				Min:        1,
				Max:        3,
				ReqXHeight: 4,
				ScaleMode:  YScaleModeAnchored,
			},
			cvsAr:     image.Rect(0, 0, 3, 6),
			wantWidth: 2,
			want: &YDetails{
				Width: 2,
				Start: image.Point{1, 0},
				End:   image.Point{1, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals, YScaleModeAnchored, nil),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc: "success for adaptive scale",
			yp: &YProperties{
				Min:        1,
				Max:        6,
				ReqXHeight: 2,
				ScaleMode:  YScaleModeAdaptive,
			},
			cvsAr:     image.Rect(0, 0, 3, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 2,
				Start: image.Point{1, 0},
				End:   image.Point{1, 2},
				Scale: mustNewYScale(1, 6, 2, nonZeroDecimals, YScaleModeAdaptive, nil),
				Labels: []*Label{
					{NewValue(1, nonZeroDecimals), image.Point{0, 1}},
					{NewValue(3.88, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc: "cvsWidth just accommodates the longest label",
			yp: &YProperties{
				Min:        0,
				Max:        3,
				ReqXHeight: 2,
			},
			cvsAr:     image.Rect(0, 0, 6, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 5,
				Start: image.Point{4, 0},
				End:   image.Point{4, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals, YScaleModeAnchored, nil),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{3, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc: "cvsWidth is more than we need",
			yp: &YProperties{
				Min:        0,
				Max:        3,
				ReqXHeight: 2,
			},
			cvsAr:     image.Rect(0, 0, 7, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 5,
				Start: image.Point{4, 0},
				End:   image.Point{4, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals, YScaleModeAnchored, nil),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{3, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc: "success for formatted labels scale",
			yp: &YProperties{
				Min:            1,
				Max:            3,
				ReqXHeight:     2,
				ScaleMode:      YScaleModeAnchored,
				ValueFormatter: testValueFormatter,
			},
			cvsAr:     image.Rect(0, 0, 3, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 2,
				Start: image.Point{1, 0},
				End:   image.Point{1, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals, YScaleModeAnchored, testValueFormatter),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals, ValueFormatter(testValueFormatter)), image.Point{0, 1}},
					{NewValue(1.72, nonZeroDecimals, ValueFormatter(testValueFormatter)), image.Point{0, 0}},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotWidth := RequiredWidth(tc.yp.Min, tc.yp.Max)
			if gotWidth != tc.wantWidth {
				t.Errorf("RequiredWidth => got %v, want %v", gotWidth, tc.wantWidth)
			}

			got, err := NewYDetails(tc.cvsAr, tc.yp)
			if (err != nil) != tc.wantErr {
				t.Errorf("Details => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Details => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNewXDetails(t *testing.T) {
	tests := []struct {
		desc    string
		xp      *XProperties
		cvsAr   image.Rectangle
		want    *XDetails
		wantErr bool
	}{
		{
			desc: "fails when min is negative",
			xp: &XProperties{
				Min:       -1,
				Max:       0,
				ReqYWidth: 0,
			},
			cvsAr:   image.Rect(0, 0, 2, 3),
			wantErr: true,
		},
		{
			desc: "fails when cvsAr isn't wide enough",
			xp: &XProperties{
				Min:       0,
				Max:       0,
				ReqYWidth: 0,
			},
			cvsAr:   image.Rect(0, 0, 1, 3),
			wantErr: true,
		},
		{
			desc: "fails when cvsAr isn't tall enough",
			xp: &XProperties{
				Min:       0,
				Max:       0,
				ReqYWidth: 0,
			},
			cvsAr:   image.Rect(0, 0, 3, 2),
			wantErr: true,
		},
		{
			desc: "works with no data points",
			xp: &XProperties{
				Min:       0,
				Max:       0,
				ReqYWidth: 0,
			},
			cvsAr: image.Rect(0, 0, 2, 3),
			want: &XDetails{
				Start: image.Point{0, 1},
				End:   image.Point{1, 1},
				Scale: mustNewXScale(0, 0, 1, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewValue(0, nonZeroDecimals),
						Pos:   image.Point{1, 2},
					},
				},
				Properties: &XProperties{
					Min:       0,
					Max:       0,
					ReqYWidth: 0,
				},
			},
		},
		{
			desc: "works with no data points, vertical",
			xp: &XProperties{
				Min:       0,
				Max:       0,
				ReqYWidth: 0,
				LO:        LabelOrientationVertical,
			},
			cvsAr: image.Rect(0, 0, 2, 3),
			want: &XDetails{
				Start: image.Point{0, 1},
				End:   image.Point{1, 1},
				Scale: mustNewXScale(0, 0, 1, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewValue(0, nonZeroDecimals),
						Pos:   image.Point{1, 2},
					},
				},
				Properties: &XProperties{
					Min:       0,
					Max:       0,
					ReqYWidth: 0,
					LO:        LabelOrientationVertical,
				},
			},
		},
		{
			desc: "accounts for non-zero yStart",
			xp: &XProperties{
				Min:       0,
				Max:       0,
				ReqYWidth: 2,
			},
			cvsAr: image.Rect(0, 0, 4, 5),
			want: &XDetails{
				Start: image.Point{2, 3},
				End:   image.Point{3, 3},
				Scale: mustNewXScale(0, 0, 1, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewValue(0, nonZeroDecimals),
						Pos:   image.Point{3, 4},
					},
				},
				Properties: &XProperties{
					Min:       0,
					Max:       0,
					ReqYWidth: 2,
				},
			},
		},
		{
			desc: "accounts for longer vertical labels, the tallest didn't fit",
			xp: &XProperties{
				Min:       0,
				Max:       1000,
				ReqYWidth: 2,
				LO:        LabelOrientationVertical,
			},
			cvsAr: image.Rect(0, 0, 10, 10),
			want: &XDetails{
				Start: image.Point{2, 5},
				End:   image.Point{9, 5},
				Scale: mustNewXScale(0, 1000, 7, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewValue(0, nonZeroDecimals),
						Pos:   image.Point{3, 6},
					},
					{
						Value: NewValue(615, nonZeroDecimals),
						Pos:   image.Point{7, 6},
					},
				},
				Properties: &XProperties{
					Min:       0,
					Max:       1000,
					ReqYWidth: 2,
					LO:        LabelOrientationVertical,
				},
			},
		},
		{
			desc: "accounts for longer vertical labels, the tallest label fits",
			xp: &XProperties{
				Min:       0,
				Max:       999,
				ReqYWidth: 2,
				LO:        LabelOrientationVertical,
			},
			cvsAr: image.Rect(0, 0, 10, 10),
			want: &XDetails{
				Start: image.Point{2, 6},
				End:   image.Point{9, 6},
				Scale: mustNewXScale(0, 999, 7, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewValue(0, nonZeroDecimals),
						Pos:   image.Point{3, 7},
					},
					{
						Value: NewValue(615, nonZeroDecimals),
						Pos:   image.Point{7, 7},
					},
				},
				Properties: &XProperties{
					Min:       0,
					Max:       999,
					ReqYWidth: 2,
					LO:        LabelOrientationVertical,
				},
			},
		},
		{
			desc: "accounts for longer custom labels, vertical",
			xp: &XProperties{
				Min:       0,
				Max:       1,
				ReqYWidth: 5,
				CustomLabels: map[int]string{
					0: "start",
					1: "end",
				},
				LO: LabelOrientationVertical,
			},
			cvsAr: image.Rect(0, 0, 20, 10),
			want: &XDetails{
				Start: image.Point{5, 4},
				End:   image.Point{19, 4},
				Scale: mustNewXScale(0, 1, 14, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewTextValue("start"),
						Pos:   image.Point{6, 5},
					},
					{
						Value: NewTextValue("end"),
						Pos:   image.Point{19, 5},
					},
				},
				Properties: &XProperties{
					Min:       0,
					Max:       1,
					ReqYWidth: 5,
					CustomLabels: map[int]string{
						0: "start",
						1: "end",
					},
					LO: LabelOrientationVertical,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := NewXDetails(tc.cvsAr, tc.xp)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewXDetails => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			t.Log(fmt.Sprintf("got: %v", got))

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewXDetails => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestRequiredHeight(t *testing.T) {
	tests := []struct {
		desc             string
		max              int
		customLabels     map[int]string
		labelOrientation LabelOrientation
		want             int
	}{
		{
			desc: "horizontal orientation",
			want: 2,
		},
		{
			desc:             "vertical orientation, no custom labels, need single row for max label",
			max:              8,
			labelOrientation: LabelOrientationVertical,
			want:             2,
		},
		{
			desc:             "vertical orientation, no custom labels, need multiple rows for max label",
			max:              100,
			labelOrientation: LabelOrientationVertical,
			want:             4,
		},
		{
			desc:             "vertical orientation, custom labels but all shorter than max label",
			max:              100,
			customLabels:     map[int]string{1: "a", 2: "b"},
			labelOrientation: LabelOrientationVertical,
			want:             4,
		},
		{
			desc:             "vertical orientation, custom labels and some longer than max label",
			max:              99,
			customLabels:     map[int]string{1: "a", 2: "bbbbb"},
			labelOrientation: LabelOrientationVertical,
			want:             6,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := RequiredHeight(tc.max, tc.customLabels, tc.labelOrientation)
			if got != tc.want {
				t.Errorf("RequiredHeight => %d, want %d", got, tc.want)
			}
		})
	}
}
