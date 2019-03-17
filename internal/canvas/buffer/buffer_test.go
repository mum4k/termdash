// Copyright 2018 Google Inc.
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

package buffer

import (
	"fmt"
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
)

func TestNewCells(t *testing.T) {
	tests := []struct {
		desc string
		text string
		opts []cell.Option
		want []*Cell
	}{
		{
			desc: "no cells for empty text",
		},
		{
			desc: "cells created from text with default options",
			text: "hello",
			want: []*Cell{
				NewCell('h'),
				NewCell('e'),
				NewCell('l'),
				NewCell('l'),
				NewCell('o'),
			},
		},
		{
			desc: "cells with options",
			text: "ha",
			opts: []cell.Option{
				cell.FgColor(cell.ColorCyan),
				cell.BgColor(cell.ColorMagenta),
			},
			want: []*Cell{
				NewCell('h', cell.FgColor(cell.ColorCyan), cell.BgColor(cell.ColorMagenta)),
				NewCell('a', cell.FgColor(cell.ColorCyan), cell.BgColor(cell.ColorMagenta)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := NewCells(tc.text, tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewCells => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNewCell(t *testing.T) {
	tests := []struct {
		desc string
		r    rune
		opts []cell.Option
		want *Cell
	}{
		{
			desc: "creates empty cell with default options",
			want: &Cell{
				Opts: &cell.Options{},
			},
		},
		{
			desc: "cell with the specified rune",
			r:    'X',
			want: &Cell{
				Rune: 'X',
				Opts: &cell.Options{},
			},
		},
		{
			desc: "cell with options",
			r:    'X',
			opts: []cell.Option{
				cell.FgColor(cell.ColorCyan),
				cell.BgColor(cell.ColorMagenta),
			},
			want: &Cell{
				Rune: 'X',
				Opts: &cell.Options{
					FgColor: cell.ColorCyan,
					BgColor: cell.ColorMagenta,
				},
			},
		},
		{
			desc: "passing full cell.Options overwrites existing",
			r:    'X',
			opts: []cell.Option{
				&cell.Options{
					FgColor: cell.ColorBlack,
					BgColor: cell.ColorBlue,
				},
			},
			want: &Cell{
				Rune: 'X',
				Opts: &cell.Options{
					FgColor: cell.ColorBlack,
					BgColor: cell.ColorBlue,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := NewCell(tc.r, tc.opts...)
			t.Logf(fmt.Sprintf("%v", got))
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewCell => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestCellCopy(t *testing.T) {
	tests := []struct {
		desc string
		cell *Cell
		want *Cell
	}{
		{
			desc: "copies empty cell",
			cell: NewCell(0),
			want: NewCell(0),
		},
		{
			desc: "copies cell with a rune",
			cell: NewCell(33),
			want: NewCell(33),
		},
		{
			desc: "copies cell with rune and options",
			cell: NewCell(42, cell.FgColor(cell.ColorCyan), cell.BgColor(cell.ColorBlack)),
			want: NewCell(
				42,
				cell.FgColor(cell.ColorCyan),
				cell.BgColor(cell.ColorBlack),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.cell.Copy()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Copy => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestCellApply(t *testing.T) {
	tests := []struct {
		desc string
		cell *Cell
		opts []cell.Option
		want *Cell
	}{
		{
			desc: "no options provided",
			cell: NewCell(0),
			want: NewCell(0),
		},
		{
			desc: "no change in options",
			cell: NewCell(0, cell.FgColor(cell.ColorCyan)),
			opts: []cell.Option{
				cell.FgColor(cell.ColorCyan),
			},
			want: NewCell(0, cell.FgColor(cell.ColorCyan)),
		},
		{
			desc: "retains previous values",
			cell: NewCell(0, cell.FgColor(cell.ColorCyan)),
			opts: []cell.Option{
				cell.BgColor(cell.ColorBlack),
			},
			want: NewCell(
				0,
				cell.FgColor(cell.ColorCyan),
				cell.BgColor(cell.ColorBlack),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.cell
			got.Apply(tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Apply => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc    string
		size    image.Point
		want    Buffer
		wantErr bool
	}{
		{
			desc:    "zero buffer is invalid",
			wantErr: true,
		},
		{
			desc:    "width cannot be negative",
			size:    image.Point{-1, 1},
			wantErr: true,
		},
		{
			desc:    "height cannot be negative",
			size:    image.Point{1, -1},
			wantErr: true,
		},
		{
			desc: "creates single cell buffer",
			size: image.Point{1, 1},
			want: Buffer{
				{
					NewCell(0),
				},
			},
		},
		{
			desc: "creates the buffer",
			size: image.Point{2, 3},
			want: Buffer{
				{
					NewCell(0),
					NewCell(0),
					NewCell(0),
				},
				{
					NewCell(0),
					NewCell(0),
					NewCell(0),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := New(tc.size)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("New => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestBufferSize(t *testing.T) {
	sizes := []image.Point{
		{1, 1},
		{2, 3},
	}

	for _, size := range sizes {
		t.Run("", func(t *testing.T) {
			b, err := New(size)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			got := b.Size()
			if diff := pretty.Compare(size, got); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

// mustNew returns a new Buffer or panics.
func mustNew(size image.Point) Buffer {
	b, err := New(size)
	if err != nil {
		panic(err)
	}
	return b
}

func TestSetCell(t *testing.T) {
	size := image.Point{3, 3}
	tests := []struct {
		desc      string
		buffer    Buffer
		point     image.Point
		r         rune
		opts      []cell.Option
		wantCells int
		want      Buffer
		wantErr   bool
	}{
		{
			desc:    "point falls before the buffer",
			buffer:  mustNew(size),
			point:   image.Point{-1, -1},
			r:       'A',
			wantErr: true,
		},
		{
			desc:    "point falls after the buffer",
			buffer:  mustNew(size),
			point:   image.Point{3, 3},
			r:       'A',
			wantErr: true,
		},
		{
			desc: "point falls on cell with partial rune",
			buffer: func() Buffer {
				b := mustNew(size)
				b[0][0].Rune = '世'
				return b
			}(),
			point:   image.Point{1, 0},
			r:       'A',
			wantErr: true,
		},
		{
			desc: "point falls on cell with full-width rune and overwrites with half-width rune",
			buffer: func() Buffer {
				b := mustNew(size)
				b[0][0].Rune = '世'
				return b
			}(),
			point:     image.Point{0, 0},
			r:         'A',
			wantCells: 1,
			want: func() Buffer {
				b := mustNew(size)
				b[0][0].Rune = 'A'
				return b
			}(),
		},
		{
			desc: "point falls on cell with full-width rune and overwrites with full-width rune",
			buffer: func() Buffer {
				b := mustNew(size)
				b[0][0].Rune = '世'
				return b
			}(),
			point:     image.Point{0, 0},
			r:         '界',
			wantCells: 2,
			want: func() Buffer {
				b := mustNew(size)
				b[0][0].Rune = '界'
				return b
			}(),
		},
		{
			desc:    "not enough space for a wide rune on the line",
			buffer:  mustNew(image.Point{3, 3}),
			point:   image.Point{2, 0},
			r:       '界',
			wantErr: true,
		},
		{
			desc:      "sets half-width rune in a cell",
			buffer:    mustNew(image.Point{3, 3}),
			point:     image.Point{1, 1},
			r:         'A',
			wantCells: 1,
			want: func() Buffer {
				b := mustNew(size)
				b[1][1].Rune = 'A'
				return b
			}(),
		},
		{
			desc:      "sets full-width rune in a cell",
			buffer:    mustNew(image.Point{3, 3}),
			point:     image.Point{1, 2},
			r:         '界',
			wantCells: 2,
			want: func() Buffer {
				b := mustNew(size)
				b[1][2].Rune = '界'
				return b
			}(),
		},
		{
			desc:   "sets cell options",
			buffer: mustNew(image.Point{3, 3}),
			point:  image.Point{1, 2},
			r:      'A',
			opts: []cell.Option{
				cell.FgColor(cell.ColorRed),
				cell.BgColor(cell.ColorBlue),
			},
			wantCells: 1,
			want: func() Buffer {
				b := mustNew(size)
				c := b[1][2]
				c.Rune = 'A'
				c.Opts = cell.NewOptions(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
				return b
			}(),
		},
		{
			desc: "overwrites only provided options",
			buffer: func() Buffer {
				b := mustNew(size)
				c := b[1][2]
				c.Opts = cell.NewOptions(cell.BgColor(cell.ColorBlue))
				return b
			}(),
			point: image.Point{1, 2},
			r:     'A',
			opts: []cell.Option{
				cell.FgColor(cell.ColorRed),
			},
			wantCells: 1,
			want: func() Buffer {
				b := mustNew(size)
				c := b[1][2]
				c.Rune = 'A'
				c.Opts = cell.NewOptions(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
				return b
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotCells, err := tc.buffer.SetCell(tc.point, tc.r, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("SetCell => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if gotCells != tc.wantCells {
				t.Errorf("SetCell => unexpected cell count, got %d, want %d", gotCells, tc.wantCells)
			}

			got := tc.buffer
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("SetCell=> unexpected buffer, diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestIsPartial(t *testing.T) {
	tests := []struct {
		desc    string
		buffer  Buffer
		point   image.Point
		want    bool
		wantErr bool
	}{
		{
			desc:    "point falls before the buffer",
			buffer:  mustNew(image.Point{1, 1}),
			point:   image.Point{-1, -1},
			wantErr: true,
		},
		{
			desc:    "point falls after the buffer",
			buffer:  mustNew(image.Point{1, 1}),
			point:   image.Point{1, 1},
			wantErr: true,
		},
		{
			desc:   "the first cell cannot be partial",
			buffer: mustNew(image.Point{1, 1}),
			point:  image.Point{0, 0},
			want:   false,
		},
		{
			desc:   "previous cell on the same line contains no rune",
			buffer: mustNew(image.Point{3, 3}),
			point:  image.Point{1, 0},
			want:   false,
		},
		{
			desc: "previous cell on the same line contains half-width rune",
			buffer: func() Buffer {
				b := mustNew(image.Point{3, 3})
				b[0][0].Rune = 'A'
				return b
			}(),
			point: image.Point{1, 0},
			want:  false,
		},
		{
			desc: "previous cell on the same line contains full-width rune",
			buffer: func() Buffer {
				b := mustNew(image.Point{3, 3})
				b[0][0].Rune = '世'
				return b
			}(),
			point: image.Point{1, 0},
			want:  true,
		},
		{
			desc:   "previous cell on previous line contains no rune",
			buffer: mustNew(image.Point{3, 3}),
			point:  image.Point{0, 1},
			want:   false,
		},
		{
			desc: "previous cell on previous line contains half-width rune",
			buffer: func() Buffer {
				b := mustNew(image.Point{3, 3})
				b[2][0].Rune = 'A'
				return b
			}(),
			point: image.Point{0, 1},
			want:  false,
		},
		{
			desc: "previous cell on previous line contains full-width rune",
			buffer: func() Buffer {
				b := mustNew(image.Point{3, 3})
				b[2][0].Rune = '世'
				return b
			}(),
			point: image.Point{0, 1},
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.buffer.IsPartial(tc.point)
			if (err != nil) != tc.wantErr {
				t.Errorf("IsPartial => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if got != tc.want {
				t.Errorf("IsPartial => got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRemWidth(t *testing.T) {
	tests := []struct {
		desc    string
		size    image.Point
		point   image.Point
		want    int
		wantErr bool
	}{
		{
			desc:    "point falls before the buffer",
			size:    image.Point{1, 1},
			point:   image.Point{-1, -1},
			wantErr: true,
		},
		{
			desc:    "point falls after the buffer",
			size:    image.Point{1, 1},
			point:   image.Point{1, 1},
			wantErr: true,
		},
		{
			desc:  "remaining width from the first cell on the line",
			size:  image.Point{3, 3},
			point: image.Point{0, 1},
			want:  3,
		},
		{
			desc:  "remaining width from the last cell on the line",
			size:  image.Point{3, 3},
			point: image.Point{2, 2},
			want:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			b, err := New(tc.size)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}
			got, err := b.RemWidth(tc.point)
			if (err != nil) != tc.wantErr {
				t.Errorf("RemWidth => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if got != tc.want {
				t.Errorf("RemWidth => got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCellsWidth(t *testing.T) {
	tests := []struct {
		desc  string
		cells []*Cell
		want  int
	}{
		{
			desc:  "ascii characters",
			cells: NewCells("hello"),
			want:  5,
		},
		{
			desc:  "string from mattn/runewidth/runewidth_test",
			cells: NewCells("■㈱の世界①"),
			want:  10,
		},
		{
			desc:  "string using termdash characters",
			cells: NewCells("⇄…⇧⇩"),
			want:  4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := CellsWidth(tc.cells); got != tc.want {
				t.Errorf("CellsWidth => %v, want %v", got, tc.want)
			}
		})
	}
}
