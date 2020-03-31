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
	"errors"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/wrap"
)

func TestHierarchical(t *testing.T) {
	tests := []struct {
		desc    string
		columns Columns
		rows    []*Row
		opts    []ContentOption

		// Got values retrieved from the first found cell.
		wantHPadding int
		wantVPadding int
		wantAlignH   align.Horizontal
		wantAlignV   align.Vertical
		wantHeight   int
		wantWrapMode wrap.Mode
	}{
		{
			desc:    "defaults when not set anywhere",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			wantHPadding: 0,
			wantVPadding: 0,
			wantAlignH:   align.HorizontalLeft,
			wantAlignV:   align.VerticalTop,
			wantHeight:   0,
			wantWrapMode: wrap.Never,
		},
		{
			desc:    "values set at content level",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			opts: []ContentOption{
				HorizontalPadding(1),
				VerticalPadding(2),
				AlignHorizontal(align.HorizontalCenter),
				AlignVertical(align.VerticalMiddle),
				ContentRowHeight(3),
				WrapAtWords(),
			},
			wantHPadding: 1,
			wantVPadding: 2,
			wantAlignH:   align.HorizontalCenter,
			wantAlignV:   align.VerticalMiddle,
			wantHeight:   3,
			wantWrapMode: wrap.AtWords,
		},
		{
			desc:    "values overridden at row level",
			columns: Columns(1),
			rows: []*Row{
				NewRowWithOpts(
					[]*Cell{
						NewCell("0"),
					},
					RowHorizontalPadding(10),
					RowVerticalPadding(20),
					RowAlignHorizontal(align.HorizontalRight),
					RowAlignVertical(align.VerticalBottom),
					RowHeight(30),
					RowWrapAtWords(),
				),
			},
			opts: []ContentOption{
				HorizontalPadding(1),
				VerticalPadding(2),
				AlignHorizontal(align.HorizontalCenter),
				AlignVertical(align.VerticalMiddle),
				ContentRowHeight(3),
			},
			wantHPadding: 10,
			wantVPadding: 20,
			wantAlignH:   align.HorizontalRight,
			wantAlignV:   align.VerticalBottom,
			wantHeight:   30,
			wantWrapMode: wrap.AtWords,
		},
		{
			desc:    "values overridden at cell level",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCellWithOpts(
						[]*Data{
							NewData("0"),
						},
						CellHorizontalPadding(10),
						CellVerticalPadding(20),
						CellAlignHorizontal(align.HorizontalRight),
						CellAlignVertical(align.VerticalBottom),
						CellHeight(30),
						CellWrapAtWords(),
					),
				),
			},
			opts: []ContentOption{
				HorizontalPadding(1),
				VerticalPadding(2),
				AlignHorizontal(align.HorizontalCenter),
				AlignVertical(align.VerticalMiddle),
				ContentRowHeight(3),
			},
			wantHPadding: 10,
			wantVPadding: 20,
			wantAlignH:   align.HorizontalRight,
			wantAlignV:   align.VerticalBottom,
			wantHeight:   30,
			wantWrapMode: wrap.AtWords,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := NewContent(tc.columns, tc.rows, tc.opts...)
			if err != nil {
				t.Fatalf("NewContent => unexpected error: %v", err)
			}

			fc, err := firstCell(c)
			if err != nil {
				t.Fatalf("firstCell => unexpected error: %v", err)
			}

			ho := fc.hierarchical
			if gotHPadding := ho.getHorizontalPadding(); gotHPadding != tc.wantHPadding {
				t.Errorf("getHorizontalPadding => %v, want %v", gotHPadding, tc.wantHPadding)
			}
			if gotVPadding := ho.getVerticalPadding(); gotVPadding != tc.wantVPadding {
				t.Errorf("getVerticalPadding => %v, want %v", gotVPadding, tc.wantVPadding)
			}
			if gotAlignH := ho.getAlignHorizontal(); gotAlignH != tc.wantAlignH {
				t.Errorf("getAlignHorizontal => %v, want %v", gotAlignH, tc.wantAlignH)
			}
			if gotAlignV := ho.getAlignVertical(); gotAlignV != tc.wantAlignV {
				t.Errorf("getAlignVertical => %v, want %v", gotAlignV, tc.wantAlignV)
			}
			if gotHeight := ho.getHeight(); gotHeight != tc.wantHeight {
				t.Errorf("getHeight => %v, want %v", gotHeight, tc.wantHeight)
			}
			if gotWrapMode := ho.getWrapMode(); gotWrapMode != tc.wantWrapMode {
				t.Errorf("getWrapMode => %v, want %v", gotWrapMode, tc.wantWrapMode)
			}
		})
	}
}

// firstCell returns the first cell found in the content or an error if there
// are none.
func firstCell(c *Content) (*Cell, error) {
	for _, tableRow := range c.rows {
		for _, tableCell := range tableRow.cells {
			return tableCell, nil
		}
	}
	return nil, errors.New("found no table cells in the content")
}

func TestDataCellOpts(t *testing.T) {
	tests := []struct {
		desc    string
		columns Columns
		rows    []*Row
		opts    []ContentOption
		// Expected options on all the data cells in order.
		want []*cell.Options
	}{
		{
			desc:    "default cell options when not set anywhere",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			want: []*cell.Options{
				cell.NewOptions(),
			},
		},
		{
			desc:    "inherited from content level",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCell("0"),
				),
			},
			opts: []ContentOption{
				ContentCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
			want: []*cell.Options{
				cell.NewOptions(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
		},
		{
			desc:    "inherits only to data with default options",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCellWithOpts(
						[]*Data{
							NewData(
								"0",
								DataCellOpts(
									cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack),
								),
							),
							NewData("1"),
						},
					),
				),
			},
			opts: []ContentOption{
				ContentCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
			want: []*cell.Options{
				cell.NewOptions(cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack)),
				cell.NewOptions(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
		},
		{
			desc:    "overridden at row level",
			columns: Columns(1),
			rows: []*Row{
				NewRowWithOpts(
					[]*Cell{
						NewCell("0"),
					},
					RowCellOpts(
						cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack),
					),
				),
			},
			opts: []ContentOption{
				ContentCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
			want: []*cell.Options{
				cell.NewOptions(cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack)),
			},
		},
		{
			desc:    "overridden at cell level",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCellWithOpts(
						[]*Data{
							NewData("0"),
						},
						CellOpts(
							cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack),
						),
					),
				),
			},
			opts: []ContentOption{
				ContentCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
			want: []*cell.Options{
				cell.NewOptions(cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack)),
			},
		},
		{
			desc:    "overridden at data level",
			columns: Columns(1),
			rows: []*Row{
				NewRow(
					NewCellWithOpts(
						[]*Data{
							NewData(
								"0",
								DataCellOpts(
									cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack),
								),
							),
						},
					),
				),
			},
			opts: []ContentOption{
				ContentCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
			want: []*cell.Options{
				cell.NewOptions(cell.FgColor(cell.ColorBlue), cell.BgColor(cell.ColorBlack)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := NewContent(tc.columns, tc.rows, tc.opts...)
			if err != nil {
				t.Fatalf("NewContent => unexpected error: %v", err)
			}

			var got []*cell.Options
			for _, row := range c.rows {
				for _, tableCell := range row.cells {
					for _, dataCell := range tableCell.data.cells {
						got = append(got, dataCell.Opts)
					}
				}
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewContent => unexpected data cell options, diff (-want, +got):\n%s", diff)
			}
		})
	}
}
