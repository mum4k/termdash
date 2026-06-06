// Copyright 2026 Google Inc.
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

package heatmap

import (
	"image"
	"strings"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

func TestNewAndValues(t *testing.T) {
	hm, err := New(CellWidth(2))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := hm.Values(nil, nil, [][]float64{{1, 2}, {3, 4}}); err != nil {
		t.Fatalf("Values => unexpected error: %v", err)
	}

	if got, want := hm.xLabels, []string{"0", "1"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("xLabels = %v, want %v", got, want)
	}
	if got, want := hm.yLabels, []string{"0", "1"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("yLabels = %v, want %v", got, want)
	}
	if got, want := hm.minValue, 1.0; got != want {
		t.Fatalf("minValue = %v, want %v", got, want)
	}
	if got, want := hm.maxValue, 4.0; got != want {
		t.Fatalf("maxValue = %v, want %v", got, want)
	}
}

func TestValuesRejectsInvalidInput(t *testing.T) {
	hm, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	for _, tc := range []struct {
		name   string
		values [][]float64
		xs     []string
		ys     []string
	}{
		{name: "empty", values: nil},
		{name: "jagged", values: [][]float64{{1, 2}, {3}}},
		{name: "x labels", values: [][]float64{{1, 2}}, xs: []string{"x0"}},
		{name: "y labels", values: [][]float64{{1, 2}}, ys: []string{"y0", "y1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := hm.Values(tc.xs, tc.ys, tc.values); err == nil {
				t.Fatal("Values => nil error, want error")
			}
		})
	}
}

func TestDrawRendersLabelsAndCells(t *testing.T) {
	hm, err := New(CellWidth(2), XLabelCellOpts(cell.FgColor(cell.ColorGreen)), YLabelCellOpts(cell.FgColor(cell.ColorYellow)))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := hm.Values([]string{"A", "B"}, []string{"LOW", "HIGH"}, [][]float64{{1, 4}, {7, 10}}); err != nil {
		t.Fatalf("Values => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 12, Y: 4})
	cvs := testcanvas.MustNew(ft.Area())
	if err := hm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)

	rendered := ft.String()
	if !strings.Contains(rendered, "LOW") || !strings.Contains(rendered, "HIGH") {
		t.Fatalf("Draw output missing Y labels: %q", rendered)
	}
	if !strings.Contains(rendered, "A") || !strings.Contains(rendered, "B") {
		t.Fatalf("Draw output missing X labels: %q", rendered)
	}

	buffer := ft.BackBuffer()
	left := buffer[4][0].Opts.BgColor
	right := buffer[6][1].Opts.BgColor
	if left == right {
		t.Fatalf("cell colors = %v and %v, want different shades", left, right)
	}
	if got, want := hm.ValueCapacity(), 9; got != want {
		t.Fatalf("ValueCapacity = %d, want %d", got, want)
	}
}

func TestDrawResizeNeededAndClearLabels(t *testing.T) {
	hm, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := hm.Values([]string{"A", "B"}, []string{"LOW", "HIGH"}, [][]float64{{1, 2}, {3, 4}}); err != nil {
		t.Fatalf("Values => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 4, Y: 2})
	cvs := testcanvas.MustNew(ft.Area())
	if err := hm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	if !strings.Contains(ft.String(), "⇄") {
		t.Fatalf("resize output = %q, want resize marker", ft.String())
	}

	hm.ClearXLabels()
	hm.ClearYLabels()
	ft = faketerm.MustNew(image.Point{X: 8, Y: 2})
	cvs = testcanvas.MustNew(ft.Area())
	if err := hm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw with cleared labels => unexpected error: %v", err)
	}
}

func TestSquareCellsAndClearXLabelsAffectMinimumSize(t *testing.T) {
	hm, err := New(SquareCells())
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := hm.Values(
		[]string{"A", "B", "C"},
		[]string{"01", "02", "03"},
		[][]float64{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		},
	); err != nil {
		t.Fatalf("Values => unexpected error: %v", err)
	}

	if got, want := hm.Options().MinimumSize, (image.Point{X: 9, Y: 4}); got != want {
		t.Fatalf("MinimumSize with SquareCells = %v, want %v", got, want)
	}

	hm.ClearXLabels()
	if got, want := hm.Options().MinimumSize, (image.Point{X: 9, Y: 3}); got != want {
		t.Fatalf("MinimumSize after ClearXLabels = %v, want %v", got, want)
	}
}

func TestGetCellColor(t *testing.T) {
	hm := &HeatMap{minValue: 0, maxValue: 100}
	if got, want := hm.getCellColor(0), cell.ColorNumber(255); got != want {
		t.Fatalf("getCellColor(min) = %v, want %v", got, want)
	}
	if got, want := hm.getCellColor(100), cell.ColorNumber(232); got != want {
		t.Fatalf("getCellColor(max) = %v, want %v", got, want)
	}
}

func TestDrawHonorsPaletteAndAxisOpts(t *testing.T) {
	hm, err := New(
		CellWidth(2),
		AxisCellOpts(cell.FgColor(cell.ColorCyan)),
		Palette(cell.ColorBlue, cell.ColorGreen, cell.ColorYellow, cell.ColorRed),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := hm.Values([]string{"A", "B"}, []string{"LOW", "HIGH"}, [][]float64{{0, 25}, {75, 100}}); err != nil {
		t.Fatalf("Values => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 12, Y: 4})
	cvs := testcanvas.MustNew(ft.Area())
	if err := hm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)

	buffer := ft.BackBuffer()
	if got, want := buffer[4][0].Opts.FgColor, cell.ColorCyan; got != want {
		t.Fatalf("axis fg color = %v, want %v", got, want)
	}
	if got, want := buffer[5][0].Opts.BgColor, cell.ColorBlue; got != want {
		t.Fatalf("low palette cell bg = %v, want %v", got, want)
	}
	if got, want := buffer[8][1].Opts.BgColor, cell.ColorRed; got != want {
		t.Fatalf("high palette cell bg = %v, want %v", got, want)
	}
}
