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

func TestFieldEditor(t *testing.T) {
	tests := []struct {
		desc       string
		width      int
		ops        func(*fieldEditor)
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
			ops: func(fe *fieldEditor) {
				fe.insert('a')
				fe.insert('b')
			},
			want:       "ab",
			wantCurIdx: 2,
		},
		{
			desc:  "longer data than the width, cursor at the end",
			width: 3,
			ops: func(fe *fieldEditor) {
				fe.insert('a')
				fe.insert('b')
				fe.insert('c')
			},
			want:       "⇦c",
			wantCurIdx: 2,
		},
		// Tests with full-width runes.
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fe := newFieldEditor()
			if tc.ops != nil {
				tc.ops(fe)
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
