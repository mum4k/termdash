// Copyright 2019 Google Inc.
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

// Binary linechartdemo displays a linechart widget.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"math"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
)

// sineInputs generates values from -1 to 1 for display on the line chart.
func sineInputs() []float64 {
	var res []float64

	for i := 0; i < 200; i++ {
		v := math.Sin(float64(i) / 100 * math.Pi)
		res = append(res, v)
	}
	return res
}

// playLineChart continuously adds values to the LineChart, once every delay.
// Exits when the context expires.
func playLineChart(ctx context.Context, lc *linechart.LineChart, delay time.Duration) {
	inputs := sineInputs()
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for i := 0; ; {
		select {
		case <-ticker.C:
			i = (i + 1) % len(inputs)
			rotated := append(inputs[i:], inputs[:i]...)
			if err := lc.Series("first", rotated,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorBlue)),
				linechart.SeriesXLabels(map[int]string{
					0: "zero",
				}),
			); err != nil {
				panic(err)
			}

			i2 := (i + 100) % len(inputs)
			rotated2 := append(inputs[i2:], inputs[:i2]...)
			if err := lc.Series("second", rotated2, linechart.SeriesCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func main() {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	const redrawInterval = 250 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	lc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
	)
	if err != nil {
		panic(err)
	}
	go playLineChart(ctx, lc, redrawInterval/3)
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(lc),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}
