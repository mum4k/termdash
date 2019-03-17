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

package table

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestColumnWidths(t *testing.T) {
	tests := []struct {
		desc     string
		columns  Columns
		rows     []*Row
		opts     []ContentOption
		cvsWidth int
		want     []columnWidth
	}{
		{
			desc:    "single column, wide enough",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
				),
				NewRow(
					NewCell(""),
				),
			},
			cvsWidth: 2,
			want:     []columnWidth{2},
		},
		{
			desc:    "single column, not wide enough",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
				),
				NewRow(
					NewCell(""),
				),
			},
			cvsWidth: 1,
			want:     []columnWidth{1},
		},
		{
			desc:    "two columns, canvas wide enough, no trimming",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
					NewCell("cde"),
				),
				NewRow(
					NewCell("a"),
					NewCell("cde"),
				),
				NewRow(
					NewCell(""),
					NewCell("cde"),
				),
			},
			cvsWidth: 5,
			want:     []columnWidth{2, 3},
		},
		{
			desc:    "two columns, canvas not wide enough, optimal to trim first",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
					NewCell("cde"),
				),
				NewRow(
					NewCell("a"),
					NewCell("cde"),
				),
				NewRow(
					NewCell(""),
					NewCell("cde"),
				),
			},
			cvsWidth: 4,
			want:     []columnWidth{1, 3},
		},
		{
			desc:    "two columns, canvas not wide enough, optimal to trim second",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
					NewCell("c"),
				),
				NewRow(
					NewCell("ab"),
					NewCell("c"),
				),
				NewRow(
					NewCell(""),
					NewCell("cde"),
				),
			},
			cvsWidth: 4,
			want:     []columnWidth{2, 2},
		},
		{
			desc:    "three columns, canvas wide enough, no trimming",
			columns: Columns(3),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
					NewCell("cde"),
					NewCell("fg"),
				),
				NewRow(
					NewCell("a"),
					NewCell("c"),
					NewCell("f"),
				),
				NewRow(
					NewCell("ab"),
					NewCell("cde"),
					NewCell("fg"),
				),
			},
			cvsWidth: 7,
			want:     []columnWidth{2, 3, 2},
		},
		{
			desc:    "three columns, canvas not wide enough, one very long cell",
			columns: Columns(3),
			rows: []*Row{
				NewRow(
					NewCell("00"),
					NewCell("11111111"),
					NewCell("22"),
				),
				NewRow(
					NewCell("00"),
					NewCell("11"),
					NewCell("22"),
				),
				NewRow(
					NewCell("00"),
					NewCell("111"),
					NewCell("22"),
				),
			},
			cvsWidth: 6,
			want:     []columnWidth{2, 2, 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			content, err := NewContent(tc.columns, tc.rows, tc.opts...)
			if err != nil {
				t.Fatalf("NewContent => unexpected error: %v", err)
			}

			got := columnWidths(content, tc.cvsWidth)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("columnWidths => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
