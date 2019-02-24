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

package cell

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
		want *Options
	}{
		{
			desc: "no provided options",
			want: &Options{},
		},
		{
			desc: "setting foreground color",
			opts: []Option{
				FgColor(ColorBlack),
			},
			want: &Options{
				FgColor: ColorBlack,
			},
		},
		{
			desc: "setting background color",
			opts: []Option{
				BgColor(ColorRed),
			},
			want: &Options{
				BgColor: ColorRed,
			},
		},
		{
			desc: "setting multiple options",
			opts: []Option{
				FgColor(ColorCyan),
				BgColor(ColorMagenta),
			},
			want: &Options{
				FgColor: ColorCyan,
				BgColor: ColorMagenta,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := NewOptions(tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewOptions => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc string
		r    rune
		opts []Option
		want Cell
	}{
		{
			desc: "creates empty cell with default options",
			want: Cell{
				Opts: &Options{},
			},
		},
		{
			desc: "cell with the specified rune",
			r:    'X',
			want: Cell{
				Rune: 'X',
				Opts: &Options{},
			},
		},
		{
			desc: "cell with options",
			r:    'X',
			opts: []Option{
				FgColor(ColorCyan),
				BgColor(ColorMagenta),
			},
			want: Cell{
				Rune: 'X',
				Opts: &Options{
					FgColor: ColorCyan,
					BgColor: ColorMagenta,
				},
			},
		},
		{
			desc: "passing full Options overwrites existing",
			r:    'X',
			opts: []Option{
				&Options{
					FgColor: ColorBlack,
					BgColor: ColorBlue,
				},
			},
			want: Cell{
				Rune: 'X',
				Opts: &Options{
					FgColor: ColorBlack,
					BgColor: ColorBlue,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := New(tc.r, tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("New => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestCellApply(t *testing.T) {
	tests := []struct {
		desc string
		cell *Cell
		opts []Option
		want *Cell
	}{
		{
			desc: "no options provided",
			cell: New(0),
			want: New(0),
		},
		{
			desc: "no change in options",
			cell: New(0, FgColor(ColorCyan)),
			opts: []Option{
				FgColor(ColorCyan),
			},
			want: New(0, FgColor(ColorCyan)),
		},
		{
			desc: "retains previous values",
			cell: New(0, FgColor(ColorCyan)),
			opts: []Option{
				BgColor(ColorBlack),
			},
			want: New(
				0,
				FgColor(ColorCyan),
				BgColor(ColorBlack),
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

func TestNewBuffer(t *testing.T) {
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
					New(0),
				},
			},
		},
		{
			desc: "creates the buffer",
			size: image.Point{2, 3},
			want: Buffer{
				{
					New(0),
					New(0),
					New(0),
				},
				{
					New(0),
					New(0),
					New(0),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := NewBuffer(tc.size)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewBuffer => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewBuffer => unexpected diff (-want, +got):\n%s", diff)
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
			b, err := NewBuffer(size)
			if err != nil {
				t.Fatalf("NewBuffer => unexpected error: %v", err)
			}

			got := b.Size()
			if diff := pretty.Compare(size, got); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

// mustNewBuffer returns a new Buffer or panics.
func mustNewBuffer(size image.Point) Buffer {
	b, err := NewBuffer(size)
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
		opts      []Option
		wantCells int
		want      Buffer
		wantErr   bool
	}{
		{
			desc:    "point falls before the buffer",
			buffer:  mustNewBuffer(size),
			point:   image.Point{-1, -1},
			r:       'A',
			wantErr: true,
		},
		{
			desc:    "point falls after the buffer",
			buffer:  mustNewBuffer(size),
			point:   image.Point{3, 3},
			r:       'A',
			wantErr: true,
		},
		{
			desc: "point falls on cell with partial rune",
			buffer: func() Buffer {
				b := mustNewBuffer(size)
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
				b := mustNewBuffer(size)
				b[0][0].Rune = '世'
				return b
			}(),
			point:     image.Point{0, 0},
			r:         'A',
			wantCells: 1,
			want: func() Buffer {
				b := mustNewBuffer(size)
				b[0][0].Rune = 'A'
				return b
			}(),
		},
		{
			desc: "point falls on cell with full-width rune and overwrites with full-width rune",
			buffer: func() Buffer {
				b := mustNewBuffer(size)
				b[0][0].Rune = '世'
				return b
			}(),
			point:     image.Point{0, 0},
			r:         '界',
			wantCells: 2,
			want: func() Buffer {
				b := mustNewBuffer(size)
				b[0][0].Rune = '界'
				return b
			}(),
		},
		{
			desc:    "not enough space for a wide rune on the line",
			buffer:  mustNewBuffer(image.Point{3, 3}),
			point:   image.Point{2, 0},
			r:       '界',
			wantErr: true,
		},
		{
			desc:      "sets half-width rune in a cell",
			buffer:    mustNewBuffer(image.Point{3, 3}),
			point:     image.Point{1, 1},
			r:         'A',
			wantCells: 1,
			want: func() Buffer {
				b := mustNewBuffer(size)
				b[1][1].Rune = 'A'
				return b
			}(),
		},
		{
			desc:      "sets full-width rune in a cell",
			buffer:    mustNewBuffer(image.Point{3, 3}),
			point:     image.Point{1, 2},
			r:         '界',
			wantCells: 2,
			want: func() Buffer {
				b := mustNewBuffer(size)
				b[1][2].Rune = '界'
				return b
			}(),
		},
		{
			desc:   "sets cell options",
			buffer: mustNewBuffer(image.Point{3, 3}),
			point:  image.Point{1, 2},
			r:      'A',
			opts: []Option{
				FgColor(ColorRed),
				BgColor(ColorBlue),
			},
			wantCells: 1,
			want: func() Buffer {
				b := mustNewBuffer(size)
				cell := b[1][2]
				cell.Rune = 'A'
				cell.Opts = NewOptions(FgColor(ColorRed), BgColor(ColorBlue))
				return b
			}(),
		},
		{
			desc: "overwrites only provided options",
			buffer: func() Buffer {
				b := mustNewBuffer(size)
				cell := b[1][2]
				cell.Opts = NewOptions(BgColor(ColorBlue))
				return b
			}(),
			point: image.Point{1, 2},
			r:     'A',
			opts: []Option{
				FgColor(ColorRed),
			},
			wantCells: 1,
			want: func() Buffer {
				b := mustNewBuffer(size)
				cell := b[1][2]
				cell.Rune = 'A'
				cell.Opts = NewOptions(FgColor(ColorRed), BgColor(ColorBlue))
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
			buffer:  mustNewBuffer(image.Point{1, 1}),
			point:   image.Point{-1, -1},
			wantErr: true,
		},
		{
			desc:    "point falls after the buffer",
			buffer:  mustNewBuffer(image.Point{1, 1}),
			point:   image.Point{1, 1},
			wantErr: true,
		},
		{
			desc:   "the first cell cannot be partial",
			buffer: mustNewBuffer(image.Point{1, 1}),
			point:  image.Point{0, 0},
			want:   false,
		},
		{
			desc:   "previous cell on the same line contains no rune",
			buffer: mustNewBuffer(image.Point{3, 3}),
			point:  image.Point{1, 0},
			want:   false,
		},
		{
			desc: "previous cell on the same line contains half-width rune",
			buffer: func() Buffer {
				b := mustNewBuffer(image.Point{3, 3})
				b[0][0].Rune = 'A'
				return b
			}(),
			point: image.Point{1, 0},
			want:  false,
		},
		{
			desc: "previous cell on the same line contains full-width rune",
			buffer: func() Buffer {
				b := mustNewBuffer(image.Point{3, 3})
				b[0][0].Rune = '世'
				return b
			}(),
			point: image.Point{1, 0},
			want:  true,
		},
		{
			desc:   "previous cell on previous line contains no rune",
			buffer: mustNewBuffer(image.Point{3, 3}),
			point:  image.Point{0, 1},
			want:   false,
		},
		{
			desc: "previous cell on previous line contains half-width rune",
			buffer: func() Buffer {
				b := mustNewBuffer(image.Point{3, 3})
				b[2][0].Rune = 'A'
				return b
			}(),
			point: image.Point{0, 1},
			want:  false,
		},
		{
			desc: "previous cell on previous line contains full-width rune",
			buffer: func() Buffer {
				b := mustNewBuffer(image.Point{3, 3})
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
			b, err := NewBuffer(tc.size)
			if err != nil {
				t.Fatalf("NewBuffer => unexpected error: %v", err)
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
