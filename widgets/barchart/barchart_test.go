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
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

func TestGauge(t *testing.T) {
	tests := []struct {
		desc          string
		bc            *BarChart
		update        func(*BarChart) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		want          func(size image.Point) *faketerm.Terminal
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{
		{
			desc: "draws empty for no values",
			bc: New(
				Char('o'),
			),
			update: func(bc *BarChart) error {
				return nil
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "displays bars",
			bc: New(
				Char('o'),
			),
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
		},
		{
			desc: "displays bars with labels",
			bc: New(
				Char('o'),
				Labels([]string{
					"1",
					"2",
					"3",
				}),
			),
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
		},
		{
			desc: "trims too long labels",
			bc: New(
				Char('o'),
				Labels([]string{
					"1",
					"22",
					"3",
				}),
			),
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
		},
		{
			desc: "displays bars with labels and values",
			bc: New(
				Char('o'),
				Labels([]string{
					"1",
					"2",
					"3",
				}),
				ShowValues(),
			),
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
		},
		{
			desc: "bars take as much width as available",
			bc: New(
				Char('o'),
			),
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
		},
		{
			desc: "respects set bar width",
			bc: New(
				Char('o'),
				BarWidth(1),
			),
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
		},
		{
			desc: "options can be set on a call to Values",
			bc:   New(),
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
		},
		{
			desc: "respects set bar gap",
			bc: New(
				Char('o'),
				BarGap(2),
			),
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
		},
		{
			desc: "respects both width and gap",
			bc: New(
				Char('o'),
				BarGap(2),
				BarWidth(2),
			),
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
		},
		{
			desc: "respects bar and label colors",
			bc: New(
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
			),
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
		},
		{
			desc: "respects value colors",
			bc: New(
				Char('o'),
				ValueColors([]cell.Color{
					cell.ColorBlue,
					cell.ColorBlack,
				}),
				ShowValues(),
			),
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
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			err = tc.update(tc.bc)
			if (err != nil) != tc.wantUpdateErr {
				t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)

			}
			if err != nil {
				return
			}

			err = tc.bc.Draw(c)
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
				t.Errorf("Rectangle => %v", diff)
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
				return New(), nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for no bars, but have labels",
			create: func() (*BarChart, error) {
				return New(
					Labels([]string{"foo"}),
				), nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for one bar, default width, gap and no labels",
			create: func() (*BarChart, error) {
				bc := New()
				if err := bc.Values([]int{1}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for two bars, default width, gap and no labels",
			create: func() (*BarChart, error) {
				bc := New()
				if err := bc.Values([]int{1, 2}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{3, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for two bars, custom width, gap and no labels",
			create: func() (*BarChart, error) {
				bc := New(
					BarWidth(3),
					BarGap(2),
				)
				if err := bc.Values([]int{1, 2}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for two bars, custom width, gap and labels",
			create: func() (*BarChart, error) {
				bc := New(
					BarWidth(3),
					BarGap(2),
				)
				if err := bc.Values([]int{1, 2}, 3, Labels([]string{"foo", "bar"})); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 2},
				WantKeyboard: false,
				WantMouse:    false,
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
