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
	"strings"
	"testing"
)

func ExampleContent() {
	rows := []*Row{
		NewHeader(
			NewCell("hello"),
			NewCell("world"),
		),
		NewRow(
			NewCell("1"),
			NewCell("2"),
		),
	}

	_, err := NewContent(Columns(2), rows)
	if err != nil {
		panic(err)
	}
}

func TestContent(t *testing.T) {
	tests := []struct {
		desc       string
		columns    Columns
		rows       []*Row
		opts       []ContentOption
		wantSubstr string
	}{
		{
			desc:    "fails when number of columns is negative",
			columns: Columns(-1),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			wantSubstr: "invalid number of columns",
		},
		{
			desc:    "fails when rows doesn't have the specified number of columns",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			wantSubstr: "all rows must occupy",
		},
		{
			desc:    "succeeds when row columns match content",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
		},
		// fails on colspan <1
		// fails on rowspan <1
		{
			desc:    "succeeds when row has a column with a colSpan",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCellWithOpts(
						[]*Data{NewData("0")},
						CellColSpan(2),
					),
				),
			},
		},

		// Content:
		// rows with zero columns
		// rows with varying number of columns
		// rows with columns that have valid colspan
		// rows with columns that have valid rowspan
		// rows with columns that have valid colspan and rowspan
		// too many column widths
		// not enough column widths
		// zero column width
		// negative column width
		// zero row height
		// negative row height
		// zero and negative padding
		// zero and negative spacing

		// Row:
		// too many header rows
		// nil row callback
		// zero and negative row height
		// zero and negative padding

		// cell:
		// zero and negative colspan
		// zero and negative rowspan
		// zero and negative cell height
		// zero and negative padding

		// data:
		// invalid space characters in data

	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := NewContent(tc.columns, tc.rows, tc.opts...)
			if (err != nil) != (tc.wantSubstr != "") {
				t.Errorf("NewContent => unexpected error: %v, wantSubstr: %q", err, tc.wantSubstr)
			}
			if err != nil && !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("NewContent => unexpected error: %v, wantSubstr: %q", err, tc.wantSubstr)
			}
		})
	}
}
