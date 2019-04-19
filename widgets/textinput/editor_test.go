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

package textinput

import (
	"fmt"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestData(t *testing.T) {
	tests := []struct {
		desc string
		data fieldData
		ops  func(*fieldData)
		want fieldData
	}{
		{
			desc: "appends to empty data",
			ops: func(fd *fieldData) {
				fd.insertAt(0, 'a')
			},
			want: fieldData{'a'},
		},
		{
			desc: "appends at the end of non-empty data",
			data: fieldData{'a'},
			ops: func(fd *fieldData) {
				fd.insertAt(1, 'b')
				fd.insertAt(2, 'c')
			},
			want: fieldData{'a', 'b', 'c'},
		},
		{
			desc: "appends at the beginning of non-empty data",
			data: fieldData{'a'},
			ops: func(fd *fieldData) {
				fd.insertAt(0, 'b')
				fd.insertAt(0, 'c')
			},
			want: fieldData{'c', 'b', 'a'},
		},
		{
			desc: "deletes the last rune, result in empty",
			data: fieldData{'a'},
			ops: func(fd *fieldData) {
				fd.deleteAt(0)
			},
			want: fieldData{},
		},
		{
			desc: "deletes the last rune, result in non-empty",
			data: fieldData{'a', 'b'},
			ops: func(fd *fieldData) {
				fd.deleteAt(1)
			},
			want: fieldData{'a'},
		},
		{
			desc: "deletes runes in the middle",
			data: fieldData{'a', 'b', 'c', 'd'},
			ops: func(fd *fieldData) {
				fd.deleteAt(1)
				fd.deleteAt(1)
			},
			want: fieldData{'a', 'd'},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.data
			if tc.ops != nil {
				tc.ops(&got)
			}
			t.Logf(fmt.Sprintf("got: %s", got))
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("fieldData => unexpected diff (-want, +got):\n%s\n got: %q\nwant: %q", diff, got, tc.want)
			}
		})
	}
}

