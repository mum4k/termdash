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

package sparkline

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/internal/draw/testdraw"
	"github.com/mum4k/termdash/internal/terminal/faketerm"
	"github.com/mum4k/termdash/internal/widgetapi"
)

func TestSparkLine(t *testing.T) {
	tests := []struct {
		desc          string
		opts          []Option
		update        func(*SparkLine) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		want          func(size image.Point) *faketerm.Terminal
		wantErr       bool
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{
		{
			desc: "fails on negative height",
			opts: []Option{
				Height(-1),
			},
			update: func(sl *SparkLine) error {
				return nil
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "draws empty for no data points",
			update: func(sl *SparkLine) error {
				return nil
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "fails on negative data points",
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 3, -1, 2})
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantUpdateErr: true,
		},
		{
			desc: "single height sparkline",
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 1, 2, 3, 4, 5, 6, 7, 8})
			},
			canvas: image.Rect(0, 0, 9, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "▁▂▃▄▅▆▇█", image.Point{1, 0}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "sparkline can be cleared",
			update: func(sl *SparkLine) error {
				if err := sl.Add([]int{0, 1, 2, 3, 4, 5, 6, 7, 8}); err != nil {
					return err
				}
				sl.Clear()
				return nil
			},
			canvas: image.Rect(0, 0, 9, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "sets sparkline color",
			opts: []Option{
				Color(cell.ColorMagenta),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 1, 2, 3, 4, 5, 6, 7, 8})
			},
			canvas: image.Rect(0, 0, 9, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "▁▂▃▄▅▆▇█", image.Point{1, 0}, draw.TextCellOpts(
					cell.FgColor(cell.ColorMagenta),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "sets sparkline color on a call to Add",
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 1, 2, 3, 4, 5, 6, 7, 8}, Color(cell.ColorMagenta))
			},
			canvas: image.Rect(0, 0, 9, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "▁▂▃▄▅▆▇█", image.Point{1, 0}, draw.TextCellOpts(
					cell.FgColor(cell.ColorMagenta),
				))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},

		{
			desc: "draws data points from the right",
			update: func(sl *SparkLine) error {
				return sl.Add([]int{7, 8})
			},
			canvas: image.Rect(0, 0, 9, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "▇█", image.Point{7, 0}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "single height sparkline with label",
			opts: []Option{
				Label("Hello"),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 1, 2, 3, 8, 3, 2, 1, 1})
			},
			canvas: image.Rect(0, 0, 9, 2),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "Hello", image.Point{0, 0})
				testdraw.MustText(c, "▁▂▃█▃▂▁▁", image.Point{1, 1}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "too long label is trimmed",
			opts: []Option{
				Label("Hello world"),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{8})
			},
			canvas: image.Rect(0, 0, 9, 2),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "Hello wo…", image.Point{0, 0})
				testdraw.MustText(c, "█", image.Point{8, 1}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "stretches up to the height of the container",
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 100, 50, 85})
			},
			canvas: image.Rect(0, 0, 4, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "█", image.Point{1, 0}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "▃", image.Point{3, 0}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "█", image.Point{1, 1}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "█", image.Point{3, 1}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "███", image.Point{1, 2}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "███", image.Point{1, 3}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "stretches up to the height of the container with label",
			opts: []Option{
				Label("zoo"),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 90, 30, 85})
			},
			canvas: image.Rect(0, 0, 4, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "zoo", image.Point{0, 0})
				testdraw.MustText(c, "█", image.Point{1, 1}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "▇", image.Point{3, 1}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "█", image.Point{1, 2}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "█", image.Point{3, 2}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "███", image.Point{1, 3}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "respects fixed height",
			opts: []Option{
				Height(2),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 100, 50, 85})
			},
			canvas: image.Rect(0, 0, 4, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "█", image.Point{1, 2}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "▆", image.Point{3, 2}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "███", image.Point{1, 3}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "draws resize needed character when canvas is smaller than requested",
			opts: []Option{
				Height(2),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 100, 50, 85})
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustResizeNeeded(c)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "respects fixed height with label",
			opts: []Option{
				Label("zoo"),
				Height(2),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 100, 50, 0})
			},
			canvas: image.Rect(0, 0, 4, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "zoo", image.Point{0, 1}, draw.TextCellOpts(
					cell.FgColor(cell.ColorDefault),
				))
				testdraw.MustText(c, "█", image.Point{1, 2}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))
				testdraw.MustText(c, "██", image.Point{1, 3}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "sets label color",
			opts: []Option{
				Label(
					"Hello",
					cell.FgColor(cell.ColorBlue),
					cell.BgColor(cell.ColorYellow),
				),
			},
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 1})
			},
			canvas: image.Rect(0, 0, 9, 2),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "Hello", image.Point{0, 0}, draw.TextCellOpts(
					cell.FgColor(cell.ColorBlue),
					cell.BgColor(cell.ColorYellow),
				))
				testdraw.MustText(c, "█", image.Point{8, 1}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "displays only data points that fit the width",
			update: func(sl *SparkLine) error {
				return sl.Add([]int{0, 1, 2, 3, 4, 5, 6, 7, 8})
			},
			canvas: image.Rect(0, 0, 3, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "▆▇█", image.Point{0, 0}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "data points not visible don't affect the determined max data point",
			update: func(sl *SparkLine) error {
				return sl.Add([]int{10, 4, 8})
			},
			canvas: image.Rect(0, 0, 2, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "▄█", image.Point{0, 0}, draw.TextCellOpts(
					cell.FgColor(DefaultColor),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			sp, err := New(tc.opts...)
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

			err = tc.update(sp)
			if (err != nil) != tc.wantUpdateErr {
				t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)
			}
			if err != nil {
				return
			}

			err = sp.Draw(c)
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
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
		want widgetapi.Options
	}{
		{
			desc: "no label and no fixed height",
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "label and no fixed height",
			opts: []Option{
				Label("foo"),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 2},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "no label and fixed height",
			opts: []Option{
				Height(3),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 3},
				MaximumSize:  image.Point{1, 3},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "label and fixed height",
			opts: []Option{
				Label("foo"),
				Height(3),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 4},
				MaximumSize:  image.Point{1, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			sp, err := New(tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}
			got := sp.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}
