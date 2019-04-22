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
			desc:   "less cells than runes from endIdx, full-width rune doesn't fit, no space for arrows",
			data:   fieldData{'a', 'b', '世', 'd'},
			cells:  2,
			endIdx: 4,
			want:   3,
		},
		{
			desc:   "full-width runes only",
			data:   fieldData{'你', '好', '世', '界'},
			cells:  7,
			endIdx: 4,
			want:   1,
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
		{
			desc:     "full-width runes only",
			data:     fieldData{'你', '好', '世', '界'},
			cells:    7,
			startIdx: 0,
			want:     3,
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

func TestCurCell(t *testing.T) {
	tests := []struct {
		desc       string
		data       fieldData
		firstRune  int
		curDataPos int
		width      int
		want       int
		wantErr    bool
	}{
		{
			desc:       "empty data",
			data:       fieldData{},
			curDataPos: 0,
			want:       0,
		},
		{
			desc:       "cursor within the first page of data",
			data:       fieldData{'a', 'b', 'c', 'd'},
			firstRune:  1,
			curDataPos: 2,
			width:      3,
			want:       1,
		},
		{
			desc:       "cursor within the first page of data, after full-width rune",
			data:       fieldData{'a', '世', 'c', 'd'},
			firstRune:  1,
			curDataPos: 2,
			width:      3,
			want:       2,
		},
		{
			desc:       "cursor within the second page of data",
			data:       fieldData{'a', 'b', 'c', 'd', 'e', 'f'},
			firstRune:  3,
			curDataPos: 4,
			width:      3,
			want:       1,
		},
		{
			desc:       "cursor within the second page of data, after full-width rune",
			data:       fieldData{'a', 'b', 'c', '世', 'e', 'f'},
			firstRune:  3,
			curDataPos: 4,
			width:      3,
			want:       2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fe := newFieldEditor()
			fe.data = tc.data
			fe.firstRune = tc.firstRune
			fe.curDataPos = tc.curDataPos
			got := fe.curCell(tc.width)
			if got != tc.want {
				t.Errorf("curCell => %d, want %d", got, tc.want)
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
			wantCurIdx: 3,
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
			wantCurIdx: 3,
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
			wantCurIdx: 2,
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
		{
			desc:  "delete does nothing when cursor at the end",
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
				fe.delete()
				return nil
			},
			want:       "⇦de",
			wantCurIdx: 3,
		},
		{
			desc:  "delete in the middle, last rune remains hidden",
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
				fe.cursorRight()
				fe.delete()
				return nil
			},
			want:       "acd⇨",
			wantCurIdx: 1,
		},
		{
			desc:  "delete in the middle, last rune becomes visible",
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
				fe.cursorRight()
				fe.delete()
				fe.delete()
				return nil
			},
			want:       "ade",
			wantCurIdx: 1,
		},
		{
			desc:  "delete in the middle, last full-width rune would be invisible, shifts to keep cursor in window",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('世')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorStart()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				fe.delete()
				fe.delete()
				return nil
			},
			want:       "⇦世",
			wantCurIdx: 1,
		},
		{
			desc:  "delete in the middle, last rune was and is visible",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorStart()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorRight()
				fe.delete()
				return nil
			},
			want:       "ac",
			wantCurIdx: 1,
		},
		{
			desc:  "delete in the middle, last full-width rune was and is visible",
			width: 5,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('世')
				if _, _, err := fe.viewFor(5); err != nil {
					return err
				}
				fe.cursorStart()
				if _, _, err := fe.viewFor(5); err != nil {
					return err
				}
				fe.cursorRight()
				fe.delete()
				return nil
			},
			want:       "a世",
			wantCurIdx: 1,
		},
		{
			desc:  "delete last rune, contains full-width runes",
			width: 5,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('世')
				if _, _, err := fe.viewFor(5); err != nil {
					return err
				}
				fe.cursorStart()
				if _, _, err := fe.viewFor(5); err != nil {
					return err
				}
				fe.delete()
				fe.delete()
				fe.delete()
				return nil
			},
			want:       "",
			wantCurIdx: 0,
		},
		{
			desc:  "half-width runes only, exact fit",
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
			desc:  "full-width runes only, exact fit",
			width: 7,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				if _, _, err := fe.viewFor(7); err != nil {
					return err
				}
				return nil
			},
			want:       "你好世",
			wantCurIdx: 6,
		},
		{
			desc:  "half-width runes only, both ends hidden",
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
			desc:  "half-width runes only, both ends invisible, scrolls to make start visible",
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
				return nil
			},
			want:       "abc⇨",
			wantCurIdx: 1,
		},
		{
			desc:  "half-width runes only, both ends invisible, deletes to make start visible",
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
				fe.deleteBefore()
				return nil
			},
			want:       "acd⇨",
			wantCurIdx: 1,
		},
		{
			desc:  "half-width runes only, deletion on second page refills the field",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				fe.insert('f')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.delete()
				return nil
			},
			want:       "⇦df",
			wantCurIdx: 2,
		},
		{
			desc:  "half-width runes only, both ends invisible, scrolls to make end visible",
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
				return nil
			},
			want:       "⇦de",
			wantCurIdx: 2,
		},
		{
			desc:  "half-width runes only, both ends invisible, deletes to make end visible",
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
				fe.delete()
				return nil
			},
			want:       "⇦de",
			wantCurIdx: 1,
		},
		{
			desc:  "full-width runes only, both ends invisible",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				fe.insert('界')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			want:       "⇦⇦世⇨",
			wantCurIdx: 2,
		},
		{
			desc:  "full-width runes only, both ends invisible, scrolls to make start visible",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				fe.insert('界')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorLeft()
				return nil
			},
			want:       "你好⇨",
			wantCurIdx: 2,
		},
		{
			desc:  "full-width runes only, both ends invisible, deletes to make start visible",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				fe.insert('界')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.deleteBefore()
				return nil
			},
			want:       "你世⇨",
			wantCurIdx: 2,
		},

		// full-width runes only, both ends invisible, scrolls to make end visible
		// full-width runes only, both ends invisible, deletes to make end visible
		// scrolls to make full-width rune appear at the beginning
		// scrolls to make full-width rune appear at the end
		// inserts after last full width rune, first is half-width
		// inserts after last full width rune, first is full-width
		// scrolls right, first is full-width, last are half-width
		// scrolls right, first is half-width, last is full-width
		// scrolls right, first and last are full-width
		// scrolls right, first and last are half-width
		// scrolls left, first is full-width, last are half-width
		// scrolls left, first is half-width, last is full-width
		// scrolls left, first and last are full-width
		// scrolls left, first and last are half-width
		// test content
		// test reset
		// test insertion of invisible runes
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
