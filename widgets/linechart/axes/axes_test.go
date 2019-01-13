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

type updateY struct {
	minVal float64
	maxVal float64
}

func TestY(t *testing.T) {
	tests := []struct {
		desc      string
		minVal    float64
		maxVal    float64
		update    *updateY
		cvsAr     image.Rectangle
		wantWidth int
		want      *YDetails
		wantErr   bool
	}{
		{
			desc:      "fails on canvas too small",
			minVal:    0,
			maxVal:    3,
			cvsAr:     image.Rect(0, 0, 3, 2),
			wantWidth: 2,
			wantErr:   true,
		},
		{
			desc:      "fails on cvsWidth less than required width",
			minVal:    0,
			maxVal:    3,
			cvsAr:     image.Rect(0, 0, 2, 4),
			wantWidth: 2,
			wantErr:   true,
		},
		{
			desc:      "fails when max is less than min",
			minVal:    0,
			maxVal:    -1,
			cvsAr:     image.Rect(0, 0, 4, 4),
			wantWidth: 3,
			wantErr:   true,
		},
		{
			desc:      "cvsWidth equals required width",
			minVal:    0,
			maxVal:    3,
			cvsAr:     image.Rect(0, 0, 3, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 2,
				Start: image.Point{1, 0},
				End:   image.Point{1, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc:      "cvsWidth just accommodates the longest label",
			minVal:    0,
			maxVal:    3,
			cvsAr:     image.Rect(0, 0, 6, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 5,
				Start: image.Point{4, 0},
				End:   image.Point{4, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{3, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
		{
			desc:      "cvsWidth is more than we need",
			minVal:    0,
			maxVal:    3,
			cvsAr:     image.Rect(0, 0, 7, 4),
			wantWidth: 2,
			want: &YDetails{
				Width: 5,
				Start: image.Point{4, 0},
				End:   image.Point{4, 2},
				Scale: mustNewYScale(0, 3, 2, nonZeroDecimals),
				Labels: []*Label{
					{NewValue(0, nonZeroDecimals), image.Point{3, 1}},
					{NewValue(1.72, nonZeroDecimals), image.Point{0, 0}},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			y := NewY(tc.minVal, tc.maxVal)
			if tc.update != nil {
				y.Update(tc.update.minVal, tc.update.maxVal)
			}

			gotWidth := y.RequiredWidth()
			if gotWidth != tc.wantWidth {
				t.Errorf("RequiredWidth => got %v, want %v", gotWidth, tc.wantWidth)
			}

			got, err := y.Details(tc.cvsAr)
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
		desc      string
		numPoints int
		yStart    image.Point
		cvsWidth  int
		cvsAr     image.Rectangle
		want      *XDetails
		wantErr   bool
	}{
		{
			desc:      "fails when numPoints is negative",
			numPoints: -1,
			yStart:    image.Point{0, 0},
			cvsAr:     image.Rect(0, 0, 2, 3),
			wantErr:   true,
		},
		{
			desc:      "fails when cvsAr isn't wide enough",
			numPoints: 1,
			yStart:    image.Point{0, 0},
			cvsAr:     image.Rect(0, 0, 1, 3),
			wantErr:   true,
		},
		{
			desc:      "fails when cvsAr isn't tall enough",
			numPoints: 1,
			yStart:    image.Point{0, 0},
			cvsAr:     image.Rect(0, 0, 3, 2),
			wantErr:   true,
		},
		{
			desc:      "works with no data points",
			numPoints: 0,
			yStart:    image.Point{0, 0},
			cvsAr:     image.Rect(0, 0, 2, 3),
			want: &XDetails{
				Start: image.Point{0, 1},
				End:   image.Point{1, 1},
				Scale: mustNewXScale(0, 1, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewValue(0, nonZeroDecimals),
						Pos:   image.Point{1, 2},
					},
				},
			},
		},
		{
			desc:      "accounts for non-zero yStart",
			numPoints: 0,
			yStart:    image.Point{2, 0},
			cvsAr:     image.Rect(0, 0, 4, 5),
			want: &XDetails{
				Start: image.Point{2, 3},
				End:   image.Point{3, 3},
				Scale: mustNewXScale(0, 1, nonZeroDecimals),
				Labels: []*Label{
					{
						Value: NewValue(0, nonZeroDecimals),
						Pos:   image.Point{3, 4},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := NewXDetails(tc.numPoints, tc.yStart, tc.cvsAr)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewXDetails => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewXDetails => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
