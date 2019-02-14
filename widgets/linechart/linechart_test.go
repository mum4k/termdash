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

package linechart

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille/testbraille"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

func TestLineChartDraws(t *testing.T) {
	tests := []struct {
		desc         string
		canvas       image.Rectangle
		opts         []Option
		writes       func(*LineChart) error
		want         func(size image.Point) *faketerm.Terminal
		wantWriteErr bool
		wantErr      bool
	}{
		{
			desc:   "series fails without name for the series",
			canvas: image.Rect(0, 0, 3, 4),
			writes: func(lc *LineChart) error {
				return lc.Series("", nil)
			},
			wantWriteErr: true,
		},
		{
			desc:   "series fails when custom label has negative key",
			canvas: image.Rect(0, 0, 3, 4),
			writes: func(lc *LineChart) error {
				return lc.Series("series", nil, SeriesXLabels(map[int]string{-1: "text"}))
			},
			wantWriteErr: true,
		},
		{
			desc:   "series fails when custom label has empty value",
			canvas: image.Rect(0, 0, 3, 4),
			writes: func(lc *LineChart) error {
				return lc.Series("series", nil, SeriesXLabels(map[int]string{1: ""}))
			},
			wantWriteErr: true,
		},
		{
			desc:   "draws resize needed character when canvas is smaller than requested",
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
			desc:   "empty without series",
			canvas: image.Rect(0, 0, 3, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{1, 0}, End: image.Point{1, 2}},
					{Start: image.Point{1, 2}, End: image.Point{2, 2}},
				}
				testdraw.MustHVLines(c, lines)

				// Zero value labels.
				testdraw.MustText(c, "0", image.Point{0, 1})
				testdraw.MustText(c, "0", image.Point{2, 3})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "sets axes cell options",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				AxesCellOpts(
					cell.BgColor(cell.ColorRed),
					cell.FgColor(cell.ColorGreen),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{1, 0}, End: image.Point{1, 2}},
					{Start: image.Point{1, 2}, End: image.Point{2, 2}},
				}
				testdraw.MustHVLines(c, lines, draw.HVLineCellOpts(cell.BgColor(cell.ColorRed), cell.FgColor(cell.ColorGreen)))

				// Zero value labels.
				testdraw.MustText(c, "0", image.Point{0, 1})
				testdraw.MustText(c, "0", image.Point{2, 3})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "sets label cell options",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				XLabelCellOpts(
					cell.BgColor(cell.ColorYellow),
					cell.FgColor(cell.ColorBlue),
				),
				YLabelCellOpts(
					cell.BgColor(cell.ColorRed),
					cell.FgColor(cell.ColorGreen),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{1, 0}, End: image.Point{1, 2}},
					{Start: image.Point{1, 2}, End: image.Point{2, 2}},
				}
				testdraw.MustHVLines(c, lines)

				// Zero value labels.
				testdraw.MustText(c, "0", image.Point{0, 1}, draw.TextCellOpts(cell.BgColor(cell.ColorRed), cell.FgColor(cell.ColorGreen)))
				testdraw.MustText(c, "0", image.Point{2, 3}, draw.TextCellOpts(cell.BgColor(cell.ColorYellow), cell.FgColor(cell.ColorBlue)))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "two Y and X labels",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{5, 0}, End: image.Point{5, 8}},
					{Start: image.Point{5, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{4, 7})
				testdraw.MustText(c, "51.68", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{6, 9})
				testdraw.MustText(c, "1", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{26, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws anchored Y axis",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{1600, 1900})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{6, 0}, End: image.Point{6, 8}},
					{Start: image.Point{6, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{5, 7})
				testdraw.MustText(c, "980.80", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{7, 9})
				testdraw.MustText(c, "1", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(7, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 5}, image.Point{25, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "draws adaptive Y axis",
			opts: []Option{
				YAxisAdaptive(),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{1600, 1900})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{7, 0}, End: image.Point{7, 8}},
					{Start: image.Point{7, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "1600", image.Point{3, 7})
				testdraw.MustText(c, "1754.88", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{8, 9})
				testdraw.MustText(c, "1", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(8, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{23, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "custom X labels, horizontal by default",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100}, SeriesXLabels(map[int]string{
					0: "start",
					1: "end",
				}))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{5, 0}, End: image.Point{5, 8}},
					{Start: image.Point{5, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{4, 7})
				testdraw.MustText(c, "51.68", image.Point{0, 3})
				testdraw.MustText(c, "start", image.Point{6, 9})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{26, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "custom X labels, horizontal with option",
			opts: []Option{
				XLabelsHorizontal(),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100}, SeriesXLabels(map[int]string{
					0: "start",
					1: "end",
				}))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{5, 0}, End: image.Point{5, 8}},
					{Start: image.Point{5, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{4, 7})
				testdraw.MustText(c, "51.68", image.Point{0, 3})
				testdraw.MustText(c, "start", image.Point{6, 9})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{26, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "custom X labels, vertical",
			opts: []Option{
				XLabelsVertical(),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100}, SeriesXLabels(map[int]string{
					0: "start",
					1: "end",
				}))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{6, 0}, End: image.Point{6, 4}},
					{Start: image.Point{6, 4}, End: image.Point{19, 4}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{5, 3})
				testdraw.MustText(c, "80.040", image.Point{0, 0})
				testdraw.MustVerticalText(c, "start", image.Point{7, 5})
				testdraw.MustVerticalText(c, "end", image.Point{19, 5})

				// Braille line.
				graphAr := image.Rect(7, 0, 20, 4)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 15}, image.Point{25, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "sets series cell options",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100}, SeriesCellOpts(cell.BgColor(cell.ColorRed), cell.FgColor(cell.ColorGreen)))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{5, 0}, End: image.Point{5, 8}},
					{Start: image.Point{5, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{4, 7})
				testdraw.MustText(c, "51.68", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{6, 9})
				testdraw.MustText(c, "1", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{26, 0}, draw.BrailleLineCellOpts(cell.BgColor(cell.ColorRed), cell.FgColor(cell.ColorGreen)))
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "multiple Y and X labels",
			canvas: image.Rect(0, 0, 20, 11),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 50, 100})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{5, 0}, End: image.Point{5, 9}},
					{Start: image.Point{5, 9}, End: image.Point{19, 9}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{4, 8})
				testdraw.MustText(c, "45.76", image.Point{0, 4})
				testdraw.MustText(c, "91.52", image.Point{0, 0})
				testdraw.MustText(c, "0", image.Point{6, 10})
				testdraw.MustText(c, "1", image.Point{12, 10})
				testdraw.MustText(c, "2", image.Point{19, 10})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 9)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 35}, image.Point{13, 18})
				testdraw.MustBrailleLine(bc, image.Point{13, 18}, image.Point{27, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "Y labels are trimmed",
			canvas: image.Rect(0, 0, 5, 4),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{3, 0}, End: image.Point{3, 2}},
					{Start: image.Point{3, 2}, End: image.Point{4, 2}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{2, 1})
				testdraw.MustText(c, "57…", image.Point{0, 0})
				testdraw.MustText(c, "0", image.Point{4, 3})

				// Braille line.
				graphAr := image.Rect(4, 0, 5, 2)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 7}, image.Point{1, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draw multiple series",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 50, 100}); err != nil {
					return err
				}
				return lc.Series("second", []float64{100, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{5, 0}, End: image.Point{5, 8}},
					{Start: image.Point{5, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{4, 7})
				testdraw.MustText(c, "51.68", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{6, 9})
				testdraw.MustText(c, "1", image.Point{12, 9})
				testdraw.MustText(c, "2", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{27, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{13, 31})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draw multiple series with different cell options, last series wins where they cross",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 50, 100}, SeriesCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
					return err
				}
				return lc.Series("second", []float64{100, 0}, SeriesCellOpts(cell.FgColor(cell.ColorBlue)))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{5, 0}, End: image.Point{5, 8}},
					{Start: image.Point{5, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{4, 7})
				testdraw.MustText(c, "51.68", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{6, 9})
				testdraw.MustText(c, "1", image.Point{12, 9})
				testdraw.MustText(c, "2", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{27, 0}, draw.BrailleLineCellOpts(cell.FgColor(cell.ColorRed)))
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{13, 31}, draw.BrailleLineCellOpts(cell.FgColor(cell.ColorBlue)))
				testbraille.MustCopyTo(bc, c)

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

			widget := New(tc.opts...)
			if tc.writes != nil {
				err := tc.writes(widget)
				if (err != nil) != tc.wantWriteErr {
					t.Errorf("Series => unexpected error: %v, wantWriteErr: %v", err, tc.wantWriteErr)
				}
				if err != nil {
					return
				}
			}

			{
				err := widget.Draw(c)
				if (err != nil) != tc.wantErr {
					t.Fatalf("Draw => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					return
				}
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			want := faketerm.MustNew(c.Size())
			if tc.want != nil {
				want = tc.want(c.Size())
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
		// if not nil, executed before obtaining the options.
		addSeries func(*LineChart) error
		want      widgetapi.Options
	}{
		{
			desc: "reserves space for axis without series",
			want: widgetapi.Options{
				MinimumSize: image.Point{3, 3},
			},
		},
		{
			desc: "reserves space for longer Y labels",
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{0, 100})
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{5, 3},
			},
		},
		{
			desc: "reserves space for negative Y labels",
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{-100, 100})
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{6, 3},
			},
		},
		{
			desc: "reserves space for longer vertical X labels",
			opts: []Option{
				XLabelsVertical(),
			},
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{4, 4},
			},
		},
		{
			desc: "reserves space for longer custom vertical X labels",
			opts: []Option{
				XLabelsVertical(),
			},
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{0, 100}, SeriesXLabels(map[int]string{0: "text"}))
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{5, 6},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			lc := New(tc.opts...)

			if tc.addSeries != nil {
				if err := tc.addSeries(lc); err != nil {
					t.Fatalf("tc.addSeries => %v", err)
				}
			}
			got := lc.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
