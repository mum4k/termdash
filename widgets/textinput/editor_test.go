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

func TestRangeWidth(t *testing.T) {
	tests := []struct {
		desc     string
		data     fieldData
		startIdx int
		endIdx   int
		want     int
	}{
		{
			desc:     "empty range",
			startIdx: 0,
			endIdx:   0,
			want:     0,
		},
		{
			desc:     "single half-width rune",
			data:     fieldData{'a', 'b', '世', 'd'},
			startIdx: 1,
			endIdx:   2,
			want:     1,
		},
		{
			desc:     "single full-width rune",
			data:     fieldData{'a', 'b', '世', 'd'},
			startIdx: 2,
			endIdx:   3,
			want:     2,
		},
		{
			desc:     "mix of multiple runes",
			data:     fieldData{'a', 'b', '世', 'd'},
			startIdx: 1,
			endIdx:   4,
			want:     4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.data.rangeWidth(tc.startIdx, tc.endIdx)
			if got != tc.want {
				t.Errorf("rangeWidth => %d, want %d", got, tc.want)
			}
		})
	}
}

func TestFieldEditor(t *testing.T) {
	t.Skip()
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
			width:   2,
			wantErr: true,
		},
		{
			desc:       "no data",
			width:      3,
			want:       "",
			wantCurIdx: 0,
		},
		{
			desc:  "data and cursor fit exactly",
			width: 3,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				return nil
			},
			want:       "ab",
			wantCurIdx: 2,
		},
		{
			desc:  "longer data than the width, cursor at the end",
			width: 3,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				return nil
			},
			want:       "⇦c",
			wantCurIdx: 2,
		},
		{
			desc:  "width decreases, adjusts cursor and shifts data",
			width: 3,
			ops: func(fe *fieldEditor) error {
				if _, _, err := fe.viewFor(4); err != nil {
					return err
				}
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				return nil
			},
			want:       "⇦c",
			wantCurIdx: 2,
		},
		{
			desc:  "cursor won't go right beyond the end of the field",
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
				fe.cursorLeft()
				return nil
			},
			want:       "⇦cd",
			wantCurIdx: 2,
		},
		{
			desc:  "scrolls content to the left, both ends invisible",
			width: 4,
			ops: func(fe *fieldEditor) error {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
				fe.insert('d')
				fe.cursorLeft()
				fe.cursorLeft()
				return nil
			},
			want:       "⇦b⇨",
			wantCurIdx: 1,
		},

		// Less text than width.
		// Tests with full-width runes.
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
