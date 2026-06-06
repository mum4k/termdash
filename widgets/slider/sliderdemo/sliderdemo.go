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

// Binary sliderdemo shows the functionality of a slider widget.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/slider"
	"github.com/mum4k/termdash/widgets/text"
)

type sliderSpec struct {
	title string
	value int
	style slider.Option
	opts  []slider.Option
}

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	status, err := text.New()
	if err != nil {
		panic(err)
	}

	writeStatus := func(name string) slider.ChangeFn {
		return func(value int) error {
			status.Reset()
			if err := status.Write("  "+name+": ", text.WriteReplace(), text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
				return err
			}
			return status.Write(fmt.Sprintf("%d%%", value), text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		}
	}
	writeHelp := func() error {
		status.Reset()
		if err := status.Write("  q/ESC: quit | Tab: change focus | arrows/Home/End: adjust focused slider", text.WriteReplace(), text.WriteCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
			return err
		}
		return status.Write(" | vertical sliders use up/down", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(152))))
	}
	if err := writeHelp(); err != nil {
		panic(err)
	}

	styleRows, err := styleWidgets(writeStatus)
	if err != nil {
		panic(err)
	}

	left := []grid.Element{}
	for i, row := range styleRows {
		opts := []container.Option{
			container.Border(linestyle.Light),
			container.BorderTitle(row.title),
			container.PaddingLeft(1),
			container.PaddingRight(1),
		}
		if i == 0 {
			opts = append(opts, container.Focused())
		}
		left = append(left, grid.RowHeightFixedWithOpts(3, opts, grid.Widget(row.widget)))
	}

	verticalLeft, err := slider.New(
		slider.Min(0),
		slider.Max(100),
		slider.Value(63),
		slider.Height(14),
		slider.Orientation(slider.OrientationVertical),
		slider.AlignHorizontal(align.HorizontalLeft),
		slider.AlignVertical(align.VerticalMiddle),
		slider.SquaresStyle(),
		slider.FillCellOpts(cell.FgColor(cell.ColorNumber(111))),
		slider.TrackCellOpts(cell.FgColor(cell.ColorNumber(240))),
		slider.KnobCellOpts(cell.FgColor(cell.ColorWhite)),
		slider.FocusedKnobCellOpts(cell.FgColor(cell.ColorCyan)),
		slider.OnChange(writeStatus("Vertical left/middle")),
	)
	if err != nil {
		panic(err)
	}
	verticalRight, err := slider.New(
		slider.Min(0),
		slider.Max(100),
		slider.Value(82),
		slider.Height(10),
		slider.Orientation(slider.OrientationVertical),
		slider.AlignHorizontal(align.HorizontalRight),
		slider.AlignVertical(align.VerticalBottom),
		slider.StarsStyle(),
		slider.FillCellOpts(cell.FgColor(cell.ColorNumber(228))),
		slider.TrackCellOpts(cell.FgColor(cell.ColorNumber(238))),
		slider.KnobCellOpts(cell.FgColor(cell.ColorWhite)),
		slider.FocusedKnobCellOpts(cell.FgColor(cell.ColorCyan)),
		slider.OnChange(writeStatus("Vertical right/bottom")),
	)
	if err != nil {
		panic(err)
	}

	builder := grid.New()
	builder.Add(
		grid.ColWidthPercWithOpts(
			72,
			[]container.Option{
				container.Border(linestyle.Round),
				container.BorderTitle("Horizontal Slider Styles"),
				container.PaddingLeft(1),
				container.PaddingRight(1),
			},
			left...,
		),
		grid.ColWidthPercWithOpts(
			28,
			[]container.Option{
				container.Border(linestyle.Round),
				container.BorderTitle("Vertical + Alignment"),
				container.PaddingLeft(1),
				container.PaddingRight(1),
			},
			grid.ColWidthPercWithOpts(
				50,
				[]container.Option{
					container.Border(linestyle.Light),
					container.BorderTitle("Left / middle"),
				},
				grid.Widget(verticalLeft),
			),
			grid.ColWidthPercWithOpts(
				50,
				[]container.Option{
					container.Border(linestyle.Light),
					container.BorderTitle("Right / bottom"),
				},
				grid.Widget(verticalRight),
			),
		),
	)
	showcase, err := builder.Build()
	if err != nil {
		panic(err)
	}

	c, err := container.New(
		t,
		container.Border(linestyle.Round),
		container.BorderTitle("Slider Demo - Styles, Orientation, Alignment"),
		container.SplitHorizontal(
			container.Top(showcase...),
			container.Bottom(
				container.PlaceWidget(status),
				container.Border(linestyle.Light),
				container.BorderTitle("Status"),
			),
			container.SplitFixedFromEnd(3),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(50*time.Millisecond)); err != nil {
		panic(err)
	}
}

type titledSlider struct {
	title  string
	widget *slider.Slider
}

func styleWidgets(writeStatus func(string) slider.ChangeFn) ([]titledSlider, error) {
	specs := []sliderSpec{
		{title: "Bar - default solid bar", value: 87, style: slider.BarStyle()},
		{title: "Segmented - dashed line", value: 70, style: slider.SegmentedStyle()},
		{title: "Segmented Blocks - block segments", value: 60, style: slider.SegmentedBlocksStyle()},
		{title: "Dots - dotted style", value: 58, style: slider.DotsStyle()},
		{title: "Segmented Dots - compact dots", value: 75, style: slider.SegmentedDotsStyle()},
		{title: "Squares - square blocks", value: 43, style: slider.SquaresStyle()},
		{title: "Segmented Squares - compact squares", value: 80, style: slider.SegmentedSquaresStyle()},
		{title: "Stars - star segments", value: 65, style: slider.StarsStyle()},
	}

	var rows []titledSlider
	for i, spec := range specs {
		opts := []slider.Option{
			slider.Min(0),
			slider.Max(100),
			slider.Value(spec.value),
			slider.Width(48),
			spec.style,
			slider.OnChange(writeStatus(spec.title)),
		}
		opts = append(opts, palette(i)...)
		opts = append(opts, spec.opts...)

		w, err := slider.New(opts...)
		if err != nil {
			return nil, err
		}
		rows = append(rows, titledSlider{
			title:  spec.title,
			widget: w,
		})
	}
	return rows, nil
}

func palette(i int) []slider.Option {
	colors := []cell.Color{
		cell.ColorNumber(221),
		cell.ColorNumber(212),
		cell.ColorNumber(151),
		cell.ColorNumber(230),
		cell.ColorNumber(152),
		cell.ColorNumber(111),
		cell.ColorNumber(117),
		cell.ColorNumber(228),
	}
	fill := colors[i%len(colors)]
	return []slider.Option{
		slider.FillCellOpts(cell.FgColor(fill)),
		slider.TrackCellOpts(cell.FgColor(cell.ColorNumber(239))),
		slider.KnobCellOpts(cell.FgColor(cell.ColorWhite)),
		slider.FocusedKnobCellOpts(cell.FgColor(cell.ColorCyan)),
	}
}