func TestCellsBefore(t *testing.T) {
	tests := []struct {
		desc   string
		data   fieldData
		cells  int
		endIdx int
		want   int
	}{
		{
			desc:   "empty data and range",
			cells:  1,
			endIdx: 0,
			want:   0,
		},
		{
			desc:   "requesting zero cells",
			data:   fieldData{'a', 'b', '世', 'd'},
			cells:  0,
			endIdx: 1,
			want:   1,
		},
		{
			desc:   "data only has one rune",
			data:   fieldData{'a'},
			cells:  1,
			endIdx: 1,
			want:   0,
		},
		{
			desc:   "non-empty data and empty range",
			data:   fieldData{'a', 'b', '世', 'd'},
			cells:  1,
			endIdx: 0,
			want:   0,
		},
		{
			desc:   "more cells than runes from endIdx",
			data:   fieldData{'a', 'b', '世', 'd'},
			cells:  10,
			endIdx: 1,
			want:   0,
		},
		{
			desc:   "less cells than runes from endIdx, stops on half-width rune",
			data:   fieldData{'a', 'b', '世', 'd'},
			cells:  1,
			endIdx: 2,
			want:   1,
		},
		{
			desc:   "less cells than runes from endIdx, stops on full-width rune",
			data:   fieldData{'a', 'b', '世', 'd'},
			cells:  2,
			endIdx: 3,
			want:   2,
		},
		{
			desc:   "less cells than runes from endIdx, full-width rune doesn't fit",
			data:   fieldData{'a', 'b', '世', 'd'},
			cells:  2,
			endIdx: 4,
			want:   3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.data.cellsBefore(tc.cells, tc.endIdx)
			if got != tc.want {
				t.Errorf("cellsBefore => %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCellsAfter(t *testing.T) {
	tests := []struct {
		desc     string
		data     fieldData
		cells    int
		startIdx int
		want     int
	}{
		{
			desc:     "empty data and range",
			cells:    1,
			startIdx: 0,
			want:     0,
		},
		{
			desc:     "empty data and range, non-zero start",
			cells:    1,
			startIdx: 1,
			want:     1,
		},
		{
			desc:     "data only has one rune",
			data:     fieldData{'a'},
			cells:    1,
			startIdx: 0,
			want:     1,
		},
		{
			desc:     "non-empty data and empty range",
			data:     fieldData{'a', 'b', '世', 'd'},
			cells:    0,
			startIdx: 1,
			want:     1,
		},
		{
			desc:     "more cells than runes from startIdx",
			data:     fieldData{'a', 'b', '世', 'd'},
			cells:    10,
			startIdx: 1,
			want:     4,
		},
		{
			desc:     "less cells than runes from startIdx, stops on half-width rune",
			data:     fieldData{'a', 'b', '世', 'd', 'e', 'f'},
			cells:    2,
			startIdx: 3,
			want:     5,
		},
		{
			desc:     "less cells than runes from startIdx, stops on full-width rune",
			data:     fieldData{'a', 'b', '世', 'd'},
			cells:    3,
			startIdx: 1,
			want:     3,
		},
		{
			desc:     "less cells than runes from startIdx, full-width rune doesn't fit",
			data:     fieldData{'a', 'b', '世', 'd'},
			cells:    3,
			startIdx: 0,
			want:     2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.data.cellsAfter(tc.cells, tc.startIdx)
			if got != tc.want {
				t.Errorf("cellsAfter => %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRunesIn(t *testing.T) {
	tests := []struct {
		desc string
		data fieldData
		vr   *visibleRange
		want string
	}{
		{
			desc: "zero range, zero data",
			vr:   &visibleRange{},
			want: "",
		},
		{
			desc: "zero range, non-zero data",
			data: fieldData{'a', 'b', '世', 'd'},
			vr:   &visibleRange{},
			want: "",
		},
		{
			desc: "range from zero, start and end visible",
			data: fieldData{'a', 'b', '世', 'd'},
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   5,
			},
			want: "ab世d",
		},
		{
			desc: "range from zero, start visible end hidden",
			data: fieldData{'a', 'b', '世', 'd'},
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   4,
			},
			want: "ab世⇨",
		},
		{
			desc: "range from zero, end not visible",
			data: fieldData{'a', 'b', '世', 'd'},
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "ab⇨",
		},
		{
			desc: "range from non-zero, end not visible",
			data: fieldData{'a', 'b', '世', 'd'},
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   4,
			},
			want: "⇦世⇨",
		},
		{
			desc: "range from non-zero, start not visible, end visible",
			data: fieldData{'a', 'b', '世', 'd'},
			vr: &visibleRange{
				startIdx: 2,
				endIdx:   5,
			},
			want: "⇦d",
		},
		{
			desc: "range from non-zero, neither start nor end visible",
			data: fieldData{'a', 'b', '世', 'd', 'e'},
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   4,
			},
			want: "⇦世⇨",
		},
		{
			desc: "range from non-zero, neither start nor end visible, range too short for arrows",
			data: fieldData{'a', 'b', '世', 'd', 'e'},
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   3,
			},
			want: "b",
		},
		{
			desc: "range longer than data",
			data: fieldData{'a', 'b', '世', 'd'},
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   5,
			},
			want: "ab世d",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.data.runesIn(tc.vr)
			if got != tc.want {
				t.Errorf("runesIn => %q, want %q", got, tc.want)
			}
		})
	}
}

func TestForRunes(t *testing.T) {
	tests := []struct {
		desc string
		vr   *visibleRange
		want int
	}{
		{
			desc: "empty range",
			vr:   &visibleRange{},
			want: 0,
		},
		{
			desc: "reserves one for the cursor at the end",
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   3,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.vr.forRunes()
			if got != tc.want {
				t.Fatalf("forRunes => %d, want %d", got, tc.want)
			}

		})
	}
}

