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
		wantErr  bool
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
			desc:    "plenty of width, prefers equal split",
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
			cvsWidth: 50,
			want:     []columnWidth{25, 25},
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
			desc:    "two columns, canvas not wide enough, user specified widths",
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
			opts: []ContentOption{
				ColumnWidthsPercent(99, 1),
			},
			cvsWidth: 4,
			want:     []columnWidth{3, 1},
		},
		{
			desc:    "fails when canvas width not enough for user specified widths",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
					NewCell("cde"),
				),
			},
			opts: []ContentOption{
				ColumnWidthsPercent(99, 1),
			},
			cvsWidth: 1,
			wantErr:  true,
		},
		{
			desc:    "fails when canvas width not enough for optimized widths",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
					NewCell("cde"),
				),
			},
			cvsWidth: 1,
			wantErr:  true,
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
			desc:    "cells that wrap aren't accounted for, wrap configured at cell level",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("ab"),
					NewCellWithOpts(
						[]*Data{
							NewData("cde"),
						},
						CellWrapAtWords(),
					),
				),
				NewRow(
					NewCell("a"),
					NewCellWithOpts(
						[]*Data{
							NewData("cde"),
						},
						CellWrapAtWords(),
					),
				),
				NewRow(
					NewCell(""),
					NewCellWithOpts(
						[]*Data{
							NewData("cde"),
						},
						CellWrapAtWords(),
					),
				),
			},
			cvsWidth: 4,
			want:     []columnWidth{2, 2},
		},
		{
			desc:    "cells that wrap aren't accounted for, wrap configured at content level",
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
			opts: []ContentOption{
				WrapAtWords(),
			},
			cvsWidth: 4,
			want:     []columnWidth{2, 2},
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

			got, err := columnWidths(content, tc.cvsWidth)
			if (err != nil) != tc.wantErr {
				t.Errorf("columnWidths => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("columnWidths => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkColumnWidths(b *testing.B) {
	for n := 0; n < b.N; n++ {
		content, err := NewContent(Columns(10), nil)
		if err != nil {
			b.Fatalf("NewContent => unexpected error: %v", err)
		}
		content.AddRows([]*Row{
			NewRow(NewCell("00")),
			NewRow(NewCell("11")),
			NewRow(NewCell("2222")),
			NewRow(NewCell("33")),
			NewRow(NewCell("44")),
			NewRow(NewCell("555")),
			NewRow(NewCell("6666")),
			NewRow(NewCell("7")),
			NewRow(NewCell("88888")),
			NewRow(NewCell("999999")),
		})
		content.AddRows([]*Row{
			NewRow(NewCell("00000")),
			NewRow(NewCell("11")),
			NewRow(NewCell("2222")),
			NewRow(NewCell("33")),
			NewRow(NewCell("444")),
			NewRow(NewCell("555")),
			NewRow(NewCell("66")),
			NewRow(NewCell("7")),
			NewRow(NewCell("888")),
			NewRow(NewCell("999")),
		})
		content.AddRows([]*Row{
			NewRow(NewCell("000000")),
			NewRow(NewCell("11")),
			NewRow(NewCell("2222")),
			NewRow(NewCell("3333333")),
			NewRow(NewCell("44")),
			NewRow(NewCell("555555555555555")),
			NewRow(NewCell("6")),
			NewRow(NewCell("7")),
			NewRow(NewCell("8")),
			NewRow(NewCell("999999")),
		})
		columnWidths(content, 30)
	}
}

func TestSplitToPercent(t *testing.T) {
	tests := []struct {
		desc          string
		cvsWidth      int
		widthsPercent []int
		want          []columnWidth
		wantErr       bool
	}{
		{
			desc:    "fails for zero canvas width",
			wantErr: true,
		},
		{
			desc:     "fails when no widths provided",
			cvsWidth: 10,
			wantErr:  true,
		},
		{
			desc:          "fails when we don't have at least one cell per column",
			cvsWidth:      2,
			widthsPercent: []int{10, 50, 20},
			wantErr:       true,
		},
		{
			desc:          "single column",
			cvsWidth:      15,
			widthsPercent: []int{100},
			want:          []columnWidth{15},
		},
		{
			desc:          "divides evenly into the percentages",
			cvsWidth:      10,
			widthsPercent: []int{10, 50, 20, 10, 10},
			want:          []columnWidth{1, 5, 2, 1, 1},
		},
		{
			desc:          "divides unevenly into the percentages, last is the largest",
			cvsWidth:      3,
			widthsPercent: []int{10, 90},
			want:          []columnWidth{1, 2},
		},
		{
			desc:          "divides unevenly into the percentages, first is the largest",
			cvsWidth:      3,
			widthsPercent: []int{90, 10},
			want:          []columnWidth{2, 1},
		},
		{
			desc:          "each column is given at least one cell",
			cvsWidth:      3,
			widthsPercent: []int{10, 80, 10},
			want:          []columnWidth{1, 1, 1},
		},
		{
			desc:          "leaves at least one cell for each remaining column",
			cvsWidth:      6,
			widthsPercent: []int{99, 1, 1, 1, 1},
			want:          []columnWidth{2, 1, 1, 1, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := splitToPercent(tc.cvsWidth, tc.widthsPercent)
			if (err != nil) != tc.wantErr {
				t.Errorf("splitToPercent => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("splitToPercent => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSplitEqually(t *testing.T) {
	tests := []struct {
		desc     string
		cvsWidth int
		columns  int
		want     []columnWidth
	}{
		{
			desc:     "single column",
			cvsWidth: 9,
			columns:  1,
			want:     []columnWidth{9},
		},
		{
			desc:     "splits evenly",
			cvsWidth: 9,
			columns:  3,
			want:     []columnWidth{3, 3, 3},
		},
		{
			desc:     "splits unevenly",
			cvsWidth: 10,
			columns:  3,
			want:     []columnWidth{3, 3, 4},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := splitEqually(tc.cvsWidth, tc.columns)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("splitEqually => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}
