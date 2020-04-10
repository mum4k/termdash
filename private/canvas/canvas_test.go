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

package canvas

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/canvas/buffer"
	"github.com/mum4k/termdash/internal/faketerm"
)

func TestNew(t *testing.T) {
	tests := []struct {
		desc     string
		area     image.Rectangle
		wantSize image.Point
		wantArea image.Rectangle
		wantErr  bool
	}{
		{
			desc:    "area min has negative X",
			area:    image.Rect(-1, 0, 0, 0),
			wantErr: true,
		},
		{
			desc:    "area min has negative Y",
			area:    image.Rect(0, -1, 0, 0),
			wantErr: true,
		},
		{
			desc:    "area max has negative X",
			area:    image.Rect(0, 0, -1, 0),
			wantErr: true,
		},
		{
			desc:    "area max has negative Y",
			area:    image.Rect(0, 0, 0, -1),
			wantErr: true,
		},
		{
			desc:    "zero area is invalid",
			area:    image.Rect(0, 0, 0, 0),
			wantErr: true,
		},
		{
			desc:     "smallest valid size",
			area:     image.Rect(0, 0, 1, 1),
			wantSize: image.Point{1, 1},
			wantArea: image.Rect(0, 0, 1, 1),
		},
		{
			desc:     "rectangular canvas 3 by 4",
			area:     image.Rect(0, 0, 3, 4),
			wantSize: image.Point{3, 4},
			wantArea: image.Rect(0, 0, 3, 4),
		},
		{
			desc:     "non-zero based area",
			area:     image.Rect(1, 1, 2, 3),
			wantSize: image.Point{1, 2},
			wantArea: image.Rect(0, 0, 1, 2),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := New(tc.area)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			gotSize := c.Size()
			if diff := pretty.Compare(tc.wantSize, gotSize); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}

			gotArea := c.Area()
			if diff := pretty.Compare(tc.wantArea, gotArea); diff != "" {
				t.Errorf("Area => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestCanvas(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		ops     func(*Canvas) error
		want    func(size image.Point) (*faketerm.Terminal, error)
		wantErr bool
	}{
		{
			desc:   "SetCellOpts fails on a point outside of the canvas",
			canvas: image.Rect(0, 0, 1, 1),
			ops: func(cvs *Canvas) error {
				return cvs.SetCellOpts(image.Point{1, 1})
			},
			wantErr: true,
		},
		{
			desc:   "SetCellOpts sets options on a cell with no options",
			canvas: image.Rect(0, 0, 2, 2),
			ops: func(cvs *Canvas) error {
				return cvs.SetCellOpts(image.Point{0, 1}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				c, err := cvs.Cell(image.Point{0, 1})
				if err != nil {
					return nil, err
				}
				if _, err := cvs.SetCell(image.Point{0, 1}, c.Rune, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetCellOpts preserves cell rune",
			canvas: image.Rect(0, 0, 2, 2),
			ops: func(cvs *Canvas) error {
				if _, err := cvs.SetCell(image.Point{0, 1}, 'X'); err != nil {
					return err
				}
				return cvs.SetCellOpts(image.Point{0, 1}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				if _, err := cvs.SetCell(image.Point{0, 1}, 'X', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetCellOpts overwrites options set previously",
			canvas: image.Rect(0, 0, 2, 2),
			ops: func(cvs *Canvas) error {
				if _, err := cvs.SetCell(image.Point{0, 1}, 'X', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return err
				}
				return cvs.SetCellOpts(image.Point{0, 1}, cell.FgColor(cell.ColorGreen), cell.BgColor(cell.ColorYellow))
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				if _, err := cvs.SetCell(image.Point{0, 1}, 'X', cell.FgColor(cell.ColorGreen), cell.BgColor(cell.ColorYellow)); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetCellOpts sets default options when no options provided",
			canvas: image.Rect(0, 0, 2, 2),
			ops: func(cvs *Canvas) error {
				if _, err := cvs.SetCell(image.Point{0, 1}, 'X', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return err
				}
				return cvs.SetCellOpts(image.Point{0, 1})
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				if _, err := cvs.SetCell(image.Point{0, 1}, 'X'); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetCellOpts is idempotent",
			canvas: image.Rect(0, 0, 2, 2),
			ops: func(cvs *Canvas) error {
				if _, err := cvs.SetCell(image.Point{0, 1}, 'X'); err != nil {
					return err
				}
				if err := cvs.SetCellOpts(image.Point{0, 1}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return err
				}
				return cvs.SetCellOpts(image.Point{0, 1}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				if _, err := cvs.SetCell(image.Point{0, 1}, 'X', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCellOpts fails on area too large",
			canvas: image.Rect(0, 0, 1, 1),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCellOpts(image.Rect(0, 0, 2, 2), cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			wantErr: true,
		},
		{
			desc:   "SetAreaCellOpts sets options in the full canvas",
			canvas: image.Rect(0, 0, 1, 1),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCellOpts(image.Rect(0, 0, 1, 1), cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				for _, p := range []image.Point{
					{0, 0},
				} {
					c, err := cvs.Cell(p)
					if err != nil {
						return nil, err
					}
					if _, err := cvs.SetCell(p, c.Rune, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
						return nil, err
					}
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCellOpts sets options in a sub-area",
			canvas: image.Rect(0, 0, 3, 3),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCellOpts(image.Rect(0, 0, 2, 2), cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				for _, p := range []image.Point{
					{0, 0},
					{0, 1},
					{1, 0},
					{1, 1},
				} {
					c, err := cvs.Cell(p)
					if err != nil {
						return nil, err
					}
					if _, err := cvs.SetCell(p, c.Rune, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
						return nil, err
					}
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCells sets cells in the full canvas",
			canvas: image.Rect(0, 0, 1, 1),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCells(image.Rect(0, 0, 1, 1), 'r')
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				if _, err := cvs.SetCell(image.Point{0, 0}, 'r'); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCells is idempotent",
			canvas: image.Rect(0, 0, 1, 1),
			ops: func(cvs *Canvas) error {
				if err := cvs.SetAreaCells(image.Rect(0, 0, 1, 1), 'r'); err != nil {
					return err
				}
				return cvs.SetAreaCells(image.Rect(0, 0, 1, 1), 'r')
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				if _, err := cvs.SetCell(image.Point{0, 0}, 'r'); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCells fails on area too large",
			canvas: image.Rect(0, 0, 1, 1),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCells(image.Rect(0, 0, 2, 2), 'r', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			wantErr: true,
		},
		{
			desc:   "SetAreaCells sets cell options",
			canvas: image.Rect(0, 0, 1, 1),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCells(image.Rect(0, 0, 1, 1), 'r', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				if _, err := cvs.SetCell(image.Point{0, 0}, 'r', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return nil, err
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCells sets cell in a sub-area",
			canvas: image.Rect(0, 0, 3, 3),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCells(image.Rect(0, 0, 2, 2), 'p')
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				for _, p := range []image.Point{
					{0, 0},
					{0, 1},
					{1, 0},
					{1, 1},
				} {
					if _, err := cvs.SetCell(p, 'p'); err != nil {
						return nil, err
					}
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCells sets full-width runes that fit",
			canvas: image.Rect(0, 0, 3, 3),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCells(image.Rect(0, 0, 2, 2), '世')
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				for _, p := range []image.Point{
					{0, 0},
					{0, 1},
				} {
					if _, err := cvs.SetCell(p, '世'); err != nil {
						return nil, err
					}
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
		{
			desc:   "SetAreaCells sets full-width runes that will leave a gap at the end of each row",
			canvas: image.Rect(0, 0, 3, 3),
			ops: func(cvs *Canvas) error {
				return cvs.SetAreaCells(image.Rect(0, 0, 3, 3), '世')
			},
			want: func(size image.Point) (*faketerm.Terminal, error) {
				ft := faketerm.MustNew(size)
				cvs, err := New(ft.Area())
				if err != nil {
					return nil, err
				}

				for _, p := range []image.Point{
					{0, 0},
					{0, 1},
					{0, 2},
				} {
					if _, err := cvs.SetCell(p, '世'); err != nil {
						return nil, err
					}
				}

				if err := cvs.Apply(ft); err != nil {
					return nil, err
				}
				return ft, nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			cvs, err := New(tc.canvas)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			if tc.ops != nil {
				err := tc.ops(cvs)
				if (err != nil) != tc.wantErr {
					t.Errorf("tc.ops => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					return
				}
			}

			size := cvs.Size()
			got, err := faketerm.New(size)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			if err := cvs.Apply(got); err != nil {
				t.Fatalf("cvs.Apply => %v", err)
			}

			var want *faketerm.Terminal
			if tc.want != nil {
				want, err = tc.want(size)
				if err != nil {
					t.Fatalf("tc.want => unexpected error: %v", err)
				}
			} else {
				w, err := faketerm.New(size)
				if err != nil {
					t.Fatalf("faketerm.New => unexpected error: %v", err)
				}
				want = w
			}

			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("cvs.SetCellOpts => %v", diff)
			}
		})
	}
}

func TestSetCellAndApply(t *testing.T) {
	tests := []struct {
		desc           string
		termSize       image.Point
		canvasArea     image.Rectangle
		point          image.Point
		r              rune
		opts           []cell.Option
		want           buffer.Buffer // Expected back buffer in the fake terminal.
		wantCells      int
		wantSetCellErr bool
		wantApplyErr   bool
	}{
		{
			desc:           "setting cell outside the designated area",
			termSize:       image.Point{2, 2},
			canvasArea:     image.Rect(0, 0, 1, 1),
			point:          image.Point{0, 2},
			wantSetCellErr: true,
		},
		{
			desc:       "sets a top-left corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{0, 0},
			r:          'X',
			wantCells:  1,
			want: buffer.Buffer{
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell('X'),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
			},
		},
		{
			desc:       "sets a full-width rune in the top-left corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{0, 0},
			r:          '界',
			wantCells:  2,
			want: buffer.Buffer{
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell('界'),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
			},
		},
		{
			desc:           "not enough space for a full-width rune",
			termSize:       image.Point{3, 3},
			canvasArea:     image.Rect(1, 1, 3, 3),
			point:          image.Point{1, 0},
			r:              '界',
			wantSetCellErr: true,
		},
		{
			desc:       "sets a top-right corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{1, 0},
			r:          'X',
			wantCells:  1,
			want: buffer.Buffer{
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell('X'),
					buffer.NewCell(0),
				},
			},
		},
		{
			desc:       "sets a bottom-left corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{0, 1},
			r:          'X',
			wantCells:  1,
			want: buffer.Buffer{
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell('X'),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
			},
		},
		{
			desc:       "sets a bottom-right corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{1, 1},
			r:          'Z',
			wantCells:  1,
			want: buffer.Buffer{
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell('Z'),
				},
			},
		},
		{
			desc:       "sets cell options",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{1, 1},
			r:          'A',
			opts: []cell.Option{
				cell.BgColor(cell.ColorRed),
			},
			wantCells: 1,
			want: buffer.Buffer{
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell(0),
				},
				{
					buffer.NewCell(0),
					buffer.NewCell(0),
					buffer.NewCell('A', cell.BgColor(cell.ColorRed)),
				},
			},
		},
		{
			desc:       "canvas size equals terminal size",
			termSize:   image.Point{1, 1},
			canvasArea: image.Rect(0, 0, 1, 1),
			point:      image.Point{0, 0},
			r:          'A',
			wantCells:  1,
			want: buffer.Buffer{
				{
					buffer.NewCell('A'),
				},
			},
		},
		{
			desc:         "terminal too small for the area",
			termSize:     image.Point{1, 1},
			canvasArea:   image.Rect(0, 0, 2, 2),
			point:        image.Point{0, 0},
			r:            'A',
			wantCells:    1,
			wantApplyErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := New(tc.canvasArea)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			gotCells, err := c.SetCell(tc.point, tc.r, tc.opts...)
			if (err != nil) != tc.wantSetCellErr {
				t.Errorf("SetCell => unexpected error: %v, wantSetCellErr: %v", err, tc.wantSetCellErr)
			}
			if err != nil {
				return
			}

			if gotCells != tc.wantCells {
				t.Errorf("SetCell => unexpected number of cells %d, want %d", gotCells, tc.wantCells)
			}

			ft, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			err = c.Apply(ft)
			if (err != nil) != tc.wantApplyErr {
				t.Errorf("Apply => unexpected error: %v, wantApplyErr: %v", err, tc.wantApplyErr)
			}
			if err != nil {
				return
			}

			got := ft.BackBuffer()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("faketerm.BackBuffer => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestClear(t *testing.T) {
	c, err := New(image.Rect(1, 1, 3, 3))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if _, err := c.SetCell(image.Point{0, 0}, 'X'); err != nil {
		t.Fatalf("SetCell => unexpected error: %v", err)
	}

	ft, err := faketerm.New(image.Point{3, 3})
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}
	// Set one cell outside of the canvas on the terminal.
	if err := ft.SetCell(image.Point{0, 0}, 'A'); err != nil {
		t.Fatalf("faketerm.SetCell => unexpected error: %v", err)
	}

	if err := c.Apply(ft); err != nil {
		t.Fatalf("Apply => unexpected error: %v", err)
	}

	want := buffer.Buffer{
		{
			buffer.NewCell('A'),
			buffer.NewCell(0),
			buffer.NewCell(0),
		},
		{
			buffer.NewCell(0),
			buffer.NewCell('X'),
			buffer.NewCell(0),
		},
		{
			buffer.NewCell(0),
			buffer.NewCell(0),
			buffer.NewCell(0),
		},
	}
	got := ft.BackBuffer()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("faketerm.BackBuffer before Clear => unexpected diff (-want, +got):\n%s", diff)
	}

	// Call Clear(), Apply() and verify that only the area belonging to the
	// canvas was cleared.
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear => unexpected error: %v", err)
	}
	if err := c.Apply(ft); err != nil {
		t.Fatalf("Apply => unexpected error: %v", err)
	}

	want = buffer.Buffer{
		{
			buffer.NewCell('A'),
			buffer.NewCell(0),
			buffer.NewCell(0),
		},
		{
			buffer.NewCell(0),
			buffer.NewCell(0),
			buffer.NewCell(0),
		},
		{
			buffer.NewCell(0),
			buffer.NewCell(0),
			buffer.NewCell(0),
		},
	}

	got = ft.BackBuffer()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("faketerm.BackBuffer after Clear => unexpected diff (-want, +got):\n%s", diff)
	}
}

// TestApplyFullWidthRunes verifies that when applying a full-width rune to the
// terminal, canvas doesn't touch the neighbor cell that holds the remaining
// part of the full-width rune.
func TestApplyFullWidthRunes(t *testing.T) {
	ar := image.Rect(0, 0, 3, 3)
	c, err := New(ar)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	fullP := image.Point{0, 0}
	if _, err := c.SetCell(fullP, '界'); err != nil {
		t.Fatalf("SetCell => unexpected error: %v", err)
	}

	ft, err := faketerm.New(area.Size(ar))
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}
	partP := image.Point{1, 0}
	if err := ft.SetCell(partP, 'A'); err != nil {
		t.Fatalf("faketerm.SetCell => unexpected error: %v", err)
	}

	if err := c.Apply(ft); err != nil {
		t.Fatalf("Apply => unexpected error: %v", err)
	}

	want, err := buffer.New(area.Size(ar))
	if err != nil {
		t.Fatalf("buffer.New => unexpected error: %v", err)
	}
	want[fullP.X][fullP.Y].Rune = '界'
	want[partP.X][partP.Y].Rune = 'A'

	got := ft.BackBuffer()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("faketerm.BackBuffer => unexpected diff (-want, +got):\n%s", diff)
	}
}

func TestCell(t *testing.T) {
	tests := []struct {
		desc    string
		cvs     func() (*Canvas, error)
		point   image.Point
		want    *buffer.Cell
		wantErr bool
	}{
		{
			desc: "requested point falls outside of the canvas",
			cvs: func() (*Canvas, error) {
				cvs, err := New(image.Rect(0, 0, 1, 1))
				if err != nil {
					return nil, err
				}
				return cvs, nil
			},
			point:   image.Point{1, 1},
			wantErr: true,
		},
		{
			desc: "returns the cell",
			cvs: func() (*Canvas, error) {
				cvs, err := New(image.Rect(0, 0, 2, 2))
				if err != nil {
					return nil, err
				}
				if _, err := cvs.SetCell(
					image.Point{1, 1}, 'A',
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				); err != nil {
					return nil, err
				}
				return cvs, nil
			},
			point: image.Point{1, 1},
			want: &buffer.Cell{
				Rune: 'A',
				Opts: cell.NewOptions(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			cvs, err := tc.cvs()
			if err != nil {
				t.Fatalf("tc.cvs => unexpected error: %v", err)
			}

			got, err := cvs.Cell(tc.point)
			if (err != nil) != tc.wantErr {
				t.Errorf("Cell => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Cell => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

// mustNew creates a new Canvas or panics.
func mustNew(ar image.Rectangle) *Canvas {
	c, err := New(ar)
	if err != nil {
		panic(err)
	}
	return c
}

// mustFill fills the canvas with the specified runes or panics.
func mustFill(c *Canvas, r rune) {
	ar := c.Area()
	for col := 0; col < ar.Max.X; col++ {
		for row := 0; row < ar.Max.Y; row++ {
			if _, err := c.SetCell(image.Point{col, row}, r); err != nil {
				panic(err)
			}
		}
	}
}

// mustSetCell sets cell at the specified point of the canvas or panics.
func mustSetCell(c *Canvas, p image.Point, r rune, opts ...cell.Option) {
	if _, err := c.SetCell(p, r, opts...); err != nil {
		panic(err)
	}
}

func TestCopyTo(t *testing.T) {
	tests := []struct {
		desc    string
		src     *Canvas
		dst     *Canvas
		want    *Canvas
		wantErr bool
	}{
		{
			desc: "fails when the canvas doesn't fit",
			src: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 3, 3))
				mustFill(c, 'X')
				return c
			}(),
			dst:     mustNew(image.Rect(0, 0, 2, 2)),
			want:    mustNew(image.Rect(0, 0, 3, 3)),
			wantErr: true,
		},
		{
			desc: "fails when the area lies outside of the destination canvas",
			src: func() *Canvas {
				c := mustNew(image.Rect(3, 3, 4, 4))
				mustFill(c, 'X')
				return c
			}(),
			dst:     mustNew(image.Rect(0, 0, 3, 3)),
			want:    mustNew(image.Rect(0, 0, 3, 3)),
			wantErr: true,
		},
		{
			desc: "copies zero based same size canvases",
			src: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 3, 3))
				mustFill(c, 'X')
				return c
			}(),
			dst: mustNew(image.Rect(0, 0, 3, 3)),
			want: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 3, 3))
				mustSetCell(c, image.Point{0, 0}, 'X')
				mustSetCell(c, image.Point{1, 0}, 'X')
				mustSetCell(c, image.Point{2, 0}, 'X')

				mustSetCell(c, image.Point{0, 1}, 'X')
				mustSetCell(c, image.Point{1, 1}, 'X')
				mustSetCell(c, image.Point{2, 1}, 'X')

				mustSetCell(c, image.Point{0, 2}, 'X')
				mustSetCell(c, image.Point{1, 2}, 'X')
				mustSetCell(c, image.Point{2, 2}, 'X')
				return c
			}(),
		},
		{
			desc: "copies smaller canvas with an offset",
			src: func() *Canvas {
				c := mustNew(image.Rect(1, 1, 2, 2))
				mustFill(c, 'X')
				return c
			}(),
			dst: mustNew(image.Rect(0, 0, 3, 3)),
			want: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 3, 3))
				mustSetCell(c, image.Point{1, 1}, 'X')
				return c
			}(),
		},
		{
			desc: "copies smaller canvas with an offset into a canvas with offset from terminal",
			src: func() *Canvas {
				c := mustNew(image.Rect(1, 1, 2, 2))
				mustFill(c, 'X')
				return c
			}(),
			dst: mustNew(image.Rect(3, 3, 6, 6)),
			want: func() *Canvas {
				c := mustNew(image.Rect(3, 3, 6, 6))
				mustSetCell(c, image.Point{1, 1}, 'X')
				return c
			}(),
		},
		{
			desc: "copies cell options",
			src: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 1, 1))
				mustSetCell(c, image.Point{0, 0}, 'X',
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				)
				return c
			}(),
			dst: mustNew(image.Rect(0, 0, 3, 1)),
			want: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 3, 1))
				mustSetCell(c, image.Point{0, 0}, 'X',
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				)
				return c
			}(),
		},
		{
			desc: "copies cells with full-width runes",
			src: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 3, 3))
				mustSetCell(c, image.Point{0, 0}, '界')
				mustSetCell(c, image.Point{1, 1}, '界')
				return c
			}(),
			dst: mustNew(image.Rect(0, 0, 3, 3)),
			want: func() *Canvas {
				c := mustNew(image.Rect(0, 0, 3, 3))
				mustSetCell(c, image.Point{0, 0}, '界')
				mustSetCell(c, image.Point{1, 1}, '界')
				return c
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.src.CopyTo(tc.dst)
			if (err != nil) != tc.wantErr {
				t.Errorf("CopyTo => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			ftSize := image.Point{10, 10}
			got, err := faketerm.New(ftSize)
			if err != nil {
				t.Fatalf("faketerm.New(tc.dst.Size()) => unexpected error: %v", err)
			}
			if err := tc.dst.Apply(got); err != nil {
				t.Fatalf("tc.dst.Apply => unexpected error: %v", err)
			}

			want, err := faketerm.New(ftSize)
			if err != nil {
				t.Fatalf("faketerm.New(tc.want.Size()) => unexpected error: %v", err)
			}

			if err := tc.want.Apply(want); err != nil {
				t.Fatalf("tc.want.Apply => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("CopyTo => %v", diff)
			}
		})
	}
}
