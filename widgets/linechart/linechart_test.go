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
	"math"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille/testbraille"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

func TestLineChartDraws(t *testing.T) {
	tests := []struct {
		desc         string
		canvas       image.Rectangle
		opts         []Option
		writes       func(*LineChart) error
		want         func(size image.Point) *faketerm.Terminal
		wantCapacity int
		wantErr      bool
		wantWriteErr bool
		wantDrawErr  bool
	}{
		{
			desc:   "fails with scroll step too low",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				ZoomStepPercent(0),
			},
			wantErr: true,
		},
		{
			desc:   "fails with scroll step too high",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				ZoomStepPercent(101),
			},
			wantErr: true,
		},
		{
			desc:   "fails with custom scale where min is NaN",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				YAxisCustomScale(math.NaN(), 1),
			},
			wantErr: true,
		},
		{
			desc:   "fails with custom scale where max is NaN",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				YAxisCustomScale(0, math.NaN()),
			},
			wantErr: true,
		},
		{
			desc:   "fails with custom scale where min > max",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				YAxisCustomScale(1, 0),
			},
			wantErr: true,
		},
		{
			desc:   "fails with custom scale where min == max",
			canvas: image.Rect(0, 0, 3, 4),
			opts: []Option{
				YAxisCustomScale(1, 1),
			},
			wantErr: true,
		},
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
			desc:         "empty without series",
			canvas:       image.Rect(0, 0, 3, 4),
			wantCapacity: 2,
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
			desc:   "empty with just one point",
			canvas: image.Rect(0, 0, 3, 4),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{1})
			},
			wantCapacity: 2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{1, 0}, End: image.Point{1, 2}},
					{Start: image.Point{1, 2}, End: image.Point{2, 2}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "…", image.Point{0, 0})
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
			wantCapacity: 2,
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
			wantCapacity: 2,
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
			wantCapacity: 28,
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
			desc: "custom Y scale, zero based positive, values fit",
			opts: []Option{
				YAxisCustomScale(0, 200),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100})
			},
			wantCapacity: 26,
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
				testdraw.MustText(c, "103.36", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{7, 9})
				testdraw.MustText(c, "1", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(7, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{25, 16})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "custom Y scale, zero based negative, values fit",
			opts: []Option{
				YAxisCustomScale(-200, 0),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, -200})
			},
			wantCapacity: 26,
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
				testdraw.MustText(c, "-200", image.Point{2, 7})
				testdraw.MustText(c, "-96.64", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{7, 9})
				testdraw.MustText(c, "1", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(7, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{25, 31})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "custom Y scale, negative and positive, values fit",
			opts: []Option{
				YAxisCustomScale(-200, 200),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 100})
			},
			wantCapacity: 30,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{19, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "-200", image.Point{0, 7})
				testdraw.MustText(c, "6.57", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{5, 9})
				testdraw.MustText(c, "1", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 16}, image.Point{29, 8})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "custom Y scale, negative only, values fit",
			opts: []Option{
				YAxisCustomScale(-200, -100),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{-200, -100})
			},
			wantCapacity: 24,
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
				testdraw.MustText(c, "-200", image.Point{3, 7})
				testdraw.MustText(c, "-148.32", image.Point{0, 3})
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
			desc: "custom Y scale, negative and positive, values don't fit so adjusted",
			opts: []Option{
				YAxisCustomScale(-200, 200),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{-400, 400})
			},
			wantCapacity: 28,
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
				testdraw.MustText(c, "-400", image.Point{1, 7})
				testdraw.MustText(c, "12.96", image.Point{0, 3})
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
			wantCapacity: 26,
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
			wantCapacity: 24,
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
			wantCapacity: 28,
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
			wantCapacity: 28,
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
			wantCapacity: 26,
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
			wantCapacity: 28,
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
			wantCapacity: 28,
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
			wantCapacity: 2,
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
			wantCapacity: 28,
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
			wantCapacity: 28,
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
		{
			desc:   "draw multiple series, the second has smaller scale than the first",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 50, 100}); err != nil {
					return err
				}
				return lc.Series("second", []float64{10, 20})
			},
			wantCapacity: 28,
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
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{13, 16})
				testdraw.MustBrailleLine(bc, image.Point{13, 16}, image.Point{27, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 28}, image.Point{13, 25})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draw multiple series, the second has larger scale than the first",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 50, 100}); err != nil {
					return err
				}
				return lc.Series("second", []float64{-10, 200})
			},
			wantCapacity: 28,
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
				testdraw.MustText(c, "-10", image.Point{2, 7})
				testdraw.MustText(c, "98.48", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{6, 9})
				testdraw.MustText(c, "1", image.Point{12, 9})
				testdraw.MustText(c, "2", image.Point{19, 9})

				// Braille line.
				graphAr := image.Rect(6, 0, 20, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 30}, image.Point{13, 22})
				testdraw.MustBrailleLine(bc, image.Point{13, 22}, image.Point{27, 15})
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{13, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "more values than capacity, X rescales",
			canvas: image.Rect(0, 0, 11, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19})
			},
			wantCapacity: 12,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{10, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{3, 7})
				testdraw.MustText(c, "9.92", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{5, 9})
				testdraw.MustText(c, "14", image.Point{9, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 11, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{1, 29})
				testdraw.MustBrailleLine(bc, image.Point{1, 29}, image.Point{1, 28})
				testdraw.MustBrailleLine(bc, image.Point{1, 28}, image.Point{2, 26})
				testdraw.MustBrailleLine(bc, image.Point{2, 26}, image.Point{2, 25})
				testdraw.MustBrailleLine(bc, image.Point{2, 25}, image.Point{3, 23})
				testdraw.MustBrailleLine(bc, image.Point{3, 23}, image.Point{3, 21})
				testdraw.MustBrailleLine(bc, image.Point{3, 21}, image.Point{4, 20})
				testdraw.MustBrailleLine(bc, image.Point{4, 20}, image.Point{5, 18})
				testdraw.MustBrailleLine(bc, image.Point{5, 18}, image.Point{5, 16})
				testdraw.MustBrailleLine(bc, image.Point{5, 16}, image.Point{6, 15})
				testdraw.MustBrailleLine(bc, image.Point{6, 15}, image.Point{6, 13})
				testdraw.MustBrailleLine(bc, image.Point{6, 13}, image.Point{7, 12})
				testdraw.MustBrailleLine(bc, image.Point{7, 12}, image.Point{8, 10})
				testdraw.MustBrailleLine(bc, image.Point{8, 10}, image.Point{8, 8})
				testdraw.MustBrailleLine(bc, image.Point{8, 8}, image.Point{9, 7})
				testdraw.MustBrailleLine(bc, image.Point{9, 7}, image.Point{9, 5})
				testdraw.MustBrailleLine(bc, image.Point{9, 5}, image.Point{10, 4})
				testdraw.MustBrailleLine(bc, image.Point{10, 4}, image.Point{10, 2})
				testdraw.MustBrailleLine(bc, image.Point{10, 2}, image.Point{11, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "more values than capacity, X unscaled",
			opts: []Option{
				XAxisUnscaled(),
			},
			canvas: image.Rect(0, 0, 11, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19})
			},
			wantCapacity: 12,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{10, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{3, 7})
				testdraw.MustText(c, "9.92", image.Point{0, 3})
				testdraw.MustText(c, "8", image.Point{5, 9})
				testdraw.MustText(c, "16", image.Point{9, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 11, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 18}, image.Point{1, 16})
				testdraw.MustBrailleLine(bc, image.Point{1, 16}, image.Point{2, 15})
				testdraw.MustBrailleLine(bc, image.Point{2, 15}, image.Point{3, 13})
				testdraw.MustBrailleLine(bc, image.Point{3, 13}, image.Point{4, 12})
				testdraw.MustBrailleLine(bc, image.Point{4, 12}, image.Point{5, 10})
				testdraw.MustBrailleLine(bc, image.Point{5, 10}, image.Point{6, 8})
				testdraw.MustBrailleLine(bc, image.Point{6, 8}, image.Point{7, 7})
				testdraw.MustBrailleLine(bc, image.Point{7, 7}, image.Point{8, 5})
				testdraw.MustBrailleLine(bc, image.Point{8, 5}, image.Point{9, 4})
				testdraw.MustBrailleLine(bc, image.Point{9, 4}, image.Point{10, 2})
				testdraw.MustBrailleLine(bc, image.Point{10, 2}, image.Point{11, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "more values than capacity, X unscaled, hides shorter series",
			opts: []Option{
				XAxisUnscaled(),
			},
			canvas: image.Rect(0, 0, 11, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}); err != nil {
					return err
				}
				return lc.Series("shorter", []float64{8, 7, 6, 5, 4, 3, 2, 1, 0})
			},
			wantCapacity: 12,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{10, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{3, 7})
				testdraw.MustText(c, "9.92", image.Point{0, 3})
				testdraw.MustText(c, "8", image.Point{5, 9})
				testdraw.MustText(c, "16", image.Point{9, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 11, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 18}, image.Point{1, 16})
				testdraw.MustBrailleLine(bc, image.Point{1, 16}, image.Point{2, 15})
				testdraw.MustBrailleLine(bc, image.Point{2, 15}, image.Point{3, 13})
				testdraw.MustBrailleLine(bc, image.Point{3, 13}, image.Point{4, 12})
				testdraw.MustBrailleLine(bc, image.Point{4, 12}, image.Point{5, 10})
				testdraw.MustBrailleLine(bc, image.Point{5, 10}, image.Point{6, 8})
				testdraw.MustBrailleLine(bc, image.Point{6, 8}, image.Point{7, 7})
				testdraw.MustBrailleLine(bc, image.Point{7, 7}, image.Point{8, 5})
				testdraw.MustBrailleLine(bc, image.Point{8, 5}, image.Point{9, 4})
				testdraw.MustBrailleLine(bc, image.Point{9, 4}, image.Point{10, 2})
				testdraw.MustBrailleLine(bc, image.Point{10, 2}, image.Point{11, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "more values than capacity, X unscaled, shorter series displayed partially",
			opts: []Option{
				XAxisUnscaled(),
			},
			canvas: image.Rect(0, 0, 11, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}); err != nil {
					return err
				}
				return lc.Series("shorter", []float64{9, 8, 7, 6, 5, 4, 3, 2, 1, 0})
			},
			wantCapacity: 12,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{10, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{3, 7})
				testdraw.MustText(c, "9.92", image.Point{0, 3})
				testdraw.MustText(c, "8", image.Point{5, 9})
				testdraw.MustText(c, "16", image.Point{9, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 11, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 18}, image.Point{1, 16})
				testdraw.MustBrailleLine(bc, image.Point{1, 16}, image.Point{2, 15})
				testdraw.MustBrailleLine(bc, image.Point{2, 15}, image.Point{3, 13})
				testdraw.MustBrailleLine(bc, image.Point{3, 13}, image.Point{4, 12})
				testdraw.MustBrailleLine(bc, image.Point{4, 12}, image.Point{5, 10})
				testdraw.MustBrailleLine(bc, image.Point{5, 10}, image.Point{6, 8})
				testdraw.MustBrailleLine(bc, image.Point{6, 8}, image.Point{7, 7})
				testdraw.MustBrailleLine(bc, image.Point{7, 7}, image.Point{8, 5})
				testdraw.MustBrailleLine(bc, image.Point{8, 5}, image.Point{9, 4})
				testdraw.MustBrailleLine(bc, image.Point{9, 4}, image.Point{10, 2})
				testdraw.MustBrailleLine(bc, image.Point{10, 2}, image.Point{11, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 29}, image.Point{1, 31})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "values fit capacity, X unscaled takes no effect",
			opts: []Option{
				XAxisUnscaled(),
			},
			canvas: image.Rect(0, 0, 11, 10),
			writes: func(lc *LineChart) error {
				return lc.Series("first", []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
			},
			wantCapacity: 12,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{10, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{3, 7})
				testdraw.MustText(c, "5.76", image.Point{0, 3})
				testdraw.MustText(c, "0", image.Point{5, 9})
				testdraw.MustText(c, "8", image.Point{9, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 11, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{1, 28})
				testdraw.MustBrailleLine(bc, image.Point{1, 28}, image.Point{2, 25})
				testdraw.MustBrailleLine(bc, image.Point{2, 25}, image.Point{3, 23})
				testdraw.MustBrailleLine(bc, image.Point{3, 23}, image.Point{4, 20})
				testdraw.MustBrailleLine(bc, image.Point{4, 20}, image.Point{5, 17})
				testdraw.MustBrailleLine(bc, image.Point{5, 17}, image.Point{6, 14})
				testdraw.MustBrailleLine(bc, image.Point{6, 14}, image.Point{7, 12})
				testdraw.MustBrailleLine(bc, image.Point{7, 12}, image.Point{8, 9})
				testdraw.MustBrailleLine(bc, image.Point{8, 9}, image.Point{9, 6})
				testdraw.MustBrailleLine(bc, image.Point{9, 6}, image.Point{10, 3})
				testdraw.MustBrailleLine(bc, image.Point{10, 3}, image.Point{11, 0})
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "highlights area for zoom",
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 100}); err != nil {
					return err
				}
				// Draw once so zoom tracker is initialized.
				cvs := testcanvas.MustNew(image.Rect(0, 0, 20, 10))
				if err := lc.Draw(cvs); err != nil {
					return err
				}
				return lc.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 5},
					Button:   mouse.ButtonLeft,
				})
			},
			wantCapacity: 28,
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

				// Highlighted area for zoom.
				testbraille.MustSetAreaCellOpts(bc, image.Rect(0, 0, 1, 8), cell.BgColor(cell.ColorNumber(235)))

				testbraille.MustCopyTo(bc, c)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "highlights area for zoom to a custom color",
			opts: []Option{
				ZoomHightlightColor(cell.ColorNumber(13)),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 100}); err != nil {
					return err
				}
				// Draw once so zoom tracker is initialized.
				cvs := testcanvas.MustNew(image.Rect(0, 0, 20, 10))
				if err := lc.Draw(cvs); err != nil {
					return err
				}
				return lc.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 5},
					Button:   mouse.ButtonLeft,
				})
			},
			wantCapacity: 28,
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

				// Highlighted area for zoom.
				testbraille.MustSetAreaCellOpts(bc, image.Rect(0, 0, 1, 8), cell.BgColor(cell.ColorNumber(13)))

				testbraille.MustCopyTo(bc, c)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "zooms in on scroll up",
			opts: []Option{
				ZoomStepPercent(50),
			},
			canvas: image.Rect(0, 0, 20, 10),
			writes: func(lc *LineChart) error {
				if err := lc.Series("first", []float64{0, 25, 75, 100}); err != nil {
					return err
				}
				// Draw once so zoom tracker is initialized.
				cvs := testcanvas.MustNew(image.Rect(0, 0, 20, 10))
				if err := lc.Draw(cvs); err != nil {
					return err
				}
				return lc.Mouse(&terminalapi.Mouse{
					Position: image.Point{8, 5},
					Button:   mouse.ButtonWheelUp,
				})
			},
			wantCapacity: 28,
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
				testdraw.MustBrailleLine(bc, image.Point{0, 31}, image.Point{13, 23})
				testdraw.MustBrailleLine(bc, image.Point{13, 23}, image.Point{27, 8})

				testbraille.MustCopyTo(bc, c)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "zoom in on unscaled X axis",
			opts: []Option{
				XAxisUnscaled(),
				ZoomStepPercent(80),
			},
			canvas: image.Rect(0, 0, 10, 10),
			writes: func(lc *LineChart) error {
				var values []float64
				for v := 0; v < 8; v++ {
					values = append(values, float64(v))
				}
				if err := lc.Series("first", values); err != nil {
					return err
				}

				// Draw once so zoom tracker is initialized.
				cvs := testcanvas.MustNew(image.Rect(0, 0, 11, 10))
				if err := lc.Draw(cvs); err != nil {
					return err
				}
				return lc.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonWheelUp,
				})
			},
			wantCapacity: 10,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{9, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{3, 7})
				testdraw.MustText(c, "3.68", image.Point{0, 3})
				testdraw.MustText(c, "1", image.Point{5, 9})
				testdraw.MustText(c, "3", image.Point{9, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 10, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 27}, image.Point{4, 22})
				testdraw.MustBrailleLine(bc, image.Point{4, 22}, image.Point{9, 18})

				testbraille.MustCopyTo(bc, c)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "shifts zoom when values on unscaled X axis roll out of the base axis view",
			opts: []Option{
				XAxisUnscaled(),
				ZoomStepPercent(80),
			},
			canvas: image.Rect(0, 0, 10, 10),
			writes: func(lc *LineChart) error {
				var values []float64
				for v := 0; v < 8; v++ {
					values = append(values, float64(v))
				}
				if err := lc.Series("first", values); err != nil {
					return err
				}

				// Draw once so zoom tracker is initialized.
				cvs := testcanvas.MustNew(image.Rect(0, 0, 11, 10))
				if err := lc.Draw(cvs); err != nil {
					return err
				}
				if err := lc.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonWheelUp,
				}); err != nil {
					return err
				}

				// Add move values
				for v := 0; v < 8; v++ {
					values = append(values, float64(v))
				}
				return lc.Series("first", values)
			},
			wantCapacity: 10,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				// Y and X axis.
				lines := []draw.HVLine{
					{Start: image.Point{4, 0}, End: image.Point{4, 8}},
					{Start: image.Point{4, 8}, End: image.Point{9, 8}},
				}
				testdraw.MustHVLines(c, lines)

				// Value labels.
				testdraw.MustText(c, "0", image.Point{3, 7})
				testdraw.MustText(c, "3.68", image.Point{0, 3})
				testdraw.MustText(c, "6", image.Point{5, 9})
				testdraw.MustText(c, "7", image.Point{9, 9})

				// Braille line.
				graphAr := image.Rect(5, 0, 10, 8)
				bc := testbraille.MustNew(graphAr)
				testdraw.MustBrailleLine(bc, image.Point{0, 5}, image.Point{8, 1})

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

			widget, err := New(tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

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
				if (err != nil) != tc.wantDrawErr {
					t.Fatalf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
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

			gotCapacity := widget.ValueCapacity()
			if gotCapacity != tc.wantCapacity {
				t.Errorf("ValueCapacity => %v, want %v", gotCapacity, tc.wantCapacity)
			}
		})
	}
}

func TestKeyboard(t *testing.T) {
	lc, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := lc.Keyboard(&terminalapi.Keyboard{}); err == nil {
		t.Errorf("Keyboard => got nil err, wanted one")
	}
}

func TestMouseDoesNothingWithoutZoomTracker(t *testing.T) {
	lc, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := lc.Mouse(&terminalapi.Mouse{}); err != nil {
		t.Errorf("Mouse => unexpected error: %v", err)
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
				MinimumSize: image.Point{3, 4},
				WantMouse: widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "reserves space for longer Y labels",
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{0, 100})
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{5, 4},
				WantMouse: widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "reserves space for negative Y labels",
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{-100, 100})
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{6, 4},
				WantMouse: widgetapi.MouseScopeWidget,
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
				MinimumSize: image.Point{4, 5},
				WantMouse: widgetapi.MouseScopeWidget,
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
				MinimumSize: image.Point{5, 7},
				WantMouse: widgetapi.MouseScopeWidget,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			lc, err := New(tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

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
