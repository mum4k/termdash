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

package gauge

import (
	"fmt"
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/draw/testdraw"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// percentCall contains arguments for a call to GaugePercent().
type percentCall struct {
	p    int
	opts []Option
}

// absoluteCall contains arguments for a call to Gauge.Absolute().
type absoluteCall struct {
	done  int
	total int
	opts  []Option
}

func TestGauge(t *testing.T) {
	tests := []struct {
		desc          string
		opts          []Option
		percent       *percentCall  // if set, the test case calls Gauge.Percent().
		absolute      *absoluteCall // if set the test case calls Gauge.Absolute().
		canvas        image.Rectangle
		meta          *widgetapi.Meta
		want          func(size image.Point) *faketerm.Terminal
		wantErr       bool
		wantUpdateErr bool // whether to expect an error on a call to Gauge.Percent() or Gauge.Absolute().
		wantDrawErr   bool
	}{
		{
			desc: "fails on negative height",
			opts: []Option{
				Height(-1),
			},
			canvas: image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "fails on negative threshold",
			opts: []Option{
				Threshold(-1, linestyle.Light),
			},
			canvas: image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "gauge without progress text",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "sets gauge color",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
				Color(cell.ColorBlue),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorBlue)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge showing percentage",
			opts: []Option{
				Char('o'),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "35%", image.Point{3, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "fails when Percent is less than zero",
			opts: []Option{
				Char('o'),
			},
			percent:       &percentCall{p: -1},
			canvas:        image.Rect(0, 0, 10, 3),
			wantUpdateErr: true,
		},
		{
			desc: "fails when Percent is more than 100",
			opts: []Option{
				Char('o'),
			},
			percent:       &percentCall{p: 101},
			canvas:        image.Rect(0, 0, 10, 3),
			wantUpdateErr: true,
		},
		{
			desc: "draws resize needed character when canvas is smaller than requested",
			opts: []Option{
				Char('o'),
				Border(linestyle.Light),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustResizeNeeded(c)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "aligns the progress text top and left",
			opts: []Option{
				Char('o'),
				HorizontalTextAlign(align.HorizontalLeft),
				VerticalTextAlign(align.VerticalTop),
			},
			percent: &percentCall{p: 0},
			canvas:  image.Rect(0, 0, 10, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "0%", image.Point{0, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "aligns the progress text top and left with border",
			opts: []Option{
				Char('o'),
				HorizontalTextAlign(align.HorizontalLeft),
				VerticalTextAlign(align.VerticalTop),
				Border(linestyle.Light),
			},
			percent: &percentCall{p: 0},
			canvas:  image.Rect(0, 0, 10, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 4))
				testdraw.MustText(c, "0%", image.Point{1, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "aligns the progress text bottom and right",
			opts: []Option{
				Char('o'),
				HorizontalTextAlign(align.HorizontalRight),
				VerticalTextAlign(align.VerticalBottom),
			},
			percent: &percentCall{p: 0},
			canvas:  image.Rect(0, 0, 10, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "0%", image.Point{8, 3})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "aligns the progress text bottom and right with border",
			opts: []Option{
				Char('o'),
				HorizontalTextAlign(align.HorizontalRight),
				VerticalTextAlign(align.VerticalBottom),
				Border(linestyle.Light),
			},
			percent: &percentCall{p: 0},
			canvas:  image.Rect(0, 0, 10, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 4))
				testdraw.MustText(c, "0%", image.Point{7, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge showing percentage with border",
			opts: []Option{
				Char('o'),
				Border(linestyle.Light),
				BorderTitle("title"),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 3),
					draw.BorderTitle("title", draw.OverrunModeThreeDot),
				)
				testdraw.MustRectangle(c, image.Rect(1, 1, 3, 2),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "35%", image.Point{3, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "respects border options",
			opts: []Option{
				Char('o'),
				Border(linestyle.Light, cell.FgColor(cell.ColorBlue)),
				BorderTitle("title"),
				BorderTitleAlign(align.HorizontalRight),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 3),
					draw.BorderCellOpts(cell.FgColor(cell.ColorBlue)),
					draw.BorderTitleAlign(align.HorizontalRight),
					draw.BorderTitle("title", draw.OverrunModeThreeDot, cell.FgColor(cell.ColorBlue)),
				)
				testdraw.MustRectangle(c, image.Rect(1, 1, 3, 2),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "35%", image.Point{3, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge showing zero percentage",
			opts: []Option{
				Char('o'),
			},
			percent: &percentCall{},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "0%", image.Point{4, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge showing 100 percent",
			opts: []Option{
				Char('o'),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 10, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "100%", image.Point{3, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge showing 100 percent with border",
			opts: []Option{
				Char('o'),
				Border(linestyle.Light),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 3))
				testdraw.MustRectangle(c, image.Rect(1, 1, 9, 2),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "100%", image.Point{3, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge showing absolute progress",
			opts: []Option{
				Char('o'),
			},
			absolute: &absoluteCall{done: 20, total: 100},
			canvas:   image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 2, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "20/100", image.Point{2, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "fails when Absolute done is negative",
			opts: []Option{
				Char('o'),
			},
			absolute:      &absoluteCall{done: -1, total: 100},
			canvas:        image.Rect(0, 0, 10, 3),
			wantUpdateErr: true,
		},
		{
			desc: "fails when Absolute total is zero",
			opts: []Option{
				Char('o'),
			},
			absolute:      &absoluteCall{done: 0, total: 0},
			canvas:        image.Rect(0, 0, 10, 3),
			wantUpdateErr: true,
		},
		{
			desc: "fails when Absolute total is less than done",
			opts: []Option{
				Char('o'),
			},
			absolute:      &absoluteCall{done: 10, total: 5},
			canvas:        image.Rect(0, 0, 10, 3),
			wantUpdateErr: true,
		},
		{
			desc: "gauge without text progress",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "passing option to Percent() overrides one provided to New()",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
			},
			percent: &percentCall{p: 35, opts: []Option{ShowTextProgress()}},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "35%", image.Point{3, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "passing option to Absolute() overrides one provided to New()",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
			},
			absolute: &absoluteCall{done: 20, total: 100, opts: []Option{ShowTextProgress()}},
			canvas:   image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 2, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "20/100", image.Point{2, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge takes full size of the canvas",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 5, 2),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 5, 2),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge with text label, half-width runes",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
				TextLabel("label"),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 10, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "(label)", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge with text label, full-width runes",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
				TextLabel("你好"),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 10, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "(你好)", image.Point{2, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge with text label, full-width runes, gauge falls on rune boundary",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
				TextLabel("你好"),
			},
			percent: &percentCall{p: 50},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 5, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "(你", image.Point{2, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testdraw.MustText(c, "好)", image.Point{5, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorDefault)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge with text label, full-width runes, gauge extended to cover full rune",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
				TextLabel("你好"),
			},
			percent: &percentCall{p: 40},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 5, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "(你", image.Point{2, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testdraw.MustText(c, "好)", image.Point{5, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorDefault)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "gauge with progress text and text label",
			opts: []Option{
				Char('o'),
				TextLabel("l"),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 10, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "100% (l)", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "text fully outside of gauge respects EmptyTextColor",
			opts: []Option{
				Char('o'),
				TextLabel("l"),
				EmptyTextColor(cell.ColorMagenta),
				FilledTextColor(cell.ColorBlue),
			},
			percent: &percentCall{p: 10},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 1, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "10% (l)", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorMagenta)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "text fully inside of gauge respects FilledTextColor",
			opts: []Option{
				Char('o'),
				TextLabel("l"),
				EmptyTextColor(cell.ColorMagenta),
				FilledTextColor(cell.ColorBlue),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 10, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "100% (l)", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlue)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "part of the text is inside and part outside of gauge",
			opts: []Option{
				Char('o'),
				TextLabel("l"),
				EmptyTextColor(cell.ColorMagenta),
				FilledTextColor(cell.ColorBlue),
			},
			percent: &percentCall{p: 50},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 5, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "50% ", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlue)),
				)
				testdraw.MustText(c, "(l)", image.Point{5, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorMagenta)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "truncates text that is outside of gauge",
			opts: []Option{
				Char('o'),
				TextLabel("long label"),
			},
			percent: &percentCall{p: 0},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "0% (long …", image.Point{0, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorDefault)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "truncates text that is outside of gauge when drawn with border",
			opts: []Option{
				Char('o'),
				TextLabel("long label"),
				Border(linestyle.Light),
			},
			percent: &percentCall{p: 0},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 3))
				testdraw.MustText(c, "0% (lon…", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorDefault)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "truncates text that is inside of gauge",
			opts: []Option{
				Char('o'),
				TextLabel("long label"),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 10, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "100% (lon…", image.Point{0, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "truncates text that is inside of gauge when drawn with border",
			opts: []Option{
				Char('o'),
				TextLabel("long label"),
				Border(linestyle.Light),
			},
			percent: &percentCall{p: 100},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 3))
				testdraw.MustRectangle(c, image.Rect(1, 1, 9, 2),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "100% (l…", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "truncates text that is inside and outside of gauge",
			opts: []Option{
				Char('o'),
				TextLabel("long label"),
			},
			percent: &percentCall{p: 50},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 5, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "50% (", image.Point{0, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testdraw.MustText(c, "long…", image.Point{5, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorDefault)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "truncates text that is inside and outside of gauge with border",
			opts: []Option{
				Char('o'),
				TextLabel("long label"),
				Border(linestyle.Light),
			},
			percent: &percentCall{p: 50},
			canvas:  image.Rect(0, 0, 10, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(c, image.Rect(0, 0, 10, 4))
				testdraw.MustRectangle(c, image.Rect(1, 1, 5, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "50% ", image.Point{1, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorBlack)),
				)
				testdraw.MustText(c, "(lo…", image.Point{5, 1},
					draw.TextCellOpts(cell.FgColor(cell.ColorDefault)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "threshold with border percentage",
			opts: []Option{
				Char('o'),
				Threshold(20, linestyle.Double),
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustText(c, "35%", image.Point{3, 1})
				testdraw.MustHVLines(c, []draw.HVLine{{
					Start: image.Point{X: 2, Y: 0},
					End:   image.Point{X: 2, Y: 2},
				}}, draw.HVLineStyle(linestyle.Double))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "threshold without border absolute",
			opts: []Option{
				Char('o'),
				Threshold(3, linestyle.Light, cell.BgColor(cell.ColorRed)),
				Border(linestyle.None),
				HideTextProgress(),
			},
			absolute: &absoluteCall{done: 4, total: 10},
			canvas:   image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustHVLines(c, []draw.HVLine{{
					Start: image.Point{X: 3, Y: 0},
					End:   image.Point{X: 3, Y: 2},
				}}, draw.HVLineStyle(linestyle.Light),
					draw.HVLineCellOpts(cell.BgColor(cell.ColorRed)))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "threshold outside of bounds (>=max)",
			opts: []Option{
				Char('o'),
				HideTextProgress(),
				Threshold(100, linestyle.Light), // ignored
			},
			percent: &percentCall{p: 35},
			canvas:  image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "progress below threshold",
			opts: []Option{
				Char('o'),
				Threshold(5, linestyle.Light, cell.BgColor(cell.ColorRed)),
				HideTextProgress(),
			},
			absolute: &absoluteCall{done: 4, total: 10},
			canvas:   image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 4, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustHVLines(c, []draw.HVLine{{
					Start: image.Point{X: 5, Y: 0},
					End:   image.Point{X: 5, Y: 2},
				}}, draw.HVLineStyle(linestyle.Light),
					draw.HVLineCellOpts(cell.BgColor(cell.ColorRed)))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "progress exactly at threshold",
			opts: []Option{
				Char('o'),
				Threshold(5, linestyle.Light, cell.BgColor(cell.ColorRed)),
				HideTextProgress(),
			},
			absolute: &absoluteCall{done: 5, total: 10},
			canvas:   image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 5, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustHVLines(c, []draw.HVLine{{
					Start: image.Point{X: 5, Y: 0},
					End:   image.Point{X: 5, Y: 2},
				}}, draw.HVLineStyle(linestyle.Light),
					draw.HVLineCellOpts(cell.BgColor(cell.ColorRed)))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "progress after threshold",
			opts: []Option{
				Char('o'),
				Threshold(5, linestyle.Light, cell.BgColor(cell.ColorRed)),
				HideTextProgress(),
			},
			absolute: &absoluteCall{done: 6, total: 10},
			canvas:   image.Rect(0, 0, 10, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustRectangle(c, image.Rect(0, 0, 6, 3),
					draw.RectChar('o'),
					draw.RectCellOpts(cell.BgColor(cell.ColorGreen)),
				)
				testdraw.MustHVLines(c, []draw.HVLine{{
					Start: image.Point{X: 5, Y: 0},
					End:   image.Point{X: 5, Y: 2},
				}}, draw.HVLineStyle(linestyle.Light),
					draw.HVLineCellOpts(cell.BgColor(cell.ColorRed)))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			g, err := New(tc.opts...)
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

			switch {
			case tc.percent != nil:
				err := g.Percent(tc.percent.p, tc.percent.opts...)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("Percent => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}

			case tc.absolute != nil:
				err := g.Absolute(tc.absolute.done, tc.absolute.total, tc.absolute.opts...)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("Absolute => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}

			}

			err = g.Draw(c, tc.meta)
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

func TestKeyboard(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := g.Keyboard(&terminalapi.Keyboard{}, &widgetapi.EventMeta{}); err == nil {
		t.Errorf("Keyboard => got nil err, wanted one")
	}
}

func TestMouse(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := g.Mouse(&terminalapi.Mouse{}, &widgetapi.EventMeta{}); err == nil {
		t.Errorf("Mouse => got nil err, wanted one")
	}
}

func TestProgressTypeString(t *testing.T) {
	tests := []struct {
		pt   progressType
		want string
	}{
		{progressType(-1), "progressTypeUnknown"},
		{progressTypePercent, "progressTypePercent"},
		{progressTypeAbsolute, "progressTypeAbsolute"},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("case(%d)", i), func(t *testing.T) {
			got := tc.pt.String()
			if tc.want != got {
				t.Errorf("String => %q, want %q", got, tc.want)
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
			desc: "reports correct minimum and maximum size",
			want: widgetapi.Options{
				MaximumSize:  image.Point{0, 0}, // Unlimited.
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "maximum size is limited when height is specified",
			opts: []Option{
				Height(2),
			},
			want: widgetapi.Options{
				MaximumSize:  image.Point{0, 2},
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
		{
			desc: "border is accounted for in maximum and minimum size",
			opts: []Option{
				Border(linestyle.Light),
				Height(2),
			},
			want: widgetapi.Options{
				MaximumSize:  image.Point{0, 4},
				MinimumSize:  image.Point{3, 3},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			g, err := New(tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}
			got := g.Options()

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}
