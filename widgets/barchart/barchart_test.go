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

package barchart

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"

	"termdash/cell"
	"termdash/internal/canvas"
	"termdash/internal/canvas/testcanvas"
	"termdash/internal/draw"
	"termdash/internal/draw/testdraw"
	"termdash/internal/faketerm"
	"termdash/widgetapi"
)

func TestBarChart(t *testing.T) {
	tests := []struct {
		desc          string
		opts          []Option
		update        func(*BarChart) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		meta          *widgetapi.Meta
		want          func(size image.Point) *faketerm.Terminal
		wantCapacity  int
		wantErr       bool
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{
		{
			desc: "fails on negative bar width",
			opts: []Option{
				BarWidth(-1),
			},
			update: func(bc *BarChart) error {
				return nil
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "fails on negative bar gap",
			opts: []Option{
				BarGap(-1),
			},
			update: func(bc *BarChart) error {
				return nil
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "draws empty for no values",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return nil
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantCapacity: 2,
		},
		{
			desc: "fails for zero max",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{0, 2, 5, 10}, 0)
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantUpdateErr: true,
		},
		{
			desc: "fails for negative max",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{0, 2, 5, 10}, -1)
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantUpdateErr: true,
		},
		{
			desc: "fails when negative value",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{0, -2, 5, 10}, 10)
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantUpdateErr: true,
		},
		{
			desc: "fails for value larger than max",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{0, 2, 5, 11}, 10)
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantUpdateErr: true,
		},
		{
			desc: "draws resize needed character when canvas is smaller than requested",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{0, 2, 5, 10}, 10)
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustResizeNeeded(c)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "displays bars",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{0, 2, 5, 10}, 10)
			},
			canvas: image.Rect(0, 0, 7, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(4, 5, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(6, 0, 7, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 4,
		},
		{
			desc: "displays bars with labels",
			opts: []Option{
				Char('o'),
				Labels([]string{
					"1",
					"2",
					"3",
				}),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2, 5, 10}, 10)
			},
			canvas: image.Rect(0, 0, 7, 11),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(4, 5, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(6, 0, 7, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)

				// Labels.
				testdraw.MustText(c, "1", image.Point{0, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "2", image.Point{2, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "3", image.Point{4, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 4,
		},
		{
			desc: "trims too long labels",
			opts: []Option{
				Char('o'),
				Labels([]string{
					"1",
					"22",
					"3",
				}),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2, 5, 10}, 10)
			},
			canvas: image.Rect(0, 0, 7, 11),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(4, 5, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(6, 0, 7, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)

				// Labels.
				testdraw.MustText(c, "1", image.Point{0, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "…", image.Point{2, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "3", image.Point{4, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 4,
		},
		{
			desc: "displays bars with labels and values",
			opts: []Option{
				Char('o'),
				Labels([]string{
					"1",
					"2",
					"3",
				}),
				ShowValues(),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2, 5, 10}, 10)
			},
			canvas: image.Rect(0, 0, 7, 11),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(4, 5, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(6, 0, 7, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				// Labels.
				testdraw.MustText(c, "1", image.Point{0, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "2", image.Point{2, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "3", image.Point{4, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				// Values.
				testdraw.MustText(c, "1", image.Point{0, 9}, draw.TextCellOpts(
					cell.FgColor(DefaultValueColor),
					cell.BgColor(DefaultBarColor),
				))
				testdraw.MustText(c, "2", image.Point{2, 9}, draw.TextCellOpts(
					cell.FgColor(DefaultValueColor),
					cell.BgColor(DefaultBarColor),
				))
				testdraw.MustText(c, "5", image.Point{4, 9}, draw.TextCellOpts(
					cell.FgColor(DefaultValueColor),
					cell.BgColor(DefaultBarColor),
				))
				testdraw.MustText(c, "…", image.Point{6, 9}, draw.TextCellOpts(
					cell.FgColor(DefaultValueColor),
					cell.BgColor(DefaultBarColor),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 4,
		},
		{
			desc: "bars take as much width as available",
			opts: []Option{
				Char('o'),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2}, 10)
			},
			canvas: image.Rect(0, 0, 5, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 2, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(3, 8, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "respects set bar width",
			opts: []Option{
				Char('o'),
				BarWidth(1),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2}, 10)
			},
			canvas: image.Rect(0, 0, 5, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "options can be set on a call to Values",
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2}, 10, Char('o'), BarWidth(1))
			},
			canvas: image.Rect(0, 0, 5, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "respects set bar gap",
			opts: []Option{
				Char('o'),
				BarGap(2),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2}, 10)
			},
			canvas: image.Rect(0, 0, 5, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(3, 8, 4, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc: "respects both width and gap",
			opts: []Option{
				Char('o'),
				BarGap(2),
				BarWidth(2),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{5, 3}, 10)
			},
			canvas: image.Rect(0, 0, 6, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 5, 2, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(4, 7, 6, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc: "respects bar and label colors",
			opts: []Option{
				Char('o'),
				BarColors([]cell.Color{
					cell.ColorBlue,
					cell.ColorYellow,
				}),
				LabelColors([]cell.Color{
					cell.ColorCyan,
					cell.ColorMagenta,
				}),
				Labels([]string{
					"1",
					"2",
				}),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2, 3}, 10)
			},
			canvas: image.Rect(0, 0, 5, 11),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorBlue)),
				)
				testdraw.MustText(c, "1", image.Point{0, 10}, draw.TextCellOpts(
					cell.FgColor(cell.ColorCyan),
				))

				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorYellow)),
				)
				testdraw.MustText(c, "2", image.Point{2, 10}, draw.TextCellOpts(
					cell.FgColor(cell.ColorMagenta),
				))

				testdraw.MustRectangle(c, image.Rect(4, 7, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "respects value colors",
			opts: []Option{
				Char('o'),
				ValueColors([]cell.Color{
					cell.ColorBlue,
					cell.ColorBlack,
				}),
				ShowValues(),
			},
			update: func(bc *BarChart) error {
				return bc.Values([]int{0, 2, 3}, 10)
			},
			canvas: image.Rect(0, 0, 5, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "0", image.Point{0, 9}, draw.TextCellOpts(
					cell.FgColor(cell.ColorBlue),
				))

				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustText(c, "2", image.Point{2, 9}, draw.TextCellOpts(
					cell.FgColor(cell.ColorBlack),
					cell.BgColor(DefaultBarColor),
				))

				testdraw.MustRectangle(c, image.Rect(4, 7, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustText(c, "3", image.Point{4, 9}, draw.TextCellOpts(
					cell.FgColor(DefaultValueColor),
					cell.BgColor(DefaultBarColor),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "regression for #174, protects against external data mutation",
			opts: []Option{
				Char('o'),
				Labels([]string{
					"1",
					"2",
					"3",
				}),
			},
			update: func(bc *BarChart) error {
				values := []int{1, 2, 5, 10}
				if err := bc.Values(values, 10); err != nil {
					return err
				}
				values[0] = 100
				return nil
			},
			canvas: image.Rect(0, 0, 7, 11),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 9, 1, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(2, 8, 3, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(4, 5, 5, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)
				testdraw.MustRectangle(c, image.Rect(6, 0, 7, 10),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(DefaultBarColor)),
				)

				// Labels.
				testdraw.MustText(c, "1", image.Point{0, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "2", image.Point{2, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testdraw.MustText(c, "3", image.Point{4, 10}, draw.TextCellOpts(
					cell.FgColor(DefaultLabelColor),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
			wantCapacity: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bc, err := New(tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			err = tc.update(bc)
			if (err != nil) != tc.wantUpdateErr {
				t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)

			}
			if err != nil {
				return
			}

			err = bc.Draw(c, tc.meta)
			if (err != nil) != tc.wantDrawErr {
				t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
			}
			if err != nil {
				return
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(c.Size()), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if gotCapacity := bc.ValueCapacity(); gotCapacity != tc.wantCapacity {
				t.Errorf("ValueCapacity => %d, want %d", gotCapacity, tc.wantCapacity)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc   string
		create func() (*BarChart, error)
		want   widgetapi.Options
	}{
		{
			desc: "minimum size for no bars",
			create: func() (*BarChart, error) {
				return New()
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "minimum size for no bars, but have labels",
			create: func() (*BarChart, error) {
				return New(
					Labels([]string{"foo"}),
				)
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "minimum size for one bar, default width, gap and no labels",
			create: func() (*BarChart, error) {
				bc, err := New()
				if err != nil {
					return nil, err
				}
				if err := bc.Values([]int{1}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "minimum width doesn't depend on the number of values",
			create: func() (*BarChart, error) {
				bc, err := New()
				if err != nil {
					return nil, err
				}
				if err := bc.Values([]int{1, 2}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "minimum size accounts for custom bar width",
			create: func() (*BarChart, error) {
				bc, err := New(
					BarWidth(3),
				)
				if err != nil {
					return nil, err
				}
				if err := bc.Values([]int{1, 2}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{3, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bc, err := tc.create()
			if err != nil {
				t.Fatalf("create => unexpected error: %v", err)
			}

			got := bc.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}

func TestValueCapacity(t *testing.T) {
	tests := []struct {
		desc                         string
		barWidth, gapWidth, cvsWidth float64
		want                         int
	}{
		{
			desc: "zero canvas width",
		},
		{
			desc:     "no gaps, fits exactly",
			barWidth: 1,
			cvsWidth: 10,
			want:     10,
		},
		{
			desc:     "gaps, fits exactly",
			barWidth: 1,
			gapWidth: 1,
			cvsWidth: 9,
			want:     5,
		},
		{
			desc:     "gaps, one cell left at the end",
			barWidth: 1,
			gapWidth: 1,
			cvsWidth: 10,
			want:     5,
		},
		{
			desc:     "wider bars, no gaps",
			barWidth: 3,
			cvsWidth: 90,
			want:     30,
		},
		{
			desc:     "wider bars and gaps",
			barWidth: 3,
			gapWidth: 2,
			cvsWidth: 12,
			want:     2,
		},
		{
			desc:     "wider bars and gaps",
			barWidth: 3,
			gapWidth: 2,
			cvsWidth: 13,
			want:     3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := valueCapacity(tc.barWidth, tc.gapWidth, tc.cvsWidth); got != tc.want {
				t.Errorf("valueCapacity(%v, %v, %v) => %d, want %d", tc.barWidth, tc.gapWidth, tc.cvsWidth, got, tc.want)
			}
		})
	}
}
