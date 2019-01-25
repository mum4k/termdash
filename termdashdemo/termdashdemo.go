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

// Binary termdashdemo demonstrates the functionality of termdash and its various widgets.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// layout prepares the screen layout by creating the container and placing
// widgets.
func layout(ctx context.Context, t terminalapi.Terminal) (*container.Container, error) {
	spGreen, spRed := newSparkLines(ctx)
	textAndSpark := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(draw.LineStyleLight),
				container.BorderTitle("Termdash demo, press Q to quit"),
				container.BorderColor(cell.ColorNumber(39)),
				container.PlaceWidget(newTextTime(ctx)),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.Border(draw.LineStyleLight),
						container.BorderTitle("A rolling text"),
						container.PlaceWidget(newRollText(ctx)),
						container.PaddingLeft(1),
						container.PaddingRight(1),
					),
					container.Right(
						container.Border(draw.LineStyleLight),
						container.BorderTitle("A SparkLine group"),
						container.SplitHorizontal(
							container.Top(container.PlaceWidget(spGreen)),
							container.Bottom(container.PlaceWidget(spRed)),
						),
					),
				),
			),
			container.SplitPercent(30),
		),
	}

	gaugeAndHeartbeat := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(draw.LineStyleLight),
				container.BorderTitle("A Gauge"),
				container.BorderColor(cell.ColorNumber(39)),
				container.PlaceWidget(newGauge(ctx)),
				container.PaddingLeft(1),
				container.PaddingRight(1),
			),
			container.Bottom(
				container.Border(draw.LineStyleLight),
				container.BorderTitle("A LineChart"),
				container.PlaceWidget(newHeartbeat(ctx)),
			),
			container.SplitPercent(20),
		),
	}

	leftSide := []container.Option{
		container.SplitHorizontal(
			container.Top(textAndSpark...),
			container.Bottom(gaugeAndHeartbeat...),
			container.SplitPercent(50),
		),
	}

	don, err := newDonut(ctx)
	if err != nil {
		return nil, err
	}

	rightSide := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(draw.LineStyleLight),
				container.BorderTitle("BarChart"),
				container.PlaceWidget(newBarChart(ctx)),
				container.BorderTitleAlignRight(),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(draw.LineStyleLight),
						container.BorderTitle("A Donut"),
						container.BorderTitleAlignRight(),
						container.PlaceWidget(don),
					),
					container.Bottom(
						container.Border(draw.LineStyleLight),
						container.BorderTitle("Multiple series"),
						container.BorderTitleAlignRight(),
						container.PlaceWidget(newSines(ctx)),
					),
					container.SplitPercent(30),
				),
			),
			container.SplitPercent(30),
		),
	}

	c, err := container.New(
		t,
		container.SplitVertical(
			container.Left(leftSide...),
			container.Right(rightSide...),
			container.SplitPercent(70),
		),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func main() {
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c, err := layout(ctx, t)
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

// periodic executes the provided closure periodically every interval.
// Exits when the context expires.
func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// newTextTime creates a new Text widget that displays the current time.
func newTextTime(ctx context.Context) *text.Text {
	t := text.New()

	go periodic(ctx, 1*time.Second, func() error {
		t.Reset()
		txt := time.Now().UTC().Format(time.UnixDate)
		return t.Write(fmt.Sprintf("\n%s", txt), text.WriteCellOpts(cell.FgColor(cell.ColorMagenta)))
	})
	return t
}

// newRollText creates a new Text widget that displays rolling text.
func newRollText(ctx context.Context) *text.Text {
	t := text.New(text.RollContent())

	i := 0
	go periodic(ctx, 1*time.Second, func() error {
		if err := t.Write(fmt.Sprintf("Writing line %d.\n", i), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(142)))); err != nil {
			return err
		}
		i++
		return nil
	})
	return t
}

// newSparkLines creates two new sparklines displaying random values.
func newSparkLines(ctx context.Context) (*sparkline.SparkLine, *sparkline.SparkLine) {
	spGreen := sparkline.New(
		sparkline.Label("Green SparkLine", cell.FgColor(cell.ColorBlue)),
		sparkline.Color(cell.ColorGreen),
	)

	const max = 100
	go periodic(ctx, 250*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spGreen.Add([]int{v})
	})

	spRed := sparkline.New(
		sparkline.Label("Red SparkLine", cell.FgColor(cell.ColorBlue)),
		sparkline.Color(cell.ColorRed),
	)
	go periodic(ctx, 500*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spRed.Add([]int{v})
	})
	return spGreen, spRed

}

// newGauge creates a demo Gauge widget.
func newGauge(ctx context.Context) *gauge.Gauge {
	g := gauge.New()

	const start = 35
	progress := start

	go periodic(ctx, 2*time.Second, func() error {
		if err := g.Percent(progress); err != nil {
			return err
		}
		progress++
		if progress > 100 {
			progress = start
		}
		return nil
	})
	return g
}

// newDonut creates a demo Donut widget.
func newDonut(ctx context.Context) (*donut.Donut, error) {
	d, err := donut.New(donut.CellOpts(
		cell.FgColor(cell.ColorNumber(33))),
	)
	if err != nil {
		return nil, err
	}

	const start = 35
	progress := start

	go periodic(ctx, 500*time.Millisecond, func() error {
		if err := d.Percent(progress); err != nil {
			return err
		}
		progress++
		if progress > 100 {
			progress = start
		}
		return nil
	})
	return d, nil
}

// newHeartbeat returns a line chart that displays a heartbeat-like progression.
func newHeartbeat(ctx context.Context) *linechart.LineChart {
	var inputs []float64
	for i := 0; i < 100; i++ {
		v := math.Pow(math.Sin(float64(i)), 63) * math.Sin(float64(i)+1.5) * 8
		inputs = append(inputs, v)
	}

	lc := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	step := 0
	go periodic(ctx, redrawInterval/3, func() error {
		step = (step + 1) % len(inputs)
		return lc.Series("heartbeat", rotate(inputs, step),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
			linechart.SeriesXLabels(map[int]string{
				0: "zero",
			}),
		)
	})
	return lc
}

// newBarChart returns a BarcChart that displays random values on multiple bars.
func newBarChart(ctx context.Context) *barchart.BarChart {
	bc := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorNumber(33),
			cell.ColorNumber(39),
			cell.ColorNumber(45),
			cell.ColorNumber(51),
			cell.ColorNumber(81),
			cell.ColorNumber(87),
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
		}),
		barchart.ShowValues(),
	)

	const (
		bars = 6
		max  = 100
	)
	values := make([]int, bars)
	go periodic(ctx, 1*time.Second, func() error {
		for i := range values {
			values[i] = int(rand.Int31n(max + 1))
		}

		return bc.Values(values, max)
	})
	return bc
}

// newSines returns a line chart that displays multiple sine series.
func newSines(ctx context.Context) *linechart.LineChart {
	var inputs []float64
	for i := 0; i < 200; i++ {
		v := math.Sin(float64(i) / 100 * math.Pi)
		inputs = append(inputs, v)
	}

	lc := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	step1 := 0
	go periodic(ctx, redrawInterval/3, func() error {
		step1 = (step1 + 1) % len(inputs)
		if err := lc.Series("first", rotate(inputs, step1),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorBlue)),
		); err != nil {
			return err
		}

		step2 := (step1 + 100) % len(inputs)
		return lc.Series("second", rotate(inputs, step2), linechart.SeriesCellOpts(cell.FgColor(cell.ColorWhite)))
	})
	return lc
}

// rotate returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotate(inputs []float64, step int) []float64 {
	return append(inputs[step:], inputs[:step]...)
}