func TestCurMinIdx(t *testing.T) {
	tests := []struct {
		desc string
		vr   *visibleRange
		want int
	}{
		{
			desc: "zero values",
			vr:   &visibleRange{},
			want: 0,
		},
		{
			desc: "first rune visible",
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   5,
			},
			want: 0,
		},
		{
			desc: "first rune hidden, wide enough for arrows",
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   6,
			},
			want: 2,
		},
		{
			desc: "first rune hidden, no space for arrows",
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   2,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.vr.curMinIdx()
			if got != tc.want {
				t.Errorf("curMinIdx => %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCurMaxIdx(t *testing.T) {
	tests := []struct {
		desc      string
		vr        *visibleRange
		runeCount int
		want      int
	}{
		{
			desc:      "zero values",
			vr:        &visibleRange{},
			runeCount: 0,
			want:      0,
		},
		{
			desc: "last rune visible and space for appending",
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   4,
			},
			runeCount: 3,
			want:      3,
		},
		{
			desc: "last rune visible, space for appending not visible",
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			runeCount: 3,
			want:      2,
		},
		{
			desc: "last rune hidden, enough runes for arrows",
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   4,
			},
			runeCount: 5,
			want:      2,
		},
		{
			desc: "last rune hidden, not enough runes for arrows",
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   2,
			},
			runeCount: 3,
			want:      1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.vr.curMaxIdx(tc.runeCount)
			if got != tc.want {
				t.Errorf("curMaxIdx => %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNormalizeToWidth(t *testing.T) {
	tests := []struct {
		desc  string
		vr    *visibleRange
		width int
		want  *visibleRange
	}{
		{
			desc:  "zero values",
			vr:    &visibleRange{},
			width: 0,
			want:  &visibleRange{},
		},
		{
			desc: "width decreased to zero",
			vr: &visibleRange{
				startIdx: 10,
				endIdx:   15,
			},
			width: 0,
			want: &visibleRange{
				startIdx: 15,
				endIdx:   15,
			},
		},
		{
			desc: "width increased from zero",
			vr: &visibleRange{
				startIdx: 15,
				endIdx:   15,
			},
			width: 5,
			want: &visibleRange{
				startIdx: 10,
				endIdx:   15,
			},
		},
		{
			desc: "width increased by more than the width of the data",
			vr: &visibleRange{
				startIdx: 10,
				endIdx:   15,
			},
			width: 20,
			want: &visibleRange{
				startIdx: 0,
				endIdx:   20,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.vr
			got.normalizeToWidth(tc.width)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("normalizeToWidth => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNormalizeToData(t *testing.T) {
	tests := []struct {
		desc string
		vr   *visibleRange
		data fieldData
		want string
	}{
		{
			desc: "zero values",
			vr:   &visibleRange{},
			data: fieldData{},
			want: "",
		},
		{
			desc: "data smaller than visible range, range already at the start",
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			data: fieldData{'a'},
			want: "a",
		},
		{
			desc: "data smaller than visible range by exactly one rune - space for the cursor",
			vr: &visibleRange{
				startIdx: 4,
				endIdx:   6,
			},
			data: fieldData{'a', 'b', 'c', 'd', 'e'},
			want: "e",
		},
		{
			desc: "data smaller than visible range, range is shifted back, not reaching zero",
			vr: &visibleRange{
				startIdx: 4,
				endIdx:   7,
			},
			data: fieldData{'a', 'b', 'c', 'd'},
			want: "⇦d",
		},
		{
			desc: "range decreases due to full-width rune",
			vr: &visibleRange{
				startIdx: 4,
				endIdx:   7,
			},
			data: fieldData{'a', 'b', 'c', '世'},
			want: "世",
		},
		{
			desc: "dataLen smaller than visible range, range is shifted back, reaches zero",
			vr: &visibleRange{
				startIdx: 4,
				endIdx:   6,
			},
			data: fieldData{'a'},
			want: "a",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tc.vr.normalizeToData(tc.data)
			got := tc.data.runesIn(tc.vr)
			if got != tc.want {
				t.Errorf("normalizeToData => %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCurRelative(t *testing.T) {
	tests := []struct {
		desc       string
		vr         *visibleRange
		curDataPos int
		want       int
		wantErr    bool
	}{
		{
			desc: "fails when cursor isn't in the range",
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   5,
			},
			curDataPos: 5,
			wantErr:    true,
		},
		{
			desc: "cursor falls at the beginning of the range",
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   6,
			},
			curDataPos: 3,
			want:       0,
		},
		{
			desc: "cursor falls at the end of the range",
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   6,
			},
			curDataPos: 5,
			want:       2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.vr.curRelative(tc.curDataPos)
			if (err != nil) != tc.wantErr {
				t.Errorf("curRelative => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if got != tc.want {
				t.Errorf("curRelative => %d, want %d", got, tc.want)
			}
		})
	}
}

func TestToCursor(t *testing.T) {
	tests := []struct {
		desc       string
		data       fieldData
		curDataPos int
		vr         *visibleRange
		want       string
	}{
		{
			desc:       "no-op without data",
			data:       fieldData{},
			curDataPos: 0,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   0,
			},
			want: "",
		},
		{
			desc:       "no-op when cursor is in",
			data:       fieldData{'a', 'b', 'c'},
			curDataPos: 1,
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   3,
			},
			want: "b",
		},
		{
			desc:       "shifts left, first rune visible",
			data:       fieldData{'a', 'b', 'c', 'd', 'e'},
			curDataPos: 0,
			vr: &visibleRange{
				startIdx: 1,
				endIdx:   4,
			},
			want: "ab⇨",
		},
		{
			desc:       "shifts left, first rune hidden",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 2,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   6,
			},
			want: "⇦c⇨",
		},
		{
			desc:       "shifts left, first rune hidden, cursor on the left arrow",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 3,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   6,
			},
			want: "⇦d⇨",
		},
		{
			desc:       "shifts left, first rune hidden, multiple visible runes",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 2,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   7,
			},
			want: "⇦cd⇨",
		},
		{
			desc:       "shifts left, too narrow for arrows",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 1,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   5,
			},
			want: "b",
		},
		{
			desc:       "shifts left, range longer than data",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 3,
			vr: &visibleRange{
				startIdx: 4,
				endIdx:   10,
			},
			want: "⇦def",
		},
		{
			desc:       "shifts left, starts on full-width rune, loses space for arrows",
			data:       fieldData{'a', 'b', '世', 'd', 'e', 'f'},
			curDataPos: 2,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   6,
			},
			want: "世",
		},
		{
			desc:       "shifts left, starts on full-width rune, last rune fits exactly",
			data:       fieldData{'a', 'b', '世', 'd', 'e', 'f', 'g'},
			curDataPos: 2,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   7,
			},
			want: "⇦世⇨",
		},
		{
			desc:       "shifts left, starts on full-width rune, last rune doesn't fit",
			data:       fieldData{'a', 'b', '世', '世', 'e', 'f', 'g'},
			curDataPos: 2,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   6,
			},
			want: "世",
		},
		{
			desc:       "shifts left, starts on full-width rune, last rune doesn't fit but arrows do",
			data:       fieldData{'a', 'b', '世', 'd', '世', 'f', 'g', 'h'},
			curDataPos: 2,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   8,
			},
			want: "⇦世d⇨",
		},
		{
			desc:       "shifts left, starts on full-width rune, last rune doesn't fit but arrows do",
			data:       fieldData{'a', '世', 'c', 'd', 'e', 'f', 'g', 'h'},
			curDataPos: 2,
			vr: &visibleRange{
				startIdx: 3,
				endIdx:   7,
			},
			want: "⇦c⇨",
		},
		{
			desc:       "shifts right, last rune visible, cursor on the last rune",
			data:       fieldData{'a', 'b', 'c', 'd', 'e'},
			curDataPos: 4,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "⇦e",
		},
		{
			desc:       "shifts right, last rune visible, cursor after the last rune",
			data:       fieldData{'a', 'b', 'c', 'd', 'e'},
			curDataPos: 5,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "⇦e",
		},
		{
			desc:       "shifts right, last rune hidden",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 4,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "⇦e⇨",
		},
		{
			desc:       "shifts right, last rune hidden, cursor on the arrow",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 3,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "⇦d⇨",
		},
		{
			desc:       "shifts right, too narrow for arrows",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 4,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   2,
			},
			want: "e",
		},
		{
			desc:       "shifts right, too narrow for arrows, cursor on the last rune",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 5,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   2,
			},
			want: "f",
		},
		{
			desc:       "shifts right, too narrow for arrows, cursor after the last rune",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 6,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   2,
			},
			want: "f",
		},
		{
			desc:       "shifts right, cursor on the penultimate rune",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			curDataPos: 4,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "⇦e⇨",
		},
		{
			desc:       "shifts right, ends on full-width rune, loses space for arrows",
			data:       fieldData{'a', 'b', 'c', 'd', '世', 'f'},
			curDataPos: 4,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "世",
		},
		{
			desc:       "shifts right, ends on full-width rune, first rune fits exactly",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', '世', 'g'},
			curDataPos: 5,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   5,
			},
			want: "⇦e世⇨",
		},
		{
			desc:       "shifts right, ends on full-width rune, first rune doesn't fit",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', '世', 'g'},
			curDataPos: 5,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   4,
			},
			want: "⇦世⇨",
		},
		{
			desc:       "shifts right, ends on full-width rune, first rune doesn't fit, no arrows",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', '世', 'g'},
			curDataPos: 5,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "世",
		},
		{
			desc:       "shifts right, arrow at the end hides full-width rune",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', '世', 'g'},
			curDataPos: 4,
			vr: &visibleRange{
				startIdx: 0,
				endIdx:   3,
			},
			want: "⇦e⇨",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fe := newFieldEditor()
			fe.data = tc.data
			fe.curDataPos = tc.curDataPos
			fe.visible = tc.vr

			fe.toCursor()
			got := fe.data.runesIn(fe.visible)
			t.Logf("got %#v", *fe.visible)
			if got != tc.want {
				t.Errorf("toCursor => %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFieldEditor(t *testing.T) {
	tests := []struct {
		desc       string
		width      int
		ops        func(*fieldEditor) error
		want       string
		wantCurIdx int
		wantErr    bool
	}{
		{
			desc:    "fails for width too small",
			width:   3,
			wantErr: true,
		},
		{
			desc:       "no data",
			width:      4,
			want:       "",
			wantCurIdx: 0,
		},
		{
			desc:  "data and cursor fit exactly",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				return nil
			},
			want:       "abc",
			wantCurIdx: 3,
		},
		{
			desc:  "longer data than the width, cursor at the end",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				return nil
			},
			want:       "⇦cd",
			wantCurIdx: 3,
		},
		{
			desc:  "longer data than the width, cursor at the end, has full-width runes",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('世')
				return nil
			},
			want:       "⇦世",
			wantCurIdx: 2,
		},
		{
			desc:  "width decreased, adjusts cursor and shifts data",
			width: 4,
			ops: func(fe *fieldEditor) error {
				if _, _, err := fe.viewFor(5); err != nil {
					return err
				}
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				return nil
			},
			want:       "⇦cd",
			wantCurIdx: 3,
		},
		{
			desc:  "cursor won't go right beyond the end of the data",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			want:       "⇦cd",
			wantCurIdx: 3,
		},
		{
			desc:  "moves cursor to the left",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				return nil
			},
			want:       "⇦cd",
			wantCurIdx: 2,
		},
		{
			desc:  "scrolls content to the left, start becomes visible",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			want:       "abc⇨",
			wantCurIdx: 1,
		},
		{
			desc:  "scrolls content to the left, both ends invisible",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			want:       "⇦cd⇨",
			wantCurIdx: 1,
		},
		{
			desc:  "scrolls left, then back right to make end visible again",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			want:       "⇦de",
			wantCurIdx: 3,
		},
		{
			desc:  "scrolls left, won't go beyond the start of data",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			want:       "abc⇨",
			wantCurIdx: 0,
		},
		{
			desc:  "scrolls left, then back right won't go beyond the end of data",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			want:       "⇦de",
			wantCurIdx: 3,
		},
		{
			desc:  "have less data than width, all fits",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				return nil
			},
			want:       "abc",
			wantCurIdx: 3,
		},
		{
			desc:  "moves cursor to the start",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorStart()
				return nil
			},
			want:       "abc⇨",
			wantCurIdx: 0,
		},
		{
			desc:  "moves cursor to the end",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorStart()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorEnd()
				return nil
			},
			want:       "⇦de",
			wantCurIdx: 3,
		},
		{
			desc:  "deletesBefore when cursor after the data",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.deleteBefore()
				return nil
			},
			want:       "⇦cd",
			wantCurIdx: 3,
		},
		{
			desc:  "deletesBefore when cursor after the data, text has full-width rune",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('世')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.deleteBefore()
				return nil
			},
			want:       "⇦世",
			wantCurIdx: 2,
		},
		{
			desc:  "deletesBefore when cursor in the middle",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.deleteBefore()
				return nil
			},
			want:       "acd⇨",
			wantCurIdx: 1,
		},
		{
			desc:  "deletesBefore when cursor in the middle, full-width runes",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('世')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.deleteBefore()
				return nil
			},
			want:       "世c⇨",
			wantCurIdx: 1,
		},
		{
			desc:  "deletesBefore does nothing when cursor at the start",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorStart()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.deleteBefore()
				return nil
			},
			want:       "abc⇨",
			wantCurIdx: 0,
		},
		// deletes the last rune, contains full-width runes
		// delete when at the empty space at the end
		// delete when in the middle, last rune visible
		// delete when in the middle, last rune hidden
		// delete at the beginning
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fe := newFieldEditor()
			if tc.ops != nil {
				if err := tc.ops(fe); err != nil {
					t.Fatalf("ops => unexpected error: %v", err)
				}
			}

			got, gotCurIdx, err := fe.viewFor(tc.width)
			if (err != nil) != tc.wantErr {
				t.Errorf("viewFor => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if got != tc.want || gotCurIdx != tc.wantCurIdx {
				t.Errorf("viewFor => (%q, %d), want (%q, %d)", got, gotCurIdx, tc.want, tc.wantCurIdx)
			}
		})
	}
}
