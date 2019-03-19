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

	"github.com/mum4k/termdash/terminal/terminalapi"
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
		{
			desc:    "fails on a cell that has zero colSpan",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCellWithOpts(
						[]*Data{NewData("0")},
						CellColSpan(0),
					),
				),
			},
			wantSubstr: "invalid CellColSpan",
		},
		{
			desc:    "fails on a cell that has zero rowSpan",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCellWithOpts(
						[]*Data{NewData("0")},
						CellRowSpan(0),
					),
				),
			},
			wantSubstr: "invalid CellRowSpan",
		},
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
		{
			desc:    "fails when the number of column widths doesn't match number of columns",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
			opts: []ContentOption{
				ColumnWidthsPercent(20, 20, 60),
			},
			wantSubstr: "invalid number of widths in ColumnWidthsPercent",
		},
		{
			desc:    "fails when the sum of column widths doesn't equal 100",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
			opts: []ContentOption{
				ColumnWidthsPercent(20, 20),
			},
			wantSubstr: "invalid sum of widths in ColumnWidthsPercent",
		},
		{
			desc:    "fails when zero height set on cell",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCellWithOpts([]*Data{NewData("1")}, CellHeight(0)),
				),
			},
			wantSubstr: "invalid height",
		},
		{
			desc:    "fails when zero height set on row",
			columns: Columns(2),
			rows: []*Row{
				NewRowWithOpts(
					[]*Cell{
						NewCell("0"),
						NewCell("1"),
					},
					RowHeight(0),
				),
			},
			wantSubstr: "invalid height",
		},
		{
			desc:    "fails when zero height set on content",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
			opts: []ContentOption{
				ContentRowHeight(0),
			},
			wantSubstr: "invalid height",
		},
		{
			desc:    "fails when zero horizontal padding set on cell",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCellWithOpts([]*Data{NewData("1")}, CellHorizontalPadding(0)),
				),
			},
			wantSubstr: "invalid horizontal padding",
		},
		{
			desc:    "fails when zero horizontal padding set on row",
			columns: Columns(2),
			rows: []*Row{
				NewRowWithOpts(
					[]*Cell{
						NewCell("0"),
						NewCell("1"),
					},
					RowHorizontalPadding(0),
				),
			},
			wantSubstr: "invalid horizontal padding",
		},
		{
			desc:    "fails when zero horizontal padding set on content",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
			opts: []ContentOption{
				HorizontalPadding(0),
			},
			wantSubstr: "invalid horizontal padding",
		},
		{
			desc:    "fails when zero vertical padding set on cell",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCellWithOpts([]*Data{NewData("1")}, CellVerticalPadding(0)),
				),
			},
			wantSubstr: "invalid vertical padding",
		},
		{
			desc:    "fails when zero vertical padding set on row",
			columns: Columns(2),
			rows: []*Row{
				NewRowWithOpts(
					[]*Cell{
						NewCell("0"),
						NewCell("1"),
					},
					RowVerticalPadding(0),
				),
			},
			wantSubstr: "invalid vertical padding",
		},
		{
			desc:    "fails when zero vertical padding set on content",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
			opts: []ContentOption{
				VerticalPadding(0),
			},
			wantSubstr: "invalid vertical padding",
		},
		{
			desc:    "fails when zero vertical spacing",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
			opts: []ContentOption{
				VerticalSpacing(0),
			},
			wantSubstr: "invalid vertical spacing",
		},
		{
			desc:    "fails when zero horizontal spacing",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("0"),
					NewCell("1"),
				),
			},
			opts: []ContentOption{
				HorizontalSpacing(0),
			},
			wantSubstr: "invalid horizontal spacing",
		},
		{
			desc:    "fails when multiple header rows provided",
			columns: Columns(1),
			rows: []*Row{
				NewHeader(
					NewCell("0"),
				),
				NewHeader(
					NewCell("0"),
				),
			},
			wantSubstr: "one header",
		},
		{
			desc:    "fails on row callback for header row",
			columns: Columns(1),
			rows: []*Row{
				NewHeaderWithOpts(
					[]*Cell{
						NewCell("0"),
					},
					RowCallback(func(terminalapi.Event) error {
						return nil
					}),
				),
			},
			wantSubstr: "header row cannot have a callback",
		},
		{
			desc:    "fails when data contain invalid runes",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("\t"),
				),
			},
			wantSubstr: "invalid data",
		},
		{
			desc:    "succeeds when data contain empty string",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell(""),
				),
			},
		},
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

func TestAddRow(t *testing.T) {
	tests := []struct {
		desc    string
		columns Columns
		rows    []*Row
		add     []*Cell
		wantErr bool
	}{
		{
			desc:    "adds a row",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			add: []*Cell{
				NewCell("1"),
			},
		},
		{
			desc:    "fails when new row doesn't have the expected number of columns",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := NewContent(tc.columns, tc.rows)
			if err != nil {
				t.Fatalf("NewContent => unexpected error: %v", err)
			}
			{
				err := c.AddRow(tc.add...)
				if (err != nil) != tc.wantErr {
					t.Errorf("AddRow => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
			}
		})
	}
}
