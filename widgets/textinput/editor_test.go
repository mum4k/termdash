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
		desc        string
		width       int
		ops         func(*fieldEditor) error
		wantView    string
		wantContent string
		wantCurIdx  int
		wantErr     bool
	}{
		{
			desc:    "fails for width too small",
			width:   3,
			wantErr: true,
		},
		{
			desc:        "no data",
			width:       4,
			wantView:    "",
			wantContent: "",
			wantCurIdx:  0,
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
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  3,
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
			wantView:    "⇦cd",
			wantContent: "abcd",
			wantCurIdx:  3,
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
			wantView:    "⇦世",
			wantContent: "abc世",
			wantCurIdx:  3,
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
			wantView:    "⇦cd",
			wantContent: "abcd",
			wantCurIdx:  3,
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
			wantView:    "⇦cd",
			wantContent: "abcd",
			wantCurIdx:  3,
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
			wantView:    "⇦cd",
			wantContent: "abcd",
			wantCurIdx:  2,
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
			wantView:    "abc⇨",
			wantContent: "abcd",
			wantCurIdx:  1,
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
			wantView:    "⇦cd⇨",
			wantContent: "abcde",
			wantCurIdx:  1,
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
			wantView:    "⇦de",
			wantContent: "abcde",
			wantCurIdx:  3,
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
			wantView:    "abc⇨",
			wantContent: "abcde",
			wantCurIdx:  0,
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
			wantView:    "⇦de",
			wantContent: "abcde",
			wantCurIdx:  3,
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
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  3,
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
			wantView:    "abc⇨",
			wantContent: "abcde",
			wantCurIdx:  0,
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
			wantView:    "⇦de",
			wantContent: "abcde",
			wantCurIdx:  3,
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
			wantView:    "⇦cd",
			wantContent: "abcd",
			wantCurIdx:  3,
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
			wantView:    "⇦世",
			wantContent: "abc世",
			wantCurIdx:  3,
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
			wantView:    "acd⇨",
			wantContent: "acde",
			wantCurIdx:  1,
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
			wantView:    "世c⇨",
			wantContent: "世cde",
			wantCurIdx:  2,
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
			wantView:    "abc⇨",
			wantContent: "abcde",
			wantCurIdx:  0,
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
			wantView:    "⇦de",
			wantContent: "abcde",
			wantCurIdx:  3,
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
			wantView:    "acd⇨",
			wantContent: "acde",
			wantCurIdx:  1,
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
			wantView:    "ade",
			wantContent: "ade",
			wantCurIdx:  1,
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
			wantView:    "⇦世",
			wantContent: "ab世",
			wantCurIdx:  1,
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
			wantView:    "ac",
			wantContent: "ac",
			wantCurIdx:  1,
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
			wantView:    "a世",
			wantContent: "a世",
			wantCurIdx:  1,
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
			wantView:    "",
			wantContent: "",
			wantCurIdx:  0,
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
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  3,
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
			wantView:    "你好世",
			wantContent: "你好世",
			wantCurIdx:  6,
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
			wantView:    "⇦cd⇨",
			wantContent: "abcde",
			wantCurIdx:  1,
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
			wantView:    "abc⇨",
			wantContent: "abcde",
			wantCurIdx:  1,
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
			wantView:    "acd⇨",
			wantContent: "acde",
			wantCurIdx:  1,
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
			wantView:    "⇦df",
			wantContent: "abcdf",
			wantCurIdx:  2,
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
			wantView:    "⇦de",
			wantContent: "abcde",
			wantCurIdx:  2,
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
			wantView:    "⇦de",
			wantContent: "abde",
			wantCurIdx:  1,
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
			wantView:    "⇦⇦世⇨",
			wantContent: "你好世界",
			wantCurIdx:  2,
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
			wantView:    "你好⇨",
			wantContent: "你好世界",
			wantCurIdx:  2,
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
			wantView:    "你世⇨",
			wantContent: "你世界",
			wantCurIdx:  2,
		},
		{
			desc:  "full-width runes only, both ends invisible, scrolls to make end visible",
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
				fe.cursorRight()
				return nil
			},
			wantView:    "⇦⇦界",
			wantContent: "你好世界",
			wantCurIdx:  2,
		},
		{
			desc:  "full-width runes only, both ends invisible, deletes to make end visible",
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
				fe.delete()
				return nil
			},
			wantView:    "⇦⇦界",
			wantContent: "你好界",
			wantCurIdx:  2,
		},
		{
			desc:  "scrolls to make full-width rune appear at the beginning",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
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
			wantView:    "你b⇨",
			wantContent: "你bcd",
			wantCurIdx:  2,
		},
		{
			desc:  "scrolls to make full-width rune appear at the end",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('你')
				fe.cursorStart()
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			wantView:    "⇦你",
			wantContent: "abc你",
			wantCurIdx:  1,
		},
		{
			desc:  "inserts after last full width rune, first is half-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('你')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.insert('e')
				return nil
			},
			wantView:    "⇦c你e",
			wantContent: "abc你e",
			wantCurIdx:  5,
		},
		{
			desc:  "inserts after last full width rune, first is half-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('世')
				fe.insert('b')
				fe.insert('你')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.insert('d')
				return nil
			},
			wantView:    "⇦你d",
			wantContent: "世b你d",
			wantCurIdx:  4,
		},
		{
			desc:  "inserts after last full width rune, hidden rune is full-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('世')
				fe.insert('你')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.insert('c')
				fe.insert('d')
				return nil
			},
			wantView:    "⇦⇦cd",
			wantContent: "世你cd",
			wantCurIdx:  4,
		},
		{
			desc:  "scrolls right, first is full-width, last are half-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('你')
				fe.insert('世')
				fe.insert('d')
				fe.insert('e')
				fe.insert('f')
				fe.insert('g')
				fe.insert('h')
				fe.cursorStart()
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			wantView:    "⇦⇦def⇨",
			wantContent: "a你世defgh",
			wantCurIdx:  3,
		},
		{
			desc:  "scrolls right, first is half-width, last is full-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('你')
				fe.insert('世')
				fe.insert('f')
				fe.insert('g')
				fe.insert('h')
				fe.cursorStart()
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			wantView:    "⇦你世⇨",
			wantContent: "abc你世fgh",
			wantCurIdx:  3,
		},
		{
			desc:  "scrolls right, first and last are full-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				fe.insert('界')
				fe.cursorStart()
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			wantView:    "⇦⇦世⇨",
			wantContent: "你好世界",
			wantCurIdx:  2,
		},
		{
			desc:  "scrolls right, first and last are half-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				fe.insert('f')
				fe.insert('g')
				fe.cursorStart()
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				fe.cursorRight()
				return nil
			},
			wantView:    "⇦cdef⇨",
			wantContent: "abcdefg",
			wantCurIdx:  4,
		},
		{
			desc:  "scrolls left, first is full-width, last are half-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('你')
				fe.insert('世')
				fe.insert('d')
				fe.insert('e')
				fe.insert('f')
				fe.insert('g')
				fe.insert('h')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			wantView:    "⇦⇦def⇨",
			wantContent: "a你世defgh",
			wantCurIdx:  2,
		},
		{
			desc:  "scrolls left, first is half-width, last is full-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('你')
				fe.insert('世')
				fe.insert('f')
				fe.insert('g')
				fe.insert('h')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			wantView:    "⇦你世⇨",
			wantContent: "abc你世fgh",
			wantCurIdx:  1,
		},
		{
			desc:  "scrolls left, first and last are full-width",
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
			wantView:    "⇦⇦世⇨",
			wantContent: "你好世界",
			wantCurIdx:  2,
		},
		{
			desc:  "scrolls left, first and last are half-width",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.insert('e')
				fe.insert('f')
				fe.insert('g')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			wantView:    "⇦cdef⇨",
			wantContent: "abcdefg",
			wantCurIdx:  1,
		},
		{
			desc:  "resets the field editor",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.reset()
				return nil
			},
			wantView:    "",
			wantContent: "",
			wantCurIdx:  0,
		},
		{
			desc:  "doesn't insert runes with rune width of zero",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('\x08')
				fe.insert('c')
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				return nil
			},
			wantView:    "ac",
			wantContent: "ac",
			wantCurIdx:  2,
		},
		{
			desc:  "all text visible, moves cursor to position zero",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorRelCell(0)
				return nil
			},
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  0,
		},
		{
			desc:  "all text visible, moves cursor to position in the middle",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorRelCell(1)
				return nil
			},
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  1,
		},
		{
			desc:  "all text visible, moves cursor back to the last character",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorStart()
				fe.cursorRelCell(2)
				return nil
			},
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  2,
		},
		{
			desc:  "all text visible, moves cursor to the appending space",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorStart()
				fe.cursorRelCell(3)
				return nil
			},
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  3,
		},
		{
			desc:  "all text visible, moves cursor before the beginning of data",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorStart()
				fe.cursorRelCell(-1)
				return nil
			},
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  0,
		},
		{
			desc:  "all text visible, moves cursor after the appending space",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				if _, _, err := fe.viewFor(6); err != nil {
					return err
				}
				fe.cursorStart()
				fe.cursorRelCell(10)
				return nil
			},
			wantView:    "abc",
			wantContent: "abc",
			wantCurIdx:  3,
		},
		{
			desc:  "moves cursor when there is no text",
			width: 6,
			ops: func(fe *fieldEditor) error {
				fe.cursorRelCell(10)
				return nil
			},
			wantView:    "",
			wantContent: "",
			wantCurIdx:  0,
		},
		{
			desc:  "both ends hidden, moves cursor onto the left arrow",
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
				fe.cursorRelCell(0)
				return nil
			},
			wantView:    "⇦cd⇨",
			wantContent: "abcde",
			wantCurIdx:  1,
		},
		{
			desc:  "both ends hidden, moves cursor onto the first character",
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
				fe.cursorRelCell(1)
				return nil
			},
			wantView:    "⇦cd⇨",
			wantContent: "abcde",
			wantCurIdx:  1,
		},
		{
			desc:  "both ends hidden, moves cursor onto the right arrow",
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
				fe.cursorRelCell(3)
				return nil
			},
			wantView:    "⇦cd⇨",
			wantContent: "abcde",
			wantCurIdx:  2,
		},
		{
			desc:  "both ends hidden, moves cursor onto the last character",
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
				fe.cursorRelCell(2)
				return nil
			},
			wantView:    "⇦cd⇨",
			wantContent: "abcde",
			wantCurIdx:  2,
		},
		{
			desc:  "moves cursor onto the first cell containing a full-width rune",
			width: 8,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				fe.insert('界')
				fe.insert('你')
				if _, _, err := fe.viewFor(8); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(8); err != nil {
					return err
				}
				fe.cursorRelCell(4)
				return nil
			},
			wantView:    "⇦⇦世界⇨",
			wantContent: "你好世界你",
			wantCurIdx:  4,
		},
		{
			desc:  "moves cursor onto the second cell containing a full-width rune",
			width: 8,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				fe.insert('界')
				fe.insert('你')
				if _, _, err := fe.viewFor(8); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(8); err != nil {
					return err
				}
				fe.cursorRelCell(5)
				return nil
			},
			wantView:    "⇦⇦世界⇨",
			wantContent: "你好世界你",
			wantCurIdx:  4,
		},
		{
			desc:  "moves cursor onto the second right arrow",
			width: 8,
			ops: func(fe *fieldEditor) error {
				fe.insert('你')
				fe.insert('好')
				fe.insert('世')
				fe.insert('界')
				fe.insert('你')
				if _, _, err := fe.viewFor(8); err != nil {
					return err
				}
				fe.cursorLeft()
				fe.cursorLeft()
				fe.cursorLeft()
				if _, _, err := fe.viewFor(8); err != nil {
					return err
				}
				fe.cursorRelCell(1)
				return nil
			},
			wantView:    "⇦⇦世界⇨",
			wantContent: "你好世界你",
			wantCurIdx:  2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fe := newFieldEditor()
			if tc.ops != nil {
				if err := tc.ops(fe); err != nil {
					t.Fatalf("ops => unexpected error: %v", err)
				}
			}

			gotView, gotCurIdx, err := fe.viewFor(tc.width)
			if (err != nil) != tc.wantErr {
				t.Errorf("viewFor(%d) => unexpected error: %v, wantErr: %v", tc.width, err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if gotView != tc.wantView || gotCurIdx != tc.wantCurIdx {
				t.Errorf("viewFor(%d) => (%q, %d), want (%q, %d)", tc.width, gotView, gotCurIdx, tc.wantView, tc.wantCurIdx)
			}

			gotContent := fe.content()
			if gotContent != tc.wantContent {
				t.Errorf("content -> %q, want %q", gotContent, tc.wantContent)
			}
		})
	}
}
