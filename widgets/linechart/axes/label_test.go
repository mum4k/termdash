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
		desc       string
		min        float64
		max        float64
		cvsHeight  int
		labelWidth int
		want       []*Label
		wantErr    bool
	}{
		{
			desc:       "fails when canvas is too small",
			min:        0,
			max:        1,
			cvsHeight:  1,
			labelWidth: 4,
			wantErr:    true,
		},
		{
			desc:       "fails when labelWidth is too small",
			min:        0,
			max:        1,
			cvsHeight:  2,
			labelWidth: 0,
			wantErr:    true,
		},
		{
			desc:       "only two rows on the canvas, labels min and max",
			min:        0,
			max:        5,
			cvsHeight:  2,
			labelWidth: 1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 1}},
				{NewValue(2.88, nonZeroDecimals), image.Point{0, 0}},
			},
		},
		{
			desc:       "aligns labels to the right",
			min:        0,
			max:        5,
			cvsHeight:  2,
			labelWidth: 5,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{4, 1}},
				{NewValue(2.88, nonZeroDecimals), image.Point{1, 0}},
			},
		},
		{
			desc:       "multiple labels, last on the top",
			min:        0,
			max:        5,
			cvsHeight:  9,
			labelWidth: 1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 8}},
				{NewValue(2.4, nonZeroDecimals), image.Point{0, 4}},
				{NewValue(4.8, nonZeroDecimals), image.Point{0, 0}},
			},
		},
		{
			desc:       "multiple labels, last on top-1",
			min:        0,
			max:        5,
			cvsHeight:  10,
			labelWidth: 1,
			want: []*Label{
				{NewValue(0, nonZeroDecimals), image.Point{0, 9}},
				{NewValue(2.08, nonZeroDecimals), image.Point{0, 5}},
				{NewValue(4.16, nonZeroDecimals), image.Point{0, 1}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			scale, err := NewYScale(tc.min, tc.max, tc.cvsHeight, nonZeroDecimals)
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
