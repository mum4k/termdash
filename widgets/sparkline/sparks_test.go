// Copyright 2018 Google Inc.
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

package sparkline

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestVisibleMax(t *testing.T) {
	tests := []struct {
		desc     string
		data     []int
		width    int
		wantData []int
		wantMax  int
	}{
		{
			desc:     "zero for no data",
			width:    3,
			wantData: nil,
			wantMax:  0,
		},
		{
			desc:     "zero for zero width",
			data:     []int{0, 1},
			width:    0,
			wantData: nil,
			wantMax:  0,
		},
		{
			desc:     "zero for negative width",
			data:     []int{0, 1},
			width:    -1,
			wantData: nil,
			wantMax:  0,
		},
		{
			desc:     "all values are zero",
			data:     []int{0, 0, 0},
			width:    3,
			wantData: []int{0, 0, 0},
			wantMax:  0,
		},
		{
			desc:     "all values are visible",
			data:     []int{8, 0, 1},
			width:    3,
			wantData: []int{8, 0, 1},
			wantMax:  8,
		},
		{
			desc:     "width greater than number of values",
			data:     []int{8, 0, 1},
			width:    10,
			wantData: []int{8, 0, 1},
			wantMax:  8,
		},
		{
			desc:     "only some values are visible",
			data:     []int{8, 2, 1},
			width:    2,
			wantData: []int{2, 1},
			wantMax:  2,
		},
		{
			desc:     "only one value is visible",
			data:     []int{8, 2, 1},
			width:    1,
			wantData: []int{1},
			wantMax:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotData, gotMax := visibleMax(tc.data, tc.width)
			if diff := pretty.Compare(tc.wantData, gotData); diff != "" {
				t.Errorf("visibleMax => unexpected visible data, diff (-want, +got):\n%s", diff)
			}
			if gotMax != tc.wantMax {
				t.Errorf("visibleMax => gotMax %v, wantMax %v", gotMax, tc.wantMax)
			}
		})
	}
}

func TestToBlocks(t *testing.T) {
	tests := []struct {
		desc      string
		value     int
		max       int
		vertCells int
		want      blocks
	}{
		{
			desc:      "zero value has no blocks",
			value:     0,
			max:       10,
			vertCells: 2,
			want:      blocks{},
		},
		{
			desc:      "negative value has no blocks",
			value:     -1,
			max:       10,
			vertCells: 2,
			want:      blocks{},
		},
		{
			desc:      "zero max has no blocks",
			value:     10,
			max:       0,
			vertCells: 2,
			want:      blocks{},
		},
		{
			desc:      "negative max has no blocks",
			value:     10,
			max:       -1,
			vertCells: 2,
			want:      blocks{},
		},
		{
			desc:      "zero vertCells has no blocks",
			value:     10,
			max:       10,
			vertCells: 0,
			want:      blocks{},
		},
		{
			desc:      "negative vertCells has no blocks",
			value:     10,
			max:       10,
			vertCells: -1,
			want:      blocks{},
		},
		{
			desc:      "single line, zero value",
			value:     0,
			max:       8,
			vertCells: 1,
			want:      blocks{},
		},
		{
			desc:      "single line, value is 1/8",
			value:     1,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 0, partSpark: sparks[0]},
		},
		{
			desc:      "single line, value is 2/8",
			value:     2,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 0, partSpark: sparks[1]},
		},
		{
			desc:      "single line, value is 3/8",
			value:     3,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 0, partSpark: sparks[2]},
		},
		{
			desc:      "single line, value is 4/8",
			value:     4,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 0, partSpark: sparks[3]},
		},
		{
			desc:      "single line, value is 5/8",
			value:     5,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 0, partSpark: sparks[4]},
		},
		{
			desc:      "single line, value is 6/8",
			value:     6,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 0, partSpark: sparks[5]},
		},
		{
			desc:      "single line, value is 7/8",
			value:     7,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 0, partSpark: sparks[6]},
		},
		{
			desc:      "single line, value is 8/8",
			value:     8,
			max:       8,
			vertCells: 1,
			want:      blocks{full: 1, partSpark: 0},
		},
		{
			desc:      "multi line, zero value",
			value:     0,
			max:       24,
			vertCells: 3,
			want:      blocks{},
		},
		{
			desc:      "multi line, lowest block is partial",
			value:     2,
			max:       24,
			vertCells: 3,
			want:      blocks{full: 0, partSpark: sparks[1]},
		},
		{
			desc:      "multi line, two full blocks, no partial block",
			value:     16,
			max:       24,
			vertCells: 3,
			want:      blocks{full: 2, partSpark: 0},
		},
		{
			desc:      "multi line, topmost block is partial",
			value:     20,
			max:       24,
			vertCells: 3,
			want:      blocks{full: 2, partSpark: sparks[3]},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := toBlocks(tc.value, tc.max, tc.vertCells)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("toBlocks => unexpected diff (-want, +got):\n%s", diff)
				if got.full != tc.want.full {
					t.Errorf("toBlocks => unexpected diff, blocks.full got %d, want %d", got.full, tc.want.full)
				}
				if got.partSpark != tc.want.partSpark {
					t.Errorf("toBlocks => unexpected diff, blocks.partSpark got '%c' (sparks[%d])), want '%c' (sparks[%d])",
						got.partSpark, findRune(got.partSpark, sparks), tc.want.partSpark, findRune(tc.want.partSpark, sparks))
				}
			}
		})
	}
}

// findRune finds the rune in the slice and returns its index.
// Returns -1 if the rune isn't in the slice.
func findRune(target rune, runes []rune) int {
	for i, r := range runes {
		if r == target {
			return i
		}
	}
	return -1
}
