// Copyright 2020 Google Inc.
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

// Binary heatmapdemo displays a heatmap widget.
// Exist when 'q' is pressed.
package main

import (
	"context"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/heatmap"
)

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	hp, err := heatmap.New(
		heatmap.CellWidth(3),
		heatmap.AxisCellOpts(cell.FgColor(cell.ColorNumber(245))),
		heatmap.XLabelCellOpts(cell.FgColor(cell.ColorNumber(153))),
		heatmap.YLabelCellOpts(cell.FgColor(cell.ColorNumber(117))),
		heatmap.Palette(
			cell.ColorNumber(236),
			cell.ColorNumber(239),
			cell.ColorNumber(24),
			cell.ColorNumber(31),
			cell.ColorNumber(38),
			cell.ColorNumber(45),
			cell.ColorNumber(81),
		),
	)
	if err != nil {
		panic(err)
	}

	if err := hp.Values(
		[]string{"00", "06", "12", "18", "24"},
		[]string{"MON", "TUE", "WED", "THU", "FRI", "SAT"},
		[][]float64{
			{12, 30, 44, 28, 18},
			{18, 52, 66, 40, 24},
			{24, 48, 72, 55, 32},
			{10, 26, 38, 22, 14},
			{16, 34, 58, 44, 20},
			{8, 18, 28, 16, 10},
		},
	); err != nil {
		panic(err)
	}

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(hp),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
